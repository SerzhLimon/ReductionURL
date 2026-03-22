// Package server реализует HTTP сервер для сервиса сокращения URL.
package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/SerzhLimon/ReductionURL/internal/audit"
	"github.com/SerzhLimon/ReductionURL/internal/config"
	grpcserver "github.com/SerzhLimon/ReductionURL/internal/gRPC"
	"github.com/SerzhLimon/ReductionURL/internal/model"
	uc "github.com/SerzhLimon/ReductionURL/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server представляет HTTP сервер для сокращения URL.
type Server struct {
	cfg        *config.Config
	uc         uc.UseCase
	audit      audit.Observer
	grpcSrv    *grpc.Server // gRPC сервер
	grpcLis    net.Listener // слушатель для gRPC
	httpServer *http.Server
}

// NewServer создает и инициализирует новый экземпляр Server.
func NewServer(cfg *config.Config, db *sql.DB) (*Server, error) {
	uc, err := uc.NewService(cfg, db)
	if err != nil {
		return nil, err
	}
	audit, err := audit.New(cfg)
	if err != nil {
		logrus.Warn("audit not init")
	}

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		return nil, err
	}

	grpcSrv := grpc.NewServer()
	shortenerServer := grpcserver.NewShortenerServer(uc, cfg)
	grpcserver.RegisterShortenerServiceServer(grpcSrv, shortenerServer)
	reflection.Register(grpcSrv)

	httpServer := &http.Server{
		Addr:    cfg.Opts.Addr,
		Handler: chi.NewRouter(),
	}

	server := &Server{
		cfg:        cfg,
		uc:         uc,
		audit:      audit,
		grpcSrv:    grpcSrv,
		grpcLis:    lis,
		httpServer: httpServer,
	}

	server.route()
	return server, nil
}

func (s *Server) route() {
	core, _ := s.httpServer.Handler.(*chi.Mux)
	core.Use(handLogger)
	core.Use(compress)
	core.Use(cookies)

	core.Post("/", s.SetURL)
	core.Post("/api/shorten", s.SetURLJson)
	core.Get("/{id}", s.GetURL)
	core.Get("/ping", s.Ping)
	core.Post("/api/shorten/batch", s.SetArrayURLJson)
	core.Get("/api/user/urls", s.GetArrayURLJson)
	core.Delete("/api/user/urls", s.DeleteArrayURLJson)
	core.Get("/api/internal/stats", s.GetStats)
}

// Shutdown останавливает HTTP сервер
func (s *Server) Shutdown(ctx context.Context) error {
	// Останавливаем gRPC
	s.grpcSrv.GracefulStop()
	// Останавливаем HTTP
	return s.httpServer.Shutdown(ctx)
}

// RunAudit запускает воркер аудита для отправки событий.
func (s *Server) RunAudit(ctx context.Context) {
	if s.audit == nil {
		return
	}
	s.audit.Run(ctx)
}

// Run запускает HTTP сервер.
func (s *Server) Run() error {
	// Запускаем gRPC в отдельной горутине
	go func() {
		logrus.Infof("gRPC server started on %s", s.grpcLis.Addr().String())
		if err := s.grpcSrv.Serve(s.grpcLis); err != nil {
			logrus.WithError(err).Error("gRPC server error")
		}
	}()

	// Запускаем HTTP
	if s.cfg.Opts.HTTPS {
		if _, err := NewHTTPS(); err != nil {
			logrus.Error(err)
			return err
		}
		logrus.Infof("HTTP server started with params: host - %s, file - %s", s.cfg.Opts.Addr, s.cfg.Opts.StorageFile)
		return http.ListenAndServeTLS(s.cfg.Opts.Addr, CertPEM, PrivateKeyPEM, s.httpServer.Handler)
	}
	logrus.Infof("HTTP server started with params: host - %s, file - %s", s.cfg.Opts.Addr, s.cfg.Opts.StorageFile)
	return http.ListenAndServe(s.cfg.Opts.Addr, s.httpServer.Handler)
}

// SetURL обрабатывает POST запрос для создания короткой ссылки из plain text.
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

	userID, _ := getUserID(req)
	s.uc.SetUser(userID)

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

	s.sendEvent(audit.CreateEvent(userID, audit.Shorten, string(body)))

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(s.cfg.Opts.BaseURL + "/" + hash))
}

func (s *Server) sendEvent(event audit.Event) {
	if s.audit == nil {
		return
	}
	s.audit.Update(event)
}

// GetURL обрабатывает GET запрос для редиректа по короткой ссылке.
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

	userID, _ := getUserID(req)
	s.sendEvent(audit.CreateEvent(userID, audit.Follow, url))

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// SetURLJson обрабатывает POST запрос для создания короткой ссылки из JSON.
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

	userID, _ := getUserID(req)
	s.uc.SetUser(userID)

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

	s.sendEvent(audit.CreateEvent(userID, audit.Shorten, string(body)))

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(response)
}

// Ping обрабатывает GET запрос для проверки доступности базы данных.
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

// SetArrayURLJson обрабатывает POST запрос для пакетного создания коротких ссылок.
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

// GetArrayURLJson обрабатывает GET запрос для получения всех ссылок пользователя.
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

// DeleteArrayURLJson обрабатывает DELETE запрос для удаления ссылок пользователя.
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

func (s *Server) GetStats(res http.ResponseWriter, req *http.Request) {
	if s.cfg.Opts.Subnet == "" {
		http.Error(res, model.ErrForbiddenIP.Error(), http.StatusForbidden)
		return
	}

	ipstr := req.Header.Get("X-Real-IP")

	ip := net.ParseIP(ipstr)
	if ip == nil {
		http.Error(res, "invalid IP", http.StatusBadRequest)
		return
	}

	result, err := s.uc.GetStats(s.cfg.Opts.Subnet, ip)
	if err != nil {
		if errors.Is(err, model.ErrForbiddenIP) {
			logrus.Errorln(err)
			http.Error(res, err.Error(), http.StatusForbidden)
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
	res.WriteHeader(http.StatusCreated)
	res.Write(response)
}
