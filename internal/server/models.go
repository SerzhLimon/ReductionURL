package server

import (
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
	Url string `json:"some_url"`
}

type SetURLJsonResponse struct {
	Url string `json:"short_url"`
}