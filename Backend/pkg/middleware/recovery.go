package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics, logs stack traces, and returns HTTP 500 JSON.
func RecoveryMiddleware(pNext http.Handler) http.Handler {
	return http.HandlerFunc(func(pWriter http.ResponseWriter, pRequest *http.Request) {
		log.Printf("[RecoveryMiddleware][IN] %s %s", pRequest.Method, pRequest.URL.Path)
		defer log.Printf("[RecoveryMiddleware][OUT] %s %s", pRequest.Method, pRequest.URL.Path)

		defer func() {
			if lErr := recover(); lErr != nil {
				log.Printf("[RecoveryMiddleware][ERROR][PANIC] Panic recovered on %s %s: %v\nStack trace:\n%s", pRequest.Method, pRequest.URL.Path, lErr, debug.Stack())
				pWriter.Header().Set("Content-Type", "application/json")
				pWriter.WriteHeader(http.StatusInternalServerError)
				lResponse := map[string]string{
					"error": "internal server error",
				}
				_ = json.NewEncoder(pWriter).Encode(lResponse)
			}
		}()

		pNext.ServeHTTP(pWriter, pRequest)
	})
}
