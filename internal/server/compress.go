package server

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		rw := &responseWriter{
			ResponseWriter: w,
			r:              r,
		}

		next.ServeHTTP(rw, r)
		rw.finish()
	})
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	shouldCompress := (contentType == "application/json" || contentType == "text/html") &&
		strings.Contains(w.r.Header.Get("Accept-Encoding"), "gzip")

	if shouldCompress && w.writer == nil {
		w.Header().Set("Content-Encoding", "gzip")
		w.writer = gzip.NewWriter(w.ResponseWriter)
	}

	if w.writer != nil {
		return w.writer.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) finish() {
	if w.writer != nil {
		w.writer.(*gzip.Writer).Close()
	}
}
