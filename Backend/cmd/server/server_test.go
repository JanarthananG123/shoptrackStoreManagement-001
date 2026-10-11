package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"shoptrack-backend/pkg/handlers"
	"shoptrack-backend/pkg/middleware"
	"shoptrack-backend/pkg/models"
	"shoptrack-backend/pkg/storage"

	"github.com/gorilla/mux"
)

func setupTestRouter() (http.Handler, *storage.MemoryStore) {
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

	return middleware.RequestLoggerMiddleware(middleware.CORSMiddleware(middleware.RecoveryMiddleware(lRouter))), lStore
}

func TestOptionsPreflightOnRoutes(pT *testing.T) {
	lHandler, _ := setupTestRouter()

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
			pT.Errorf("[%s] Expected status code %d for OPTIONS, got %d", lEndpoint, http.StatusOK, lRec.Code)
		}

		if lRec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8081" {
			pT.Errorf("[%s] Expected Access-Control-Allow-Origin to be 'http://localhost:8081', got '%s'", lEndpoint, lRec.Header().Get("Access-Control-Allow-Origin"))
		}
		if lRec.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
			pT.Errorf("[%s] Expected Access-Control-Allow-Methods 'GET, POST, PUT, DELETE, OPTIONS', got '%s'", lEndpoint, lRec.Header().Get("Access-Control-Allow-Methods"))
		}
		if lRec.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
			pT.Errorf("[%s] Expected Access-Control-Allow-Headers 'Content-Type, Authorization', got '%s'", lEndpoint, lRec.Header().Get("Access-Control-Allow-Headers"))
		}
	}
}

func TestPanicRecovery(pT *testing.T) {
	lHandler, _ := setupTestRouter()

	lReq := httptest.NewRequest(http.MethodGet, "/panic", nil)
	lRec := httptest.NewRecorder()

	lHandler.ServeHTTP(lRec, lReq)

	if lRec.Code != http.StatusInternalServerError {
		pT.Fatalf("Expected status code %d for panic route, got %d", http.StatusInternalServerError, lRec.Code)
	}

	var lBody map[string]string
	lErr := json.NewDecoder(lRec.Body).Decode(&lBody)
	if lErr != nil {
		pT.Fatalf("Failed to decode response: %v", lErr)
	}

	if lBody["error"] != "internal server error" {
		pT.Fatalf("Expected error 'internal server error', got '%s'", lBody["error"])
	}
}

func TestGetProductsAndCreateProduct(pT *testing.T) {
	lHandler, _ := setupTestRouter()

	// 1. GET products
	lGetReq := httptest.NewRequest(http.MethodGet, "/api/v1/getProducts", nil)
	lGetRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lGetRec, lGetReq)

	if lGetRec.Code != http.StatusOK {
		pT.Fatalf("Expected GET products 200, got %d", lGetRec.Code)
	}

	// 2. POST product
	lNewProd := `{"sku":"TEST-SKU-001","name":"Test Keyboard","price":99.99,"stockQuantity":5}`
	lPostReq := httptest.NewRequest(http.MethodPost, "/api/v1/getProducts", strings.NewReader(lNewProd))
	lPostReq.Header.Set("Content-Type", "application/json")
	lPostRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lPostRec, lPostReq)

	if lPostRec.Code != http.StatusCreated {
		pT.Fatalf("Expected POST product 201, got %d. Body: %s", lPostRec.Code, lPostRec.Body.String())
	}
}

