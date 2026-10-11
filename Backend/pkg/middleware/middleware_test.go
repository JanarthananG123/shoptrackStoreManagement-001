package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"shoptrack-backend/pkg/middleware"
)

func TestCORSMiddleware_Options(pT *testing.T) {
	lCalled := false
	lNext := http.HandlerFunc(func(pW http.ResponseWriter, pR *http.Request) {
		lCalled = true
	})

	lHandler := middleware.CORSMiddleware(lNext)

	lReq := httptest.NewRequest(http.MethodOptions, "/test", nil)
	lRec := httptest.NewRecorder()

	lHandler.ServeHTTP(lRec, lReq)

	if lCalled {
		pT.Errorf("Expected next handler not to be called for OPTIONS, but it was")
	}

	if lRec.Code != http.StatusOK {
		pT.Errorf("Expected status code %d, got %d", http.StatusOK, lRec.Code)
	}

	if lRec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8081" {
		pT.Errorf("Expected Access-Control-Allow-Origin to be 'http://localhost:8081', got '%s'", lRec.Header().Get("Access-Control-Allow-Origin"))
	}
	if lRec.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		pT.Errorf("Expected Access-Control-Allow-Methods to be 'GET, POST, PUT, DELETE, OPTIONS', got '%s'", lRec.Header().Get("Access-Control-Allow-Methods"))
	}
	if lRec.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		pT.Errorf("Expected Access-Control-Allow-Headers to be 'Content-Type, Authorization', got '%s'", lRec.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestCORSMiddleware_NextCalled(pT *testing.T) {
	lCalled := false
	lNext := http.HandlerFunc(func(pW http.ResponseWriter, pR *http.Request) {
		lCalled = true
		pW.WriteHeader(http.StatusAccepted)
	})

	lHandler := middleware.CORSMiddleware(lNext)

	lReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	lRec := httptest.NewRecorder()

	lHandler.ServeHTTP(lRec, lReq)

	if !lCalled {
		pT.Errorf("Expected next handler to be called for GET, but it was not")
	}

	if lRec.Code != http.StatusAccepted {
		pT.Errorf("Expected status code %d, got %d", http.StatusAccepted, lRec.Code)
	}

	if lRec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8081" {
		pT.Errorf("Expected Access-Control-Allow-Origin to be 'http://localhost:8081', got '%s'", lRec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestRecoveryMiddleware_Panic(pT *testing.T) {
	lPanicHandler := http.HandlerFunc(func(pW http.ResponseWriter, pR *http.Request) {
		panic("simulated fatal crash")
	})

	lHandler := middleware.RecoveryMiddleware(lPanicHandler)

	lReq := httptest.NewRequest(http.MethodGet, "/panic", nil)
	lRec := httptest.NewRecorder()

	lHandler.ServeHTTP(lRec, lReq)

	if lRec.Code != http.StatusInternalServerError {
		pT.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, lRec.Code)
	}

	if lRec.Header().Get("Content-Type") != "application/json" {
		pT.Errorf("Expected Content-Type application/json, got %s", lRec.Header().Get("Content-Type"))
	}

	var lBody map[string]string
	lErr := json.NewDecoder(lRec.Body).Decode(&lBody)
	if lErr != nil {
		pT.Fatalf("Failed to decode response JSON: %v", lErr)
	}

	if lBody["error"] != "internal server error" {
		pT.Errorf("Expected error 'internal server error', got '%s'", lBody["error"])
	}
}
