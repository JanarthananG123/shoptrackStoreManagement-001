package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"shoptrack-backend/pkg/models"
	"shoptrack-backend/pkg/storage"
)

// ProductHandler handles incoming HTTP requests for products and orders.
type ProductHandler struct {
	store *storage.MemoryStore
}

// NewProductHandler creates a new ProductHandler instance.
func NewProductHandler(pStore *storage.MemoryStore) *ProductHandler {
	log.Println("[ProductHandler][NewProductHandler][IN]")
	defer log.Println("[ProductHandler][NewProductHandler][OUT]")
	return &ProductHandler{
		store: pStore,
	}
}

func (pHandler *ProductHandler) sendJSON(pWriter http.ResponseWriter, pStatusCode int, pData interface{}, pMessage string, pSuccess bool) {
	log.Printf("[ProductHandler][sendJSON] StatusCode: %d, Success: %t, Message: %s", pStatusCode, pSuccess, pMessage)
	pWriter.Header().Set("Content-Type", "application/json")
	pWriter.WriteHeader(pStatusCode)

	lResp := models.StandardResponse{
		Success: pSuccess,
		Data:    pData,
		Message: pMessage,
	}
	if lErr := json.NewEncoder(pWriter).Encode(lResp); lErr != nil {
		log.Printf("[ProductHandler][sendJSON][ERROR] Failed to encode JSON response: %v", lErr)
	}
}

// HealthCheckHandler returns the service health status.
func (pHandler *ProductHandler) HealthCheckHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	log.Println("[ProductHandler][HealthCheckHandler][IN]")
	defer log.Println("[ProductHandler][HealthCheckHandler][OUT]")

	lHealthInfo := map[string]string{
		"status":    "UP",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	pHandler.sendJSON(pWriter, http.StatusOK, lHealthInfo, "Service is healthy", true)
}

// GetProductsHandler returns all inventory products.
func (pHandler *ProductHandler) GetProductsHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	log.Println("[ProductHandler][GetProductsHandler][IN]")
	defer log.Println("[ProductHandler][GetProductsHandler][OUT]")

	lProducts, lErr := pHandler.store.GetAll()
	if lErr != nil {
		log.Printf("[ProductHandler][GetProductsHandler][ERROR] Failed to retrieve products: %v", lErr)
		pHandler.sendJSON(pWriter, http.StatusInternalServerError, nil, "Failed to retrieve products", false)
		return
	}
	log.Printf("[ProductHandler][GetProductsHandler] Retrieved %d products successfully", len(lProducts))
	pHandler.sendJSON(pWriter, http.StatusOK, lProducts, "Products retrieved successfully", true)
}

// GetProductByIDHandler returns a specific product by its ID.
func (pHandler *ProductHandler) GetProductByIDHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	lVars := mux.Vars(pRequest)
	lID := lVars["id"]

	log.Printf("[ProductHandler][GetProductByIDHandler][IN] ID: %s", lID)
	defer log.Printf("[ProductHandler][GetProductByIDHandler][OUT] ID: %s", lID)

	lProduct, lFound := pHandler.store.GetByID(lID)
	if !lFound {
		log.Printf("[ProductHandler][GetProductByIDHandler][ERROR] Product with ID '%s' not found", lID)
		pHandler.sendJSON(pWriter, http.StatusNotFound, nil, fmt.Sprintf("Product with ID '%s' not found", lID), false)
		return
	}

	pHandler.sendJSON(pWriter, http.StatusOK, lProduct, "Product retrieved successfully", true)
}

// CreateProductHandler validates and creates a new product in memory storage.
func (pHandler *ProductHandler) CreateProductHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	log.Println("[ProductHandler][CreateProductHandler][IN]")
	defer log.Println("[ProductHandler][CreateProductHandler][OUT]")

	var lPayload models.Product
	lErr := json.NewDecoder(pRequest.Body).Decode(&lPayload)
	if lErr != nil {
		log.Printf("[ProductHandler][CreateProductHandler][ERROR] Invalid JSON payload: %v", lErr)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Invalid JSON payload", false)
		return
	}

	lPayload.Name = strings.TrimSpace(lPayload.Name)
	lPayload.SKU = strings.TrimSpace(lPayload.SKU)

	if lPayload.SKU == "" {
		log.Println("[ProductHandler][CreateProductHandler][ERROR] SKU cannot be empty")
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "SKU cannot be empty", false)
		return
	}
	if lPayload.Name == "" {
		log.Println("[ProductHandler][CreateProductHandler][ERROR] Product name cannot be empty")
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Product name cannot be empty", false)
		return
	}
	if lPayload.Price <= 0 {
		log.Printf("[ProductHandler][CreateProductHandler][ERROR] Invalid price: %f", lPayload.Price)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Price must be greater than 0", false)
		return
	}
	if lPayload.StockQuantity < 0 {
		log.Printf("[ProductHandler][CreateProductHandler][ERROR] Invalid stock quantity: %d", lPayload.StockQuantity)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "Stock quantity cannot be negative", false)
		return
	}

	if lPayload.ID == "" {
		lPayload.ID = fmt.Sprintf("prod-%d", time.Now().UnixNano())
	}

	lErr = pHandler.store.Save(lPayload)
	if lErr != nil {
		log.Printf("[ProductHandler][CreateProductHandler][ERROR] Failed to save product '%s': %v", lPayload.ID, lErr)
		pHandler.sendJSON(pWriter, http.StatusInternalServerError, nil, "Failed to save product", false)
		return
	}

	log.Printf("[ProductHandler][CreateProductHandler] Product created successfully: ID=%s, SKU=%s, Name=%s", lPayload.ID, lPayload.SKU, lPayload.Name)
	pHandler.sendJSON(pWriter, http.StatusCreated, lPayload, "Product created successfully", true)
}

