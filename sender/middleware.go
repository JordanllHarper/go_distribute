package main

import (
	"log"
	"net/http"
)

type middlewareResponseWriter struct {
	http.ResponseWriter
	buf        []byte
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *middlewareResponseWriter {
	return &middlewareResponseWriter{w, []byte{}, http.StatusOK}
}

func (mrw *middlewareResponseWriter) WriteHeader(code int) {
	mrw.statusCode = code
	mrw.ResponseWriter.WriteHeader(code)
}

func (mrw *middlewareResponseWriter) Write(b []byte) (int, error) {
	mrw.buf = b
	return mrw.ResponseWriter.Write(b)
}

// Add this explicit implementation to keep Flusher support intact
func (mrw *middlewareResponseWriter) Flush() {
	if flusher, ok := mrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func addLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			loggingWriter := NewLoggingResponseWriter(w)
			method := r.Method
			pathVal := r.URL.Path
			log.Printf("%s request ==> %s\n", method, pathVal)
			next.ServeHTTP(loggingWriter, r)
			if loggingWriter.statusCode >= 400 {
				log.Printf("%d response from %s - %s\n", loggingWriter.statusCode, pathVal, string(loggingWriter.buf))
				return
			}
			log.Printf("%d response <== %s\n", loggingWriter.statusCode, pathVal)
		})
}
