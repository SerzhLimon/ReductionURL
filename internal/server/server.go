package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	s "github.com/SerzhLimon/ReductionURL/internal/service"
)

const (
	addr = "localhost:8080"
)

type Server struct {
	core *http.ServeMux
	uc   *s.Service
}

func NewServer() *Server {
	server := &Server{
		core: http.NewServeMux(),
		uc:   s.NewService(),
	}
	server.core.HandleFunc("/", server.Handlers)
	return server
}

func (s *Server) Run() {
	fmt.Println("server started ...")
	if err := http.ListenAndServe(addr, s.core); err != nil {
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
	res.Write([]byte(hash))
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

func (s *Server) Handlers(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		s.SetURL(res, req)
	case http.MethodGet:
		s.GetURL(res, req)
	default:
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
	}
}
