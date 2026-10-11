package storage

import (
	"fmt"
	"log"
	"sync"
	"time"

	"shoptrack-backend/pkg/models"
)

// MemoryStore provides thread-safe in-memory storage for products and orders.
type MemoryStore struct {
	mu       sync.RWMutex
	products map[string]models.Product
	orders   map[string]models.Order
}

// NewMemoryStore creates and seeds a new MemoryStore instance.
func NewMemoryStore() *MemoryStore {
	log.Println("[Storage][NewMemoryStore][IN]")
	defer log.Println("[Storage][NewMemoryStore][OUT]")

	lStore := &MemoryStore{
		products: make(map[string]models.Product),
		orders:   make(map[string]models.Order),
	}
	lStore.seed()
	return lStore
}

func (pStore *MemoryStore) seed() {
	log.Println("[Storage][seed][IN]")
	defer log.Println("[Storage][seed][OUT]")

	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	lInitialProducts := []models.Product{
		{
			ID:            "prod-1",
			SKU:           "CHAIR-ERG-001",
			Name:          "Ergonomic Executive Office Chair",
			Price:         24999.00,
			StockQuantity: 15,
		},
		{
			ID:            "prod-2",
			SKU:           "KB-MECH-002",
			Name:          "Mechanical RGB Gaming Keyboard",
			Price:         10499.00,
			StockQuantity: 8,
		},
		{
			ID:            "prod-3",
			SKU:           "DOCK-USBC-003",
			Name:          "Triple-Display USB-C Docking Station",
			Price:         15999.00,
			StockQuantity: 0,
		},
	}

	for _, lProd := range lInitialProducts {
		pStore.products[lProd.ID] = lProd
	}

	lInitialOrders := []models.Order{
		{
			ID:          "ORD-1001",
			Customer:    "Aarav Sharma",
			ItemsCount:  2,
			TotalAmount: 35498.00,
			Status:      "Completed",
			CreatedAt:   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
		},
		{
			ID:          "ORD-1002",
			Customer:    "Priya Patel",
			ItemsCount:  1,
			TotalAmount: 24999.00,
			Status:      "Completed",
			CreatedAt:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			ID:          "ORD-1003",
			Customer:    "Rohan Verma",
			ItemsCount:  3,
			TotalAmount: 51497.00,
			Status:      "Completed",
			CreatedAt:   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
	}

	for _, lOrder := range lInitialOrders {
		pStore.orders[lOrder.ID] = lOrder
	}
	log.Printf("[Storage][seed] Seeded %d products and %d orders", len(lInitialProducts), len(lInitialOrders))
}

// GetAll retrieves all products stored in memory.
func (pStore *MemoryStore) GetAll() ([]models.Product, error) {
	log.Println("[Storage][GetAll][IN]")
	defer log.Println("[Storage][GetAll][OUT]")

	pStore.mu.RLock()
	defer pStore.mu.RUnlock()

	lList := make([]models.Product, 0, len(pStore.products))
	for _, lProd := range pStore.products {
		lList = append(lList, lProd)
	}
	log.Printf("[Storage][GetAll] Returning %d products", len(lList))
	return lList, nil
}

// GetByID retrieves a single product by its ID.
func (pStore *MemoryStore) GetByID(pID string) (models.Product, bool) {
	log.Printf("[Storage][GetByID][IN] ID: %s", pID)
	defer log.Printf("[Storage][GetByID][OUT] ID: %s", pID)

	pStore.mu.RLock()
	defer pStore.mu.RUnlock()

	lProd, lExists := pStore.products[pID]
	if !lExists {
		log.Printf("[Storage][GetByID] Product with ID '%s' not found", pID)
	}
	return lProd, lExists
}

// Save stores or updates a product in memory.
func (pStore *MemoryStore) Save(pProduct models.Product) error {
	log.Printf("[Storage][Save][IN] ID: %s, Name: %s", pProduct.ID, pProduct.Name)
	defer log.Printf("[Storage][Save][OUT] ID: %s", pProduct.ID)

	if pProduct.ID == "" {
		lErr := fmt.Errorf("product ID cannot be empty")
		log.Printf("[Storage][Save][ERROR] %v", lErr)
		return lErr
	}

	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	pStore.products[pProduct.ID] = pProduct
	log.Printf("[Storage][Save] Successfully stored product ID: %s", pProduct.ID)
	return nil
}

// GetOrders retrieves all sales orders stored in memory.
func (pStore *MemoryStore) GetOrders() ([]models.Order, error) {
	log.Println("[Storage][GetOrders][IN]")
	defer log.Println("[Storage][GetOrders][OUT]")

	pStore.mu.RLock()
	defer pStore.mu.RUnlock()

	lList := make([]models.Order, 0, len(pStore.orders))
	for _, lOrder := range pStore.orders {
		lList = append(lList, lOrder)
	}
	log.Printf("[Storage][GetOrders] Returning %d orders", len(lList))
	return lList, nil
}

// SaveOrder stores a new sales order in memory.
func (pStore *MemoryStore) SaveOrder(pOrder models.Order) error {
	log.Printf("[Storage][SaveOrder][IN] ID: %s, Customer: %s, Total: ₹%.2f", pOrder.ID, pOrder.Customer, pOrder.TotalAmount)
	defer log.Printf("[Storage][SaveOrder][OUT] ID: %s", pOrder.ID)

	if pOrder.ID == "" {
		lErr := fmt.Errorf("order ID cannot be empty")
		log.Printf("[Storage][SaveOrder][ERROR] %v", lErr)
		return lErr
	}

	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	pStore.orders[pOrder.ID] = pOrder
	log.Printf("[Storage][SaveOrder] Successfully stored order ID: %s", pOrder.ID)
	return nil
}

// DeductStock decreases product stock in a thread-safe manner.
func (pStore *MemoryStore) DeductStock(pProductID string, pQuantity int) error {
	log.Printf("[Storage][DeductStock][IN] ProductID: %s, Quantity: %d", pProductID, pQuantity)
	defer log.Printf("[Storage][DeductStock][OUT] ProductID: %s", pProductID)

	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	lProd, lExists := pStore.products[pProductID]
	if !lExists {
		lErr := fmt.Errorf("product '%s' not found", pProductID)
		log.Printf("[Storage][DeductStock][ERROR] %v", lErr)
		return lErr
	}
	if lProd.StockQuantity < pQuantity {
		lErr := fmt.Errorf("insufficient stock for '%s': available %d, requested %d", lProd.Name, lProd.StockQuantity, pQuantity)
		log.Printf("[Storage][DeductStock][ERROR] %v", lErr)
		return lErr
	}

	lProd.StockQuantity -= pQuantity
	pStore.products[pProductID] = lProd
	log.Printf("[Storage][DeductStock] Updated stock for '%s': remaining %d", lProd.Name, lProd.StockQuantity)
	return nil
}
