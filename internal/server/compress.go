package server

import (
	"compress/gzip"
	"io"
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

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Vary", "Accept-Encoding")

			gw := &gzipWriter{
				ResponseWriter: w,
				Writer:         gzip.NewWriter(w),
			}
			defer gw.Writer.(*gzip.Writer).Close()

			next.ServeHTTP(gw, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	io.Writer
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	if g.Header().Get("Content-Type") == "" {
		g.Header().Set("Content-Type", http.DetectContentType(b))
	}

	contentType := g.Header().Get("Content-Type")
	if contentType == "application/json" || contentType == "text/html" {
		g.Header().Set("Content-Encoding", "gzip")
		return g.Writer.Write(b)
	}

	return g.ResponseWriter.Write(b)
}
