package storage

import (
	"fmt"
	"log"
	"sort"
	"strings"
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
			Items: []models.OrderItem{
				{ProductID: "prod-1", Quantity: 1, UnitPrice: 24999.00},
				{ProductID: "prod-2", Quantity: 1, UnitPrice: 10499.00},
			},
		},
		{
			ID:          "ORD-1002",
			Customer:    "Priya Patel",
			ItemsCount:  1,
			TotalAmount: 24999.00,
			Status:      "Completed",
			CreatedAt:   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			Items: []models.OrderItem{
				{ProductID: "prod-1", Quantity: 1, UnitPrice: 24999.00},
			},
		},
		{
			ID:          "ORD-1003",
			Customer:    "Rohan Verma",
			ItemsCount:  3,
			TotalAmount: 51497.00,
			Status:      "Completed",
			CreatedAt:   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			Items: []models.OrderItem{
				{ProductID: "prod-1", Quantity: 1, UnitPrice: 24999.00},
				{ProductID: "prod-2", Quantity: 1, UnitPrice: 10499.00},
				{ProductID: "prod-3", Quantity: 1, UnitPrice: 15999.00},
			},
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

// SaveOrder stores a new sales order in memory with collision prevention.
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

	if _, lExists := pStore.orders[pOrder.ID]; lExists {
		lErr := fmt.Errorf("[STSO001] order id already exists: %s", pOrder.ID)
		log.Printf("[Storage][SaveOrder][ERROR] %v", lErr)
		return lErr
	}

	pStore.orders[pOrder.ID] = pOrder
	log.Printf("[Storage][SaveOrder] Successfully stored order ID: %s", pOrder.ID)
	return nil
}