func TestOrders(pT *testing.T) {
	lHandler, lStore := setupTestRouter()

	// 1. GET orders
	lGetReq := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	lGetRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lGetRec, lGetReq)

	if lGetRec.Code != http.StatusOK {
		pT.Fatalf("Expected GET orders 200, got %d", lGetRec.Code)
	}

	// 2. Reject empty orders [HNCO001]
	lEmptyOrder := `{"customer":"Empty Cart","items":[]}`
	lEmptyReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lEmptyOrder))
	lEmptyReq.Header.Set("Content-Type", "application/json")
	lEmptyRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lEmptyRec, lEmptyReq)

	if lEmptyRec.Code != http.StatusBadRequest {
		pT.Fatalf("Expected 400 for empty items, got %d. Body: %s", lEmptyRec.Code, lEmptyRec.Body.String())
	}
	if !strings.Contains(lEmptyRec.Body.String(), "HNCO001") {
		pT.Fatalf("Expected error code HNCO001 for empty items, got: %s", lEmptyRec.Body.String())
	}

	// 3. Price tampering test: client sends unit_price: 0.01 and totalAmount: 0.02
	// Backend must ignore client prices, use store price (24999.00), and deduct stock (15 - 2 = 13).
	lTamperOrder := `{"customer":"Neha Singh","totalAmount":0.02,"items":[{"product_id":"prod-1","quantity":2,"unit_price":0.01}]}`
	lTamperReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lTamperOrder))
	lTamperReq.Header.Set("Content-Type", "application/json")
	lTamperRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lTamperRec, lTamperReq)

	if lTamperRec.Code != http.StatusCreated {
		pT.Fatalf("Expected POST order 201, got %d. Body: %s", lTamperRec.Code, lTamperRec.Body.String())
	}

	var lOrderResp struct {
		Success bool         `json:"success"`
		Data    models.Order `json:"data"`
	}
	if lErr := json.NewDecoder(lTamperRec.Body).Decode(&lOrderResp); lErr != nil {
		pT.Fatalf("Failed to decode checkout response: %v", lErr)
	}

	lExpectedTotal := 2 * 24999.00
	if lOrderResp.Data.TotalAmount != lExpectedTotal {
		pT.Errorf("Expected TotalAmount ₹%.2f, got ₹%.2f (client price was not ignored!)", lExpectedTotal, lOrderResp.Data.TotalAmount)
	}
	if len(lOrderResp.Data.Items) > 0 && lOrderResp.Data.Items[0].UnitPrice != 24999.00 {
		pT.Errorf("Expected UnitPrice ₹24999.00, got ₹%.2f", lOrderResp.Data.Items[0].UnitPrice)
	}

	// Verify stock deduction for prod-1 (15 - 2 = 13)
	lProd1, lFound1 := lStore.GetByID("prod-1")
	if !lFound1 {
		pT.Fatalf("Product prod-1 not found in store")
	}
	if lProd1.StockQuantity != 13 {
		pT.Errorf("Expected remaining stock for prod-1 to be 13, got %d", lProd1.StockQuantity)
	}

	// 4. Duplicate line items in single order: quantities should be aggregated before checking & deducting
	// prod-2 has initial stock 8. Request 2 + 3 = 5 items. Remaining stock should be 8 - 5 = 3.
	lDuplicateOrder := `{"customer":"Ravi Sharma","items":[{"product_id":"prod-2","quantity":2},{"product_id":"prod-2","quantity":3}]}`
	lDupReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lDuplicateOrder))
	lDupReq.Header.Set("Content-Type", "application/json")
	lDupRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lDupRec, lDupReq)

	if lDupRec.Code != http.StatusCreated {
		pT.Fatalf("Expected 201 for duplicate line items order, got %d. Body: %s", lDupRec.Code, lDupRec.Body.String())
	}

	lProd2, lFound2 := lStore.GetByID("prod-2")
	if !lFound2 {
		pT.Fatalf("Product prod-2 not found in store")
	}
	if lProd2.StockQuantity != 3 {
		pT.Errorf("Expected remaining stock for prod-2 to be 3, got %d", lProd2.StockQuantity)
	}

	// 5. Insufficient stock should return 400 with error code [STCA003] and rollback without partial deduction
	// prod-2 has 3 stock left. Requesting 5 should fail and stock should stay 3.
	lExcessOrder := `{"customer":"Too Much","items":[{"product_id":"prod-2","quantity":5}]}`
	lExcessReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lExcessOrder))
	lExcessReq.Header.Set("Content-Type", "application/json")
	lExcessRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lExcessRec, lExcessReq)

	if lExcessRec.Code != http.StatusBadRequest {
		pT.Fatalf("Expected 400 for excess order, got %d. Body: %s", lExcessRec.Code, lExcessRec.Body.String())
	}
	if !strings.Contains(lExcessRec.Body.String(), "STCA003") {
		pT.Fatalf("Expected error code STCA003 in body, got: %s", lExcessRec.Body.String())
	}

	// Verify rollback: stock of prod-2 must remain 3
	lProd2After, _ := lStore.GetByID("prod-2")
	if lProd2After.StockQuantity != 3 {
		pT.Errorf("Expected stock of prod-2 to remain 3 after failed checkout, got %d", lProd2After.StockQuantity)
	}

	// 6. Unknown product returns 400 with [STCA002]
	lUnknownOrder := `{"customer":"Unknown Prod","items":[{"product_id":"prod-999","quantity":1}]}`
	lUnknownReq := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(lUnknownOrder))
	lUnknownReq.Header.Set("Content-Type", "application/json")
	lUnknownRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lUnknownRec, lUnknownReq)

	if lUnknownRec.Code != http.StatusBadRequest {
		pT.Fatalf("Expected 400 for unknown product, got %d. Body: %s", lUnknownRec.Code, lUnknownRec.Body.String())
	}
	if !strings.Contains(lUnknownRec.Body.String(), "STCA002") {
		pT.Fatalf("Expected error code STCA002 in body, got: %s", lUnknownRec.Body.String())
	}
}

