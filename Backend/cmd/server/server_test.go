package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"shoptrack-backend/pkg/handlers"
	"shoptrack-backend/pkg/middleware"
	"shoptrack-backend/pkg/storage"

	"github.com/gorilla/mux"
)

func setupTestRouter() http.Handler {
	lStore := storage.NewMemoryStore()
	lHandler := handlers.NewProductHandler(lStore)

	lRouter := mux.NewRouter()

	lRouter.Use(middleware.RequestLoggerMiddleware)
	lRouter.Use(middleware.CORSMiddleware)
	lRouter.Use(middleware.RecoveryMiddleware)

	lRouter.MethodNotAllowedHandler = middleware.CORSMiddleware(http.HandlerFunc(func(pW http.ResponseWriter, pR *http.Request) {
		pW.WriteHeader(http.StatusMethodNotAllowed)
	}))

	lApiV1 := lRouter.PathPrefix("/api/v1").Subrouter()
	registerRoutes(lApiV1, lHandler)

	lApi := lRouter.PathPrefix("/api").Subrouter()
	registerRoutes(lApi, lHandler)

	lRouter.HandleFunc("/panic", func(pW http.ResponseWriter, pR *http.Request) {
		panic("test panic route")
	}).Methods(http.MethodGet)

	return middleware.RequestLoggerMiddleware(middleware.CORSMiddleware(middleware.RecoveryMiddleware(lRouter)))
}

func TestOptionsPreflightOnRoutes(t *testing.T) {
	lHandler := setupTestRouter()

	lEndpoints := []string{
		"/api/v1/getProducts",
		"/api/v1/getHealth",
		"/api/getProducts",
		"/api/getHealth",
		"/api/orders",
	}

	for _, lEndpoint := range lEndpoints {
		lReq := httptest.NewRequest(http.MethodOptions, lEndpoint, nil)
		lRec := httptest.NewRecorder()

		lHandler.ServeHTTP(lRec, lReq)

		if lRec.Code != http.StatusOK {
			t.Errorf("[%s] Expected status code %d for OPTIONS, got %d", lEndpoint, http.StatusOK, lRec.Code)
		}

		if lRec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("[%s] Expected Access-Control-Allow-Origin to be '*', got '%s'", lEndpoint, lRec.Header().Get("Access-Control-Allow-Origin"))
		}
		if lRec.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
			t.Errorf("[%s] Expected Access-Control-Allow-Methods 'GET, POST, PUT, DELETE, OPTIONS', got '%s'", lEndpoint, lRec.Header().Get("Access-Control-Allow-Methods"))
		}
		if lRec.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
			t.Errorf("[%s] Expected Access-Control-Allow-Headers 'Content-Type, Authorization', got '%s'", lEndpoint, lRec.Header().Get("Access-Control-Allow-Headers"))
		}
	}
}

func TestPanicRecovery(t *testing.T) {
	lHandler := setupTestRouter()

	lReq := httptest.NewRequest(http.MethodGet, "/panic", nil)
	lRec := httptest.NewRecorder()

	lHandler.ServeHTTP(lRec, lReq)

	if lRec.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status code %d for panic route, got %d", http.StatusInternalServerError, lRec.Code)
	}

	var lBody map[string]string
	lErr := json.NewDecoder(lRec.Body).Decode(&lBody)
	if lErr != nil {
		t.Fatalf("Failed to decode response: %v", lErr)
	}

	if lBody["error"] != "internal server error" {
		t.Fatalf("Expected error 'internal server error', got '%s'", lBody["error"])
	}
}

func TestGetProductsAndCreateProduct(t *testing.T) {
	lHandler := setupTestRouter()

	// 1. GET products
	lGetReq := httptest.NewRequest(http.MethodGet, "/api/v1/getProducts", nil)
	lGetRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lGetRec, lGetReq)

	if lGetRec.Code != http.StatusOK {
		t.Fatalf("Expected GET products 200, got %d", lGetRec.Code)
	}

	// 2. POST product
	lNewProd := `{"sku":"TEST-SKU-001","name":"Test Keyboard","price":99.99,"stockQuantity":5}`
	lPostReq := httptest.NewRequest(http.MethodPost, "/api/v1/getProducts", strings.NewReader(lNewProd))
	lPostReq.Header.Set("Content-Type", "application/json")
	lPostRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lPostRec, lPostReq)

	if lPostRec.Code != http.StatusCreated {
		t.Fatalf("Expected POST product 201, got %d. Body: %s", lPostRec.Code, lPostRec.Body.String())
	}
}

func TestOrders(t *testing.T) {
	lHandler := setupTestRouter()

	// 1. GET orders
	lGetReq := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	lGetRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lGetRec, lGetReq)

	if lGetRec.Code != http.StatusOK {
		t.Fatalf("Expected GET orders 200, got %d", lGetRec.Code)
	}

	// 2. POST order
	lNewOrder := `{"customer":"Amit Kumar","itemsCount":2,"totalAmount":1500.0,"status":"Completed"}`
	lPostReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lNewOrder))
	lPostReq.Header.Set("Content-Type", "application/json")
	lPostRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lPostRec, lPostReq)

	if lPostRec.Code != http.StatusCreated {
		t.Fatalf("Expected POST order 201, got %d. Body: %s", lPostRec.Code, lPostRec.Body.String())
	}

	// 3. POST order with line items & stock deduction
	lItemOrder := `{"customer":"Neha Singh","items":[{"product_id":"prod-1","quantity":2,"unit_price":24999.0}]}`
	lItemPostReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lItemOrder))
	lItemPostReq.Header.Set("Content-Type", "application/json")
	lItemPostRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lItemPostRec, lItemPostReq)

	if lItemPostRec.Code != http.StatusCreated {
		t.Fatalf("Expected POST line-item order 201, got %d. Body: %s", lItemPostRec.Code, lItemPostRec.Body.String())
	}

	// 4. POST order with insufficient stock should return 400 and error code [HNCO002]
	lExcessOrder := `{"customer":"Too Much","items":[{"product_id":"prod-1","quantity":999,"unit_price":24999.0}]}`
	lExcessReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lExcessOrder))
	lExcessReq.Header.Set("Content-Type", "application/json")
	lExcessRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lExcessRec, lExcessReq)

	if lExcessRec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for excess order, got %d. Body: %s", lExcessRec.Code, lExcessRec.Body.String())
	}
	if !strings.Contains(lExcessRec.Body.String(), "HNCO002") {
		t.Fatalf("Expected error code HNCO002 in body, got: %s", lExcessRec.Body.String())
	}
}