// CheckoutAtomic performs an atomic inventory validation, deduction, price enforcement, and order creation.
func (pStore *MemoryStore) CheckoutAtomic(pCustomer string, pItems []models.OrderItem) (*models.Order, error) {
	log.Printf("[Storage][CheckoutAtomic][IN] Customer: %s, Items count: %d", pCustomer, len(pItems))
	defer log.Println("[Storage][CheckoutAtomic][OUT]")

	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	if len(pItems) == 0 {
		lErr := fmt.Errorf("[STCA001] order must contain at least one item")
		log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
		return nil, lErr
	}

	// Aggregate quantities by product_id to handle duplicate product lines in a single order payload
	lAggregated := make(map[string]int)
	for _, lItem := range pItems {
		if lItem.Quantity <= 0 {
			lErr := fmt.Errorf("[STCA001] item quantity must be greater than 0")
			log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
			return nil, lErr
		}
		lAggregated[lItem.ProductID] += lItem.Quantity
	}

	// Check stock and existence for every product
	for lProductID, lRequiredQty := range lAggregated {
		lProd, lExists := pStore.products[lProductID]
		if !lExists {
			lErr := fmt.Errorf("[STCA002] product not found: %s", lProductID)
			log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
			return nil, lErr
		}
		if lProd.StockQuantity < lRequiredQty {
			lErr := fmt.Errorf("[STCA003] insufficient stock for '%s': available %d, requested %d", lProd.Name, lProd.StockQuantity, lRequiredQty)
			log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
			return nil, lErr
		}
	}

	// Once ALL items pass validation, proceed to deduction and total calculation in the same lock
	var lTotalAmount float64
	var lItemsCount int
	var lOrderItems []models.OrderItem

	for _, lItem := range pItems {
		lProd := pStore.products[lItem.ProductID]
		lActualPrice := lProd.Price // Enforce store price, ignore client-provided price
		lTotalAmount += float64(lItem.Quantity) * lActualPrice
		lItemsCount += lItem.Quantity

		lOrderItems = append(lOrderItems, models.OrderItem{
			ProductID: lItem.ProductID,
			Quantity:  lItem.Quantity,
			UnitPrice: lActualPrice,
		})
	}

	// Deduct aggregated quantities from inventory
	for lProductID, lRequiredQty := range lAggregated {
		lProd := pStore.products[lProductID]
		lProd.StockQuantity -= lRequiredQty
		pStore.products[lProductID] = lProd
		log.Printf("[Storage][CheckoutAtomic] Deducted %d from '%s', remaining stock: %d", lRequiredQty, lProd.Name, lProd.StockQuantity)
	}

	// Generate a unique Order ID server-side
	lOrderID := fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	for {
		if _, lExists := pStore.orders[lOrderID]; !lExists {
			break
		}
		lOrderID = fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	}

	lCustomerName := strings.TrimSpace(pCustomer)
	if lCustomerName == "" {
		lCustomerName = "Walk-in Customer"
	}

	lOrder := models.Order{
		ID:          lOrderID,
		Customer:    lCustomerName,
		ItemsCount:  lItemsCount,
		TotalAmount: lTotalAmount,
		Status:      "Completed",
		CreatedAt:   time.Now().Format(time.RFC3339),
		Items:       lOrderItems,
	}

	pStore.orders[lOrderID] = lOrder
	log.Printf("[Storage][CheckoutAtomic] Order successfully saved: ID=%s, Customer=%s, Total=₹%.2f, Items=%d", lOrderID, lCustomerName, lTotalAmount, lItemsCount)
	return &lOrder, nil
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

// GetSalesAnalytics computes aggregate sales metrics and top selling products within an optional date range.
func (pStore *MemoryStore) GetSalesAnalytics(pStartDate *time.Time, pEndDate *time.Time) (*models.SalesAnalytics, error) {
	log.Println("[Storage][GetSalesAnalytics][IN]")
	defer log.Println("[Storage][GetSalesAnalytics][OUT]")

	pStore.mu.RLock()
	defer pStore.mu.RUnlock()

	var lTotalRevenue float64
	var lOrderCount int
	var lTotalUnitsSold int

	lProductSales := make(map[string]*models.TopSellingProduct)

	for _, lOrder := range pStore.orders {
		// Parse order CreatedAt
		lOrderTime, lErr := time.Parse(time.RFC3339, lOrder.CreatedAt)
		if lErr != nil {
			// Fallback for non-RFC3339 timestamps if any
			lOrderTime, lErr = time.Parse("2006-01-02", lOrder.CreatedAt)
		}

		// Apply date filters if parsed successfully
		if lErr == nil {
			if pStartDate != nil && lOrderTime.Before(*pStartDate) {
				continue
			}
			if pEndDate != nil && lOrderTime.After(*pEndDate) {
				continue
			}
		}

		lTotalRevenue += lOrder.TotalAmount
		lOrderCount++
		lTotalUnitsSold += lOrder.ItemsCount

		// Aggregate product line items
		for _, lItem := range lOrder.Items {
			lTopProd, lExists := lProductSales[lItem.ProductID]
			if !lExists {
				lName := lItem.ProductID
				if lProd, lFound := pStore.products[lItem.ProductID]; lFound {
					lName = lProd.Name
				}
				lTopProd = &models.TopSellingProduct{
					ProductID:   lItem.ProductID,
					Name:        lName,
					UnitsSold:   0,
					TotalEarned: 0,
				}
				lProductSales[lItem.ProductID] = lTopProd
			}
			lTopProd.UnitsSold += lItem.Quantity
			lTopProd.TotalEarned += float64(lItem.Quantity) * lItem.UnitPrice
		}
	}

	var lAverageOrderVal float64
	if lOrderCount > 0 {
		lAverageOrderVal = lTotalRevenue / float64(lOrderCount)
	}

	// Convert map to slice and sort descending by UnitsSold, then TotalEarned
	lTopList := make([]models.TopSellingProduct, 0, len(lProductSales))
	for _, lTopProd := range lProductSales {
		lTopList = append(lTopList, *lTopProd)
	}

	sort.Slice(lTopList, func(pI, pJ int) bool {
		if lTopList[pI].UnitsSold == lTopList[pJ].UnitsSold {
			return lTopList[pI].TotalEarned > lTopList[pJ].TotalEarned
		}
		return lTopList[pI].UnitsSold > lTopList[pJ].UnitsSold
	})

	// Take top 5
	if len(lTopList) > 5 {
		lTopList = lTopList[:5]
	}

	lPeriod := "all"
	if pStartDate != nil && pEndDate != nil {
		lPeriod = fmt.Sprintf("%s to %s", pStartDate.Format("2006-01-02"), pEndDate.Format("2006-01-02"))
	} else if pStartDate != nil {
		lPeriod = fmt.Sprintf("from %s", pStartDate.Format("2006-01-02"))
	} else if pEndDate != nil {
		lPeriod = fmt.Sprintf("until %s", pEndDate.Format("2006-01-02"))
	}

	lAnalytics := &models.SalesAnalytics{
		TotalRevenue:    lTotalRevenue,
		OrderCount:      lOrderCount,
		AverageOrderVal: lAverageOrderVal,
		TotalUnitsSold:  lTotalUnitsSold,
		TopProducts:     lTopList,
		Period:          lPeriod,
	}

	log.Printf("[Storage][GetSalesAnalytics] Revenue: ₹%.2f, Orders: %d, Units: %d, Period: %s", lTotalRevenue, lOrderCount, lTotalUnitsSold, lPeriod)
	return lAnalytics, nil
}

