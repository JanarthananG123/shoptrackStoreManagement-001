package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"shoptrack-backend/pkg/handlers"
	"shoptrack-backend/pkg/storage"
)

var gStoreInstance *storage.MemoryStore

func main() {
	gStoreInstance = storage.NewMemoryStore()
	lHandler := handlers.NewProductHandler(gStoreInstance)

	lRouter := mux.NewRouter()
	lApi := lRouter.PathPrefix("/api/v1").Subrouter()

	lApi.HandleFunc("/health", lHandler.HealthCheckHandler).Methods(http.MethodGet)
	lApi.HandleFunc("/products", lHandler.GetProductsHandler).Methods(http.MethodGet)
	lApi.HandleFunc("/products/{id}", lHandler.GetProductByIDHandler).Methods(http.MethodGet)
	lApi.HandleFunc("/products", lHandler.CreateProductHandler).Methods(http.MethodPost)

	lCorsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://127.0.0.1:8081"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	lServer := &http.Server{
		Handler:      lCorsHandler.Handler(lRouter),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Server running on http://localhost:8080")
	log.Fatal(lServer.ListenAndServe())
}
