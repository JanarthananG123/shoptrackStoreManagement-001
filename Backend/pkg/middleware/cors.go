package middleware

import (
	"log"
	"net/http"
)

// CORSMiddleware attaches CORS headers and intercepts OPTIONS preflight requests.
func CORSMiddleware(pNext http.Handler) http.Handler {
	return http.HandlerFunc(func(pWriter http.ResponseWriter, pRequest *http.Request) {
		log.Printf("[CORSMiddleware][IN] %s %s", pRequest.Method, pRequest.URL.Path)
		defer log.Printf("[CORSMiddleware][OUT] %s %s", pRequest.Method, pRequest.URL.Path)

		pWriter.Header().Set("Access-Control-Allow-Origin", "*")
		pWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		pWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if pRequest.Method == http.MethodOptions {
			log.Printf("[CORSMiddleware] Preflight OPTIONS request handled with 200 OK for %s", pRequest.URL.Path)
			pWriter.WriteHeader(http.StatusOK)
			return
		}

		pNext.ServeHTTP(pWriter, pRequest)
	})
}
