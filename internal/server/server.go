package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	uc "github.com/SerzhLimon/ReductionURL/internal/service"
)

type Server struct {
	cfg *config.Config
	core *chi.Mux
	uc   uc.UseCase
}

func NewServer(cfg *config.Config) *Server {
	server := &Server{
		core: chi.NewRouter(),
		uc:   uc.NewService(),
	}
	server.route()
	return server
}

func (s *Server) route() {
	s.core.Post("/", s.SetURL)
	s.core.Get("/{id}", s.GetURL)
}

func (s *Server) Run() {
	fmt.Println("server started ...")
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
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://" + s.cfg.Opts.BaseURL + "/" + hash))
}

func (s *Server) GetURL(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method must be GET", http.StatusBadRequest)
		return
	}

	hash := strings.TrimPrefix(req.URL.Path, "/")

	url, err := s.uc.GetURL(hash)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