// GetOrdersHandler retrieves all completed sales orders.
func (pHandler *ProductHandler) GetOrdersHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	log.Println("[ProductHandler][GetOrdersHandler][IN]")
	defer log.Println("[ProductHandler][GetOrdersHandler][OUT]")

	lOrders, lErr := pHandler.store.GetOrders()
	if lErr != nil {
		log.Printf("[ProductHandler][GetOrdersHandler][ERROR] Failed to retrieve orders: %v", lErr)
		pHandler.sendJSON(pWriter, http.StatusInternalServerError, nil, "Failed to retrieve orders", false)
		return
	}
	log.Printf("[ProductHandler][GetOrdersHandler] Retrieved %d orders successfully", len(lOrders))
	pHandler.sendJSON(pWriter, http.StatusOK, lOrders, "Orders retrieved successfully", true)
}

// CheckoutOrderHandler handles order checkout transactions.
func (pHandler *ProductHandler) CheckoutOrderHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	log.Println("[ProductHandler][CheckoutOrderHandler][IN]")
	defer log.Println("[ProductHandler][CheckoutOrderHandler][OUT]")

	type lCheckoutPayload struct {
		ID          string             `json:"id"`
		Customer    string             `json:"customer"`
		ItemsCount  int                `json:"itemsCount"`
		TotalAmount float64            `json:"totalAmount"`
		Status      string             `json:"status"`
		CreatedAt   string             `json:"createdAt"`
		Items       []models.OrderItem `json:"items"`
	}

	var lPayload lCheckoutPayload
	lErr := json.NewDecoder(pRequest.Body).Decode(&lPayload)
	if lErr != nil {
		log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO001] Decode error: %v", lErr)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNCO001] Invalid order payload", false)
		return
	}

	lCustomer := strings.TrimSpace(lPayload.Customer)
	if lCustomer == "" {
		lCustomer = "Walk-in Customer"
	}

	lTotalAmount := lPayload.TotalAmount
	lItemsCount := lPayload.ItemsCount

	if len(lPayload.Items) > 0 {
		lTotalAmount = 0
		lItemsCount = 0

		// Validate all items and stock availability first
		for _, lItem := range lPayload.Items {
			if lItem.Quantity <= 0 {
				log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO001] Invalid item quantity: %d", lItem.Quantity)
				pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNCO001] Item quantity must be greater than 0", false)
				return
			}
			lProd, lExists := pHandler.store.GetByID(lItem.ProductID)
			if !lExists {
				log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO003] Product '%s' not found", lItem.ProductID)
				pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, fmt.Sprintf("[HNCO003] Product '%s' not found", lItem.ProductID), false)
				return
			}
			if lProd.StockQuantity < lItem.Quantity {
				log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO002] Insufficient stock for '%s': requested %d, available %d", lProd.Name, lItem.Quantity, lProd.StockQuantity)
				pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, fmt.Sprintf("[HNCO002] Insufficient stock for '%s' (available: %d, requested: %d)", lProd.Name, lProd.StockQuantity, lItem.Quantity), false)
				return
			}
		}

		// Deduct stock and calculate order amounts
		for _, lItem := range lPayload.Items {
			lDeductErr := pHandler.store.DeductStock(lItem.ProductID, lItem.Quantity)
			if lDeductErr != nil {
				log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO002] Stock deduction error for '%s': %v", lItem.ProductID, lDeductErr)
				pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, fmt.Sprintf("[HNCO002] %v", lDeductErr), false)
				return
			}
			lUnitPrice := lItem.UnitPrice
			if lUnitPrice <= 0 {
				if lProd, lExists := pHandler.store.GetByID(lItem.ProductID); lExists {
					lUnitPrice = lProd.Price
				}
			}
			lTotalAmount += float64(lItem.Quantity) * lUnitPrice
			lItemsCount += lItem.Quantity
		}
	} else if lTotalAmount <= 0 && lItemsCount <= 0 {
		log.Println("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO001] Order must contain at least one item")
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNCO001] Order must contain at least one item", false)
		return
	}

	lOrderID := lPayload.ID
	if lOrderID == "" {
		lOrderID = fmt.Sprintf("ORD-%d", time.Now().UnixNano()/1000000)
	}

	lCreatedAt := lPayload.CreatedAt
	if lCreatedAt == "" {
		lCreatedAt = time.Now().Format(time.RFC3339)
	}

	lStatus := lPayload.Status
	if lStatus == "" {
		lStatus = "Completed"
	}

	lOrder := models.Order{
		ID:          lOrderID,
		Customer:    lCustomer,
		ItemsCount:  lItemsCount,
		TotalAmount: lTotalAmount,
		Status:      lStatus,
		CreatedAt:   lCreatedAt,
		Items:       lPayload.Items,
	}

	lErr = pHandler.store.SaveOrder(lOrder)
	if lErr != nil {
		log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO004] Failed to save order '%s': %v", lOrderID, lErr)
		pHandler.sendJSON(pWriter, http.StatusInternalServerError, nil, "[HNCO004] Failed to save order", false)
		return
	}

	log.Printf("[ProductHandler][CheckoutOrderHandler] Order checkout successful: ID=%s, Customer=%s, Total=₹%.2f, Items=%d", lOrder.ID, lOrder.Customer, lOrder.TotalAmount, lOrder.ItemsCount)
	pHandler.sendJSON(pWriter, http.StatusCreated, lOrder, "Order completed successfully", true)
}
