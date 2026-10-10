package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"shoptrack-backend/pkg/models"
	"shoptrack-backend/pkg/storage"
)

// ProductHandler handles incoming HTTP requests for products.
type ProductHandler struct {
	store *storage.MemoryStore
}

// NewProductHandler creates a new ProductHandler instance.
func NewProductHandler(pStore *storage.MemoryStore) *ProductHandler {
	return &ProductHandler{
		store: pStore,
	}
}

func (pHandler *ProductHandler) sendJSON(pWriter http.ResponseWriter, pStatusCode int, pData interface{}, pMessage string, pSuccess bool) {
	pWriter.Header().Set("Content-Type", "application/json")
	pWriter.WriteHeader(pStatusCode)

	lResp := models.StandardResponse{
		Success: pSuccess,
		Data:    pData,
		Message: pMessage,
	}
	_ = json.NewEncoder(pWriter).Encode(lResp)
}

// HealthCheckHandler returns the service health status.
func (pHandler *ProductHandler) HealthCheckHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	lHealthInfo := map[string]string{
		"status":    "UP",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	pHandler.sendJSON(pWriter, http.StatusOK, lHealthInfo, "Service is healthy", true)
}

// GetProductsHandler returns all inventory products.
func (pHandler *ProductHandler) GetProductsHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	lProducts, lErr := pHandler.store.GetAll()
	if lErr != nil {
		pHandler.sendJSON(pWriter, http.StatusInternalServerError, nil, "Failed to retrieve products", false)
		return
	}
	pHandler.sendJSON(pWriter, http.StatusOK, lProducts, "Products retrieved successfully", true)
}

// GetProductByIDHandler returns a specific product by its ID.
func (pHandler *ProductHandler) GetProductByIDHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	lVars := mux.Vars(pRequest)
	lID := lVars["id"]

	lProduct, lFound := pHandler.store.GetByID(lID)
	if !lFound {
		pHandler.sendJSON(pWriter, http.StatusNotFound, nil, fmt.Sprintf("Product with ID '%s' not found", lID), false)
		return
	}

	pHandler.sendJSON(pWriter, http.StatusOK, lProduct, "Product retrieved successfully", true)
}

// CreateProductHandler validates and creates a new product in memory storage.
func (pHandler *ProductHandler) CreateProductHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	var lPayload models.Product
	lErr := json.NewDecoder(pRequest.Body).Decode(&lPayload)
	if lErr != nil {
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Invalid JSON payload", false)
		return
	}

	lPayload.Name = strings.TrimSpace(lPayload.Name)
	lPayload.SKU = strings.TrimSpace(lPayload.SKU)

	if lPayload.SKU == "" {
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "SKU cannot be empty", false)
		return
	}
	if lPayload.Name == "" {
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Product name cannot be empty", false)
		return
	}
	if lPayload.Price <= 0 {
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Price must be greater than 0", false)
		return
	}
	if lPayload.StockQuantity < 0 {
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Stock quantity cannot be negative", false)
		return
	}

	if lPayload.ID == "" {
		lPayload.ID = fmt.Sprintf("prod-%d", time.Now().UnixNano())
	}

	lErr = pHandler.store.Save(lPayload)
	if lErr != nil {
		pHandler.sendJSON(pWriter, http.StatusInternalServerError, nil, "Failed to save product", false)
		return
	}

	pHandler.sendJSON(pWriter, http.StatusCreated, lPayload, "Product created successfully", true)
}
