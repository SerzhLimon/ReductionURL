package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SerzhLimon/ReductionURL/internal/config"
)

// Mock для use case
type MockUseCase struct {
	mock.Mock
}

func newWrapServer() *Server {
	cfg := &config.Config{
		Opts: &config.Options{
			Addr:    "http://localhost:8080",
			BaseURL: "http://localhost:8080",
		},
	}
	uc := &MockUseCase{}
	return &Server{
		cfg: cfg,
		uc:  uc,
	}
}

func (m *MockUseCase) GetURL(hash string) (string, error) {
	args := m.Called(hash)
	return args.String(0), args.Error(1)
}

func (m *MockUseCase) SetURL(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func TestServerGetURL(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string

		ucIsOn    bool
		mockHash  string
		mockURL   string
		mockError error

		expectedStatus   int
		expectedLocation string
	}{
		{
			name:   "successful redirect",
			method: http.MethodGet,
			path:   "/42b3e75f92145d25",

			ucIsOn:    true,
			mockHash:  "42b3e75f92145d25",
			mockURL:   "https://practicum.yandex.ru/",
			mockError: nil,

			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			path:   "/abc123",

			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "not found",
			method: http.MethodGet,
			path:   "/qweasdzxc",

			ucIsOn:    true,
			mockHash:  "qweasdzxc",
			mockURL:   "",
			mockError: fmt.Errorf("not found"),

			expectedStatus:   http.StatusBadRequest,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newWrapServer()

			mockUC := s.uc.(*MockUseCase)
			if tt.ucIsOn {
				mockUC.On("GetURL", tt.mockHash).Return(tt.mockURL, tt.mockError)
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			res := httptest.NewRecorder()

			s.GetURL(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, res.Header().Get("Location"))
			}
			if tt.ucIsOn {
				mockUC.AssertExpectations(t)
			}
		})
	}
}

func TestServerSetURL(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		contentType string

		ucIsOn    bool
		mockURL   string
		mockHash  string
		mockError error

		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful post",
			method:      http.MethodPost,
			url:         "https://practicum.yandex.ru/",
			contentType: "text/plain",

			ucIsOn:    true,
			mockURL:   "https://practicum.yandex.ru/",
			mockHash:  "42b3e75f92145d25",
			mockError: nil,

			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/42b3e75f92145d25",
		},
		{
			name:        "bad content type",
			method:      http.MethodPost,
			url:         "https://practicum.yandex.ru/",
			contentType: "application/json",

			ucIsOn:    false,
			mockURL:   "https://practicum.yandex.ru/",
			mockHash:  "42b3e75f92145d25",
			mockError: nil,

			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Content-Type must be text/plain\n",
		},
		{
			name:        "invalid method",
			method:      http.MethodTrace,
			url:         "https://practicum.yandex.ru/",
			contentType: "text/plain",

			ucIsOn:    false,
			mockURL:   "https://practicum.yandex.ru/",
			mockHash:  "42b3e75f92145d25",
			mockError: nil,

			expectedStatus: http.StatusBadRequest,
			expectedBody:   "method must be POST\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newWrapServer()

			mockUC := s.uc.(*MockUseCase)
			if tt.ucIsOn {
				mockUC.On("SetURL", tt.mockURL).Return(tt.mockHash, tt.mockError)
			}

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.url))
			req.Header.Set("Content-Type", tt.contentType)
			res := httptest.NewRecorder()

			s.SetURL(res, req)
			assert.Equal(t, tt.expectedStatus, res.Code)
			assert.Equal(t, tt.expectedBody, res.Body.String())

			if tt.ucIsOn {
				mockUC.AssertExpectations(t)
			}
		})
	}
}
