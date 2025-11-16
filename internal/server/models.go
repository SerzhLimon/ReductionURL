package server

import (
	"compress/gzip"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	uc "github.com/SerzhLimon/ReductionURL/internal/service"
)

type Server struct {
	cfg  *config.Config
	core *chi.Mux
	uc   uc.UseCase
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

type SetURLJsonRequest struct {
	URL string `json:"url"`
}

type SetURLJsonResponse struct {
	URL string `json:"result"`
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	acceptsGzip bool
	wroteHeader bool
}
