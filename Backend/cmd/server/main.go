package main

import (
	"log"
	"net/http"
	"time"

	"shoptrack-backend/pkg/handlers"
	"shoptrack-backend/pkg/logger"
	"shoptrack-backend/pkg/middleware"
	"shoptrack-backend/pkg/storage"

	"github.com/gorilla/mux"
)

var gStoreInstance *storage.MemoryStore

func registerRoutes(pRouter *mux.Router, pHandler *handlers.ProductHandler) {
	log.Println("[Router][registerRoutes][IN]")
	defer log.Println("[Router][registerRoutes][OUT]")

	pRouter.HandleFunc("/getHealth", pHandler.HealthCheckHandler).Methods(http.MethodGet)
	pRouter.HandleFunc("/getProducts", pHandler.GetProductsHandler).Methods(http.MethodGet)
	pRouter.HandleFunc("/getProducts/{id}", pHandler.GetProductByIDHandler).Methods(http.MethodGet)
	pRouter.HandleFunc("/getProducts", pHandler.CreateProductHandler).Methods(http.MethodPost)
	pRouter.HandleFunc("/orders", pHandler.GetOrdersHandler).Methods(http.MethodGet)
	pRouter.HandleFunc("/orders", pHandler.CheckoutOrderHandler).Methods(http.MethodPost)
	pRouter.HandleFunc("/panic", func(pWriter http.ResponseWriter, pRequest *http.Request) {
		log.Println("[Router][panic][IN] Simulating panic")
		panic("simulated server panic for verification")
	}).Methods(http.MethodGet)
}

func main() {
	log.Println("[Server][main][IN]")
	defer log.Println("[Server][main][OUT]")

	lLogFile, lErr := logger.InitLogger()
	if lErr != nil {
		log.Printf("[Server][main][ERROR] Failed to initialize file logger: %v", lErr)
	} else if lLogFile != nil {
		defer logger.CloseLogger()
	}

	gStoreInstance = storage.NewMemoryStore()
	lHandler := handlers.NewProductHandler(gStoreInstance)

	lRouter := mux.NewRouter()

	// Register middlewares on Gorilla Mux router
	lRouter.Use(middleware.RequestLoggerMiddleware)
	lRouter.Use(middleware.CORSMiddleware)
	lRouter.Use(middleware.RecoveryMiddleware)

	// Ensure method not allowed handler also applies CORS / preflight handling
	lRouter.MethodNotAllowedHandler = middleware.CORSMiddleware(http.HandlerFunc(func(pWriter http.ResponseWriter, pRequest *http.Request) {
		log.Printf("[Router][MethodNotAllowed] %s %s", pRequest.Method, pRequest.URL.Path)
		pWriter.WriteHeader(http.StatusMethodNotAllowed)
	}))

	// Register routes on /api/v1 (existing) and /api (services base)
	lApiV1 := lRouter.PathPrefix("/api/v1").Subrouter()
	registerRoutes(lApiV1, lHandler)

	lApi := lRouter.PathPrefix("/api").Subrouter()
	registerRoutes(lApi, lHandler)

	// Global panic route for testing
	lRouter.HandleFunc("/panic", func(pWriter http.ResponseWriter, pRequest *http.Request) {
		log.Println("[Router][panic][IN] Simulating global panic")
		panic("simulated server panic for verification")
	}).Methods(http.MethodGet)

	lServer := &http.Server{
		Handler:      middleware.RequestLoggerMiddleware(middleware.CORSMiddleware(middleware.RecoveryMiddleware(lRouter))),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Server running on http://localhost:8080")
	if lServeErr := lServer.ListenAndServe(); lServeErr != nil && lServeErr != http.ErrServerClosed {
		log.Printf("[Server][main][ERROR] Server exited with error: %v", lServeErr)
	}
}
