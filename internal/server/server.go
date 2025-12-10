package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
	uc "github.com/SerzhLimon/ReductionURL/internal/service"
)

type Server struct {
	cfg  *config.Config
	core *chi.Mux
	uc   uc.UseCase
}

func NewServer(cfg *config.Config, db *sql.DB) (*Server, error) {
	uc, err := uc.NewService(cfg, db)
	if err != nil {
		return nil, err
	}
	server := &Server{
		cfg:  cfg,
		core: chi.NewRouter(),
		uc:   uc,
	}
	server.route()
	return server, nil
}

func (s *Server) route() {
	s.core.Use(handLogger)
	s.core.Use(compress)
	s.core.Use(cookies)


	s.core.Post("/", s.SetURL)
	s.core.Post("/api/shorten", s.SetURLJson)
	s.core.Get("/{id}", s.GetURL)
	s.core.Get("/ping", s.Ping)
	s.core.Post("/api/shorten/batch", s.SetArrayURLJson)
	s.core.Get("/api/user/urls", s.GetArrayURLJson)
	s.core.Delete("/api/user/urls", s.DeleteArrayURLJson)
}

func (s *Server) Run() {
	logrus.Infof("server started with params: host - %s, file - %s", s.cfg.Opts.Addr, s.cfg.Opts.StorageFile)
	if err := http.ListenAndServe(s.cfg.Opts.Addr, s.core); err != nil {
		log.Fatalln(err)
	}
}

func (s *Server) SetURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "method must be POST", http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "text/plain" {
		http.Error(res, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "cannot read body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	hash, err := s.uc.SetURL(string(body))
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusConflict)
			res.Write([]byte(s.cfg.Opts.BaseURL + "/" + hash))
			return
		}
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(s.cfg.Opts.BaseURL + "/" + hash))
}

func (s *Server) GetURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method must be GET", http.StatusBadRequest)
		return
	}

	hash := strings.TrimPrefix(req.URL.Path, "/")

	url, err := s.uc.GetURL(hash)
	if err != nil {
		if errors.Is(err, model.ErrDeletedURL) {
			http.Error(res, err.Error(), http.StatusGone)
			return
		}
		logrus.Error(err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) SetURLJson(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "method must be POST", http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "cannot read body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var request model.SetURLJsonRequest
	if err = json.Unmarshal(body, &request); err != nil {
		logrus.Errorln(err)
		http.Error(res, "cannot unmarshal body", http.StatusBadRequest)
		return
	}

	hash, err := s.uc.SetURL(request.URL)
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			hashJSON := model.SetURLJsonResponse{
				URL: s.cfg.Opts.BaseURL + "/" + hash,
			}
			response, _ := json.Marshal(hashJSON)
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)
			res.Write(response)
			return
		}
		logrus.Errorln(err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	hashJSON := model.SetURLJsonResponse{
		URL: s.cfg.Opts.BaseURL + "/" + hash,
	}

	response, err := json.Marshal(hashJSON)
	if err != nil {
		logrus.Errorln(err)
		http.Error(res, "cannot marshal body", http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(response)
}

func (s *Server) Ping(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method must be Get", http.StatusBadRequest)
		return
	}

	status := http.StatusOK
	err := s.uc.Ping()
	if err != nil {
		status = http.StatusInternalServerError
	}
	res.WriteHeader(status)
}

func (s *Server) SetArrayURLJson(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "method must be POST", http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "cannot read body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var request []model.SetArrayURLRequest
	if err = json.Unmarshal(body, &request); err != nil {
		logrus.Errorln(err)
		http.Error(res, "cannot unmarshal body", http.StatusBadRequest)
		return
	}

	result, err := s.uc.SetArrayURL(request)
	if err != nil {
		logrus.Errorln(err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		logrus.Errorln(err)
		http.Error(res, "cannot marshal body", http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(response)
}

func (s *Server) GetArrayURLJson(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method must be GET", http.StatusBadRequest)
		return
	}

	_, err := getUserID(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNoContent)
		return
	}

	result, err := s.uc.GetArrayURL()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(res, err.Error(), http.StatusNoContent)
			return
		}
		logrus.Errorln(err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		logrus.Errorln(err)
		http.Error(res, "cannot marshal body", http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(response)
}

func (s *Server) DeleteArrayURLJson(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(res, "method must be DELETE", http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "cannot read body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var hashArray []string
	if err = json.Unmarshal(body, &hashArray); err != nil {
		logrus.Errorln(err)
		http.Error(res, "cannot unmarshal body", http.StatusBadRequest)
		return
	}

	_, err = getUserID(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNoContent)
		return
	}

	s.uc.DeleteArrayURL(hashArray)

	res.WriteHeader(http.StatusAccepted)
}