func TestSalesAnalytics(pT *testing.T) {
	lHandler, _ := setupTestRouter()

	// 1. GET /api/analytics/sales without date parameters (all orders)
	lAllReq := httptest.NewRequest(http.MethodGet, "/api/analytics/sales", nil)
	lAllRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lAllRec, lAllReq)

	if lAllRec.Code != http.StatusOK {
		pT.Fatalf("Expected 200 for analytics, got %d. Body: %s", lAllRec.Code, lAllRec.Body.String())
	}

	var lResp struct {
		Success bool                  `json:"success"`
		Data    models.SalesAnalytics `json:"data"`
	}
	if lErr := json.NewDecoder(lAllRec.Body).Decode(&lResp); lErr != nil {
		pT.Fatalf("Failed to decode analytics response: %v", lErr)
	}

	if !lResp.Success {
		pT.Errorf("Expected success to be true")
	}
	if lResp.Data.OrderCount < 3 {
		pT.Errorf("Expected at least 3 seeded orders, got %d", lResp.Data.OrderCount)
	}
	if lResp.Data.TotalRevenue <= 0 {
		pT.Errorf("Expected positive total revenue, got %.2f", lResp.Data.TotalRevenue)
	}
	if len(lResp.Data.TopProducts) == 0 {
		pT.Errorf("Expected top products to be populated")
	}

	// 2. GET with valid date range
	lDateReq := httptest.NewRequest(http.MethodGet, "/api/analytics/sales?startDate=2026-01-01&endDate=2026-12-31", nil)
	lDateRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lDateRec, lDateReq)

	if lDateRec.Code != http.StatusOK {
		pT.Fatalf("Expected 200 for date range analytics, got %d", lDateRec.Code)
	}

	// 3. GET with invalid date format -> [HNSA001]
	lInvalidReq := httptest.NewRequest(http.MethodGet, "/api/analytics/sales?startDate=not-a-date", nil)
	lInvalidRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lInvalidRec, lInvalidReq)

	if lInvalidRec.Code != http.StatusBadRequest {
		pT.Fatalf("Expected 400 for invalid date, got %d", lInvalidRec.Code)
	}
	if !strings.Contains(lInvalidRec.Body.String(), "HNSA001") {
		pT.Errorf("Expected error code HNSA001 in body, got: %s", lInvalidRec.Body.String())
	}

	// 4. GET with startDate after endDate -> [HNSA002]
	lInvertedReq := httptest.NewRequest(http.MethodGet, "/api/analytics/sales?startDate=2026-12-31&endDate=2026-01-01", nil)
	lInvertedRec := httptest.NewRecorder()
	lHandler.ServeHTTP(lInvertedRec, lInvertedReq)

	if lInvertedRec.Code != http.StatusBadRequest {
		pT.Fatalf("Expected 400 for inverted date range, got %d", lInvertedRec.Code)
	}
	if !strings.Contains(lInvertedRec.Body.String(), "HNSA002") {
		pT.Errorf("Expected error code HNSA002 in body, got: %s", lInvertedRec.Body.String())
	}
}
