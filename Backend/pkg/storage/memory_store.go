package storage

import (
	"fmt"
	"sync"

	"shoptrack-backend/pkg/models"
)

// MemoryStore provides thread-safe in-memory storage for products.
type MemoryStore struct {
	mu       sync.RWMutex
	products map[string]models.Product
}

// NewMemoryStore creates and seeds a new MemoryStore instance.
func NewMemoryStore() *MemoryStore {
	lStore := &MemoryStore{
		products: make(map[string]models.Product),
	}
	lStore.seed()
	return lStore
}

func (pStore *MemoryStore) seed() {
	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	lInitialProducts := []models.Product{
		{
			ID:            "prod-1",
			SKU:           "CHAIR-ERG-001",
			Name:          "Ergonomic Executive Office Chair",
			Price:         299.99,
			StockQuantity: 15,
		},
		{
			ID:            "prod-2",
			SKU:           "KB-MECH-002",
			Name:          "Mechanical RGB Gaming Keyboard",
			Price:         129.50,
			StockQuantity: 8,
		},
		{
			ID:            "prod-3",
			SKU:           "DOCK-USBC-003",
			Name:          "Triple-Display USB-C Docking Station",
			Price:         189.00,
			StockQuantity: 0,
		},
	}

	for _, lProd := range lInitialProducts {
		pStore.products[lProd.ID] = lProd
	}
}

// GetAll retrieves all products stored in memory.
func (pStore *MemoryStore) GetAll() ([]models.Product, error) {
	pStore.mu.RLock()
	defer pStore.mu.RUnlock()

	lList := make([]models.Product, 0, len(pStore.products))
	for _, lProd := range pStore.products {
		lList = append(lList, lProd)
	}
	return lList, nil
}

// GetByID retrieves a single product by its ID.
func (pStore *MemoryStore) GetByID(pID string) (models.Product, bool) {
	pStore.mu.RLock()
	defer pStore.mu.RUnlock()

	lProd, lExists := pStore.products[pID]
	return lProd, lExists
}

// Save stores or updates a product in memory.
func (pStore *MemoryStore) Save(pProduct models.Product) error {
	if pProduct.ID == "" {
		return fmt.Errorf("product ID cannot be empty")
	}

	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	pStore.products[pProduct.ID] = pProduct
	return nil
}
