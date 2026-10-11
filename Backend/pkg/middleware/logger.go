package middleware

import (
	"log"
	"net/http"
	"time"
)

type lLoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (pWriter *lLoggingResponseWriter) WriteHeader(pCode int) {
	pWriter.statusCode = pCode
	pWriter.ResponseWriter.WriteHeader(pCode)
}

// RequestLoggerMiddleware logs the entry and exit of every HTTP request along with duration and errors.
func RequestLoggerMiddleware(pNext http.Handler) http.Handler {
	return http.HandlerFunc(func(pWriter http.ResponseWriter, pRequest *http.Request) {
		lStartTime := time.Now()
		log.Printf("[HTTP][IN] %s %s from %s", pRequest.Method, pRequest.URL.Path, pRequest.RemoteAddr)

		lWrappedWriter := &lLoggingResponseWriter{
			ResponseWriter: pWriter,
			statusCode:     http.StatusOK,
		}

		pNext.ServeHTTP(lWrappedWriter, pRequest)

		lDuration := time.Since(lStartTime)
		if lWrappedWriter.statusCode >= 400 {
			log.Printf("[HTTP][ERROR] %s %s returned %d in %v", pRequest.Method, pRequest.URL.Path, lWrappedWriter.statusCode, lDuration)
		}
		log.Printf("[HTTP][OUT] %s %s -> %d in %v", pRequest.Method, pRequest.URL.Path, lWrappedWriter.statusCode, lDuration)
	})
}
