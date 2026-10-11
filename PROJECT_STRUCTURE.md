# ShopTrack - Project Structure & Audit Inventory

Comprehensive snapshot and architectural inventory of the **ShopTrack** repository, documenting backend modules, storage methods, handlers, middlewares, frontend routing, services, components, and complete active source dumps.

---

## 1. Directory Tree

```
D:\ShopTrack\
├── .gitignore
├── CommonConfiguration\
│   └── DbConfig.txt
├── PROJECT_STRUCTURE.md
├── README.md
├── Backend\
│   ├── go.mod
│   ├── go.sum
│   ├── cmd\
│   │   └── server\
│   │       ├── main.go
│   │       └── server_test.go
│   └── pkg\
│       ├── handlers\
│       │   └── product_handler.go
│       ├── logger\
│       │   ├── logger.go
│       │   └── logger_test.go
│       ├── middleware\
│       │   ├── cors.go
│       │   ├── logger.go
│       │   ├── middleware_test.go
│       │   └── recovery.go
│       ├── models\
│       │   └── models.go
│       └── storage\
│           └── memory_store.go
└── frontend\
    ├── .env.development
    ├── package-lock.json
    ├── package.json
    ├── vue.config.js
    ├── public\
    │   ├── favicon.ico
    │   └── index.html
    └── src\
        ├── App.vue
        ├── main.js
        ├── api\
        │   └── axiosConfig.js
        ├── assets\
        │   └── app.css
        ├── components\
        │   ├── AppNavbar.vue
        │   └── ProductDialog.vue
        ├── plugins\
        │   └── vuetify.js
        ├── router\
        │   └── index.js
        ├── services\
        │   ├── apiClient.js
        │   ├── index.js
        │   ├── orderService.js
        │   └── productService.js
        └── views\
            ├── ProductsView.vue
            └── SalesView.vue
```

*(Note: `node_modules/`, `.git/`, `dist/`, and runtime log folders under `Backend/logs/` are excluded from inventory analysis).*

---

## 2. Backend Inventory

### Data Models (`Backend/pkg/models/models.go`)
- **`Product`**:
  - `ID`: `string` (`json:"id"`)
  - `SKU`: `string` (`json:"sku"`)
  - `Name`: `string` (`json:"name"`)
  - `Price`: `float64` (`json:"price"`)
  - `StockQuantity`: `int` (`json:"stockQuantity"`)
- **`OrderItem`**:
  - `ProductID`: `string` (`json:"product_id"`)
  - `Quantity`: `int` (`json:"quantity"`)
  - `UnitPrice`: `float64` (`json:"unit_price"`)
- **`Order`**:
  - `ID`: `string` (`json:"id"`)
  - `Customer`: `string` (`json:"customer"`)
  - `ItemsCount`: `int` (`json:"itemsCount"`)
  - `TotalAmount`: `float64` (`json:"totalAmount"`)
  - `Status`: `string` (`json:"status"`)
  - `CreatedAt`: `string` (`json:"createdAt"`)
  - `Items`: `[]OrderItem` (`json:"items,omitempty"`)
- **`TopSellingProduct`**:
  - `ProductID`: `string` (`json:"product_id"`)
  - `Name`: `string` (`json:"name"`)
  - `UnitsSold`: `int` (`json:"units_sold"`)
  - `TotalEarned`: `float64` (`json:"total_earned"`)
- **`SalesAnalytics`**:
  - `TotalRevenue`: `float64` (`json:"total_revenue"`),
  - `OrderCount`: `int` (`json:"order_count"`),
  - `AverageOrderVal`: `float64` (`json:"average_order_value"`),
  - `TotalUnitsSold`: `int` (`json:"total_units_sold"`),
  - `TopProducts`: `[]TopSellingProduct` (`json:"top_products"`),
  - `Period`: `string` (`json:"period"`)
- **`StandardResponse`**:
  - `Success`: `bool` (`json:"success"`)
  - `Data`: `interface{}` (`json:"data,omitempty"`)
  - `Message`: `string` (`json:"message,omitempty"`)

---

### Storage Methods (`Backend/pkg/storage/memory_store.go`)
- **`NewMemoryStore() *MemoryStore`**: Allocates memory maps and seeds default initial products and orders.
- **`(*MemoryStore) seed()`**: Thread-safe initialization of default products (`prod-1`, `prod-2`, `prod-3`) and orders (`ORD-1001`, `ORD-1002`, `ORD-1003`) with item breakdowns.
- **`(*MemoryStore) GetAll() ([]models.Product, error)`**: Returns copy slice of all inventory products under read lock.
- **`(*MemoryStore) GetByID(pID string) (models.Product, bool)`**: Retrieves single product by ID under read lock.
- **`(*MemoryStore) Save(pProduct models.Product) error`**: Inserts or updates inventory product under write lock.
- **`(*MemoryStore) GetOrders() ([]models.Order, error)`**: Retrieves all completed sales orders under read lock.
- **`(*MemoryStore) CheckoutAtomic(pCustomer string, pItems []models.OrderItem) (*models.Order, error)`**:
  - Executes under single `sync.RWMutex.Lock()`.
  - Enforces sanity bounds (`[STCA004]` customer <= 100 UTF-8 runes, `[STCA005]` items <= 50, item quantity 1..10,000, addition overflow protection).
  - Rejects empty orders (`[STCA001]`).
  - Aggregates quantities by `product_id` across duplicate line items.
  - Validates item existence (`[STCA002]`) and inventory stock adequacy (`[STCA003]`).
  - Enforces server product catalog pricing (`lProd.Price`), discarding any client-provided prices.
  - Consolidates order items to exactly one line per unique `product_id`.
  - Atomically decrements quantities and records `models.Order` with collision-free server ID (`ORD-<timestamp>`).
- **`(*MemoryStore) GetSalesAnalytics(pStartDate *time.Time, pEndDate *time.Time) (*models.SalesAnalytics, error)`**:
  - Executes under `sync.RWMutex.RLock()`.
  - Filters orders by date range if provided.
  - Aggregates revenue, order volume, AOV, and units sold.
  - Aggregates and ranks Top 5 best-selling products by units sold and revenue.

---

### HTTP Endpoints & Handlers (`Backend/cmd/server/main.go` & `Backend/pkg/handlers/product_handler.go`)
All routes are registered on both `/api` and `/api/v1` prefixes via Gorilla Mux:

| HTTP Method | Endpoint Paths | Handler Method | Description |
|---|---|---|---|
| `GET` | `/api/getHealth`<br>`/api/v1/getHealth` | `HealthCheckHandler` | Service heartbeat & ISO timestamp |
| `GET` | `/api/getProducts`<br>`/api/v1/getProducts` | `GetProductsHandler` | List all inventory products |
| `GET` | `/api/getProducts/{id}`<br>`/api/v1/getProducts/{id}` | `GetProductByIDHandler` | Retrieve product details by ID |
| `POST` | `/api/getProducts`<br>`/api/v1/getProducts` | `CreateProductHandler` | Validate & add new product to store |
| `GET` | `/api/orders`<br>`/api/v1/orders` | `GetOrdersHandler` | List all completed sales orders |
| `POST` | `/api/orders`<br>`/api/v1/orders` | `CheckoutOrderHandler` | Validate items, execute atomic checkout, enforce server prices |
| `GET` | `/api/analytics/sales`<br>`/api/getSalesAnalytics`<br>`/api/v1/analytics/sales` | `GetSalesAnalyticsHandler` | Aggregated sales reporting & top 5 products with optional date filter |
| `OPTIONS` | `/*` (Global preflight) | Handled by `CORSMiddleware` | Returns HTTP 200 OK with allowed headers & methods (`http://localhost:8081`) |
| `GET` | `/panic`<br>`/api/panic`<br>`/api/v1/panic` | Inline closure | Simulates runtime panic for middleware recovery verification |

---

### Middlewares (`Backend/pkg/middleware/`)
1. **`RequestLoggerMiddleware` (`logger.go`)**:
   - Logs incoming method, URL, remote IP, response status, and request duration.
   - Logs error indicators for HTTP status codes >= 400.
2. **`CORSMiddleware` (`cors.go`)**:
   - Restricts `Access-Control-Allow-Origin` to `http://localhost:8081`.
   - Exposes methods: `GET, POST, PUT, DELETE, OPTIONS`.
   - Exposes headers: `Content-Type, Authorization`.
   - Returns HTTP 200 OK immediately for preflight `OPTIONS` requests.
3. **`RecoveryMiddleware` (`recovery.go`)**:
   - Captures panics via `defer ... recover()`.
   - Emits structured debug stack trace logs.
   - Returns clean HTTP 500 JSON payload: `{"error": "internal server error"}`.

---

## 3. Frontend Inventory

### Modular Service Layer (`frontend/src/services/`)
- **`apiClient.js`**: Preconfigured Axios instance with `baseURL: 'http://localhost:8080/api'`, 5000ms timeout, and JSON headers.
- **`productService.js`**:
  - `fetchProducts()`: `GET /getProducts`
  - `createProduct(pData)`: `POST /getProducts`
  - `updateStock(pId, pQuantity)`: `PATCH /getProducts/:id`
- **`orderService.js`**:
  - `fetchOrders()`: `GET /orders`
  - `checkoutOrder(pPayload)`: `POST /orders`
  - `fetchSalesAnalytics(pParams)`: `GET /analytics/sales` with query params
- **`index.js` (`EventService`)**:
  - Consolidates domain submodules: `EventService.products`, `EventService.orders`.
  - Direct convenience methods: `fetchProducts()`, `createProduct()`, `updateStock()`, `fetchOrders()`, `checkoutOrder()`, `fetchSalesAnalytics()`, `checkHealth()`.

---

### Router Configuration (`frontend/src/router/index.js`)
- `/` -> `ProductsView.vue` (Name: `Products`)
- `/products` -> Redirects to `/`
- `/sales` -> `SalesView.vue` (Name: `Sales`)
- Mode: `history`

---

### UI Views & Components
1. **`AppNavbar.vue` (`frontend/src/components/AppNavbar.vue`)**:
   - Store branding header (`Codzee Store & Inventory Manager`).
   - Route tabs for **Inventory** (`/`) and **Sales** (`/sales`).
   - Dynamic API Status indicator chip (`ONLINE` / `OFFLINE`).
   - Light/Dark theme toggle button (`$vuetify.theme.dark`).
2. **`ProductsView.vue` (`frontend/src/views/ProductsView.vue`)**:
   - Top KPI cards: Total Products, Low Stock Warning (<=10), Out of Stock (0).
   - Searchable Vuetify data table for inventory with stock status chips.
   - Action controls: Refresh inventory, Add Product button, Product detail modal.
3. **`ProductDialog.vue` (`frontend/src/components/ProductDialog.vue`)**:
   - Input fields: SKU Code, Product Name, Unit Price (₹), Stock Quantity.
   - Client-side validation rules and submission via `EventService.createProduct()`.
4. **`SalesView.vue` (`frontend/src/views/SalesView.vue`)**:
   - **Executive KPI Cards**: Real-time synchronized metrics for Total Revenue (₹), Completed Orders, Average Order Value (₹), Units Sold.
   - **Date Filter Toolbar**: Period presets (`All Time`, `Today`, `Last 7 Days`, `Last 30 Days`, `Custom Range`) with start/end date selectors and filter validation.
   - **CSV Export**: One-click transaction export to CSV with proper escaping and error catching (`[SVSA001]`).
   - **Completed Sales History**: Interactive data table listing completed orders with item details, formatted dates, status badges, and receipt detail modal.
   - **Top 5 Best-Selling Products Widget**: Rankings (#1 to #5), units sold, revenue earned in ₹, and volume progress bars.
   - **Record Sale Modal Dialog**:
     - Dynamic line items selection with product search and live stock badges.
     - Live line totals and grand total calculations formatted in ₹.
     - Quantity boundary validation against inventory (`[SVSO001]`, `[SVSO002]`).
     - Submits payload to `EventService.orders.checkoutOrder()`.

---

## 4. File Contents Dump

### `Backend/pkg/models/models.go`
```go
package models

// Product represents an inventory item in the store.
type Product struct {
	ID            string  `json:"id"`
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	StockQuantity int     `json:"stockQuantity"`
}

// OrderItem represents a line item in an order.
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// Order represents a sales order transaction in the store.
type Order struct {
	ID          string      `json:"id"`
	Customer    string      `json:"customer"`
	ItemsCount  int         `json:"itemsCount"`
	TotalAmount float64     `json:"totalAmount"`
	Status      string      `json:"status"`
	CreatedAt   string      `json:"createdAt"`
	Items       []OrderItem `json:"items,omitempty"`
}

// StandardResponse defines a unified API response structure.
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// TopSellingProduct represents product sales volume and earnings.
type TopSellingProduct struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	UnitsSold   int     `json:"units_sold"`
	TotalEarned float64 `json:"total_earned"`
}

// SalesAnalytics aggregates store revenue, order volume, and top items.
type SalesAnalytics struct {
	TotalRevenue    float64             `json:"total_revenue"`
	OrderCount      int                 `json:"order_count"`
	AverageOrderVal float64             `json:"average_order_value"`
	TotalUnitsSold  int                 `json:"total_units_sold"`
	TopProducts     []TopSellingProduct `json:"top_products"`
	Period          string              `json:"period"`
}
```

---

### `Backend/pkg/storage/memory_store.go`
```go
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

// CheckoutError defines a structured error for checkout operations with code and message.
type CheckoutError struct {
	Code    string
	Message string
}

func (pErr *CheckoutError) Error() string {
	return fmt.Sprintf("[%s] %s", pErr.Code, pErr.Message)
}

// CheckoutAtomic performs an atomic inventory validation, deduction, price enforcement, and order creation.
func (pStore *MemoryStore) CheckoutAtomic(pCustomer string, pItems []models.OrderItem) (*models.Order, error) {
	log.Printf("[Storage][CheckoutAtomic][IN] Customer: %s, Items count: %d", pCustomer, len(pItems))
	defer log.Println("[Storage][CheckoutAtomic][OUT]")

	pStore.mu.Lock()
	defer pStore.mu.Unlock()

	if utf8.RuneCountInString(pCustomer) > 100 {
		lErr := &CheckoutError{Code: "STCA004", Message: "customer name too long"}
		log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
		return nil, lErr
	}

	if len(pItems) == 0 {
		lErr := &CheckoutError{Code: "STCA001", Message: "order must contain at least one item"}
		log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
		return nil, lErr
	}

	if len(pItems) > 50 {
		lErr := &CheckoutError{Code: "STCA005", Message: "too many line items"}
		log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
		return nil, lErr
	}

	// Aggregate quantities by product_id to handle duplicate product lines in a single order payload
	lAggregated := make(map[string]int)
	var lSeenProductIDs []string
	for _, lItem := range pItems {
		if lItem.Quantity <= 0 || lItem.Quantity > 10000 {
			lErr := &CheckoutError{Code: "STCA001", Message: "item quantity must be between 1 and 10,000"}
			log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
			return nil, lErr
		}
		if _, lExists := lAggregated[lItem.ProductID]; !lExists {
			lSeenProductIDs = append(lSeenProductIDs, lItem.ProductID)
		}
		if lAggregated[lItem.ProductID] > 1000000-lItem.Quantity {
			lErr := &CheckoutError{Code: "STCA001", Message: "accumulated quantity exceeds maximum permissible limit"}
			log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
			return nil, lErr
		}
		lAggregated[lItem.ProductID] += lItem.Quantity
	}

	// Check stock and existence for every product
	for lProductID, lRequiredQty := range lAggregated {
		lProd, lExists := pStore.products[lProductID]
		if !lExists {
			lErr := &CheckoutError{Code: "STCA002", Message: fmt.Sprintf("product not found: %s", lProductID)}
			log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
			return nil, lErr
		}
		if lProd.StockQuantity < lRequiredQty {
			lErr := &CheckoutError{Code: "STCA003", Message: fmt.Sprintf("insufficient stock for '%s': available %d, requested %d", lProd.Name, lProd.StockQuantity, lRequiredQty)}
			log.Printf("[Storage][CheckoutAtomic][ERROR] %v", lErr)
			return nil, lErr
		}
	}

	// Once ALL items pass validation, proceed to deduction and total calculation in the same lock
	var lTotalAmount float64
	var lItemsCount int
	var lOrderItems []models.OrderItem

	// Build lOrderItems using the aggregated product map so stored orders contain exactly one consolidated line per unique product_id
	for _, lProductID := range lSeenProductIDs {
		lQty := lAggregated[lProductID]
		lProd := pStore.products[lProductID]
		lActualPrice := lProd.Price // Enforce store price, ignore client-provided price
		lTotalAmount += float64(lQty) * lActualPrice
		lItemsCount += lQty

		lOrderItems = append(lOrderItems, models.OrderItem{
			ProductID: lProductID,
			Quantity:  lQty,
			UnitPrice: lActualPrice,
		})
	}

	// Deduct aggregated quantities from inventory
	for _, lProductID := range lSeenProductIDs {
		lRequiredQty := lAggregated[lProductID]
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
```

---

### `Backend/pkg/handlers/product_handler.go`
```go
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

// CheckoutOrderHandler handles order checkout transactions atomically.
func (pHandler *ProductHandler) CheckoutOrderHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	log.Println("[ProductHandler][CheckoutOrderHandler][IN]")
	defer log.Println("[ProductHandler][CheckoutOrderHandler][OUT]")

	type lCheckoutPayload struct {
		Customer string             `json:"customer"`
		Items    []models.OrderItem `json:"items"`
	}

	var lPayload lCheckoutPayload
	lErr := json.NewDecoder(pRequest.Body).Decode(&lPayload)
	if lErr != nil {
		log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO001] Decode error: %v", lErr)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNCO001] Invalid order payload", false)
		return
	}

	if len(lPayload.Items) == 0 {
		log.Println("[ProductHandler][CheckoutOrderHandler][ERROR] [HNCO001] Order must contain at least one item")
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNCO001] Order must contain at least one item", false)
		return
	}

	// Atomically validate inventory, enforce server-side prices, deduct stock, and save order
	lOrder, lCheckoutErr := pHandler.store.CheckoutAtomic(lPayload.Customer, lPayload.Items)
	if lCheckoutErr != nil {
		log.Printf("[ProductHandler][CheckoutOrderHandler][ERROR] Checkout failed: %v", lCheckoutErr)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, lCheckoutErr.Error(), false)
		return
	}

	log.Printf("[ProductHandler][CheckoutOrderHandler] Order checkout successful: ID=%s, Customer=%s, Total=₹%.2f, Items=%d", lOrder.ID, lOrder.Customer, lOrder.TotalAmount, lOrder.ItemsCount)
	pHandler.sendJSON(pWriter, http.StatusCreated, lOrder, "Order completed successfully", true)
}

// ParseAnalyticsDate parses ISO RFC3339 or YYYY-MM-DD query parameter strings.
func ParseAnalyticsDate(pDateStr string, pIsEndOfDay bool) (*time.Time, error) {
	lTrimmed := strings.TrimSpace(pDateStr)
	if lTrimmed == "" {
		return nil, nil
	}

	// Try RFC3339
	if lParsed, lErr := time.Parse(time.RFC3339, lTrimmed); lErr == nil {
		return &lParsed, nil
	}

	// Try YYYY-MM-DD
	if lParsed, lErr := time.Parse("2006-01-02", lTrimmed); lErr == nil {
		if pIsEndOfDay {
			lEndOfDay := time.Date(lParsed.Year(), lParsed.Month(), lParsed.Day(), 23, 59, 59, 999999999, time.UTC)
			return &lEndOfDay, nil
		}
		lStartOfDay := time.Date(lParsed.Year(), lParsed.Month(), lParsed.Day(), 0, 0, 0, 0, time.UTC)
		return &lStartOfDay, nil
	}

	return nil, fmt.Errorf("invalid date format: %s", pDateStr)
}

// GetSalesAnalyticsHandler computes and returns sales analytics with optional date filtering.
func (pHandler *ProductHandler) GetSalesAnalyticsHandler(pWriter http.ResponseWriter, pRequest *http.Request) {
	log.Println("[ProductHandler][GetSalesAnalyticsHandler][IN]")
	defer log.Println("[ProductHandler][GetSalesAnalyticsHandler][OUT]")

	lStartStr := pRequest.URL.Query().Get("startDate")
	lEndStr := pRequest.URL.Query().Get("endDate")

	lStartDate, lStartErr := ParseAnalyticsDate(lStartStr, false)
	if lStartErr != nil {
		log.Printf("[ProductHandler][GetSalesAnalyticsHandler][ERROR] [HNSA001] Invalid startDate: %v", lStartErr)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNSA001] Invalid startDate format (expected RFC3339 or YYYY-MM-DD)", false)
		return
	}

	lEndDate, lEndErr := ParseAnalyticsDate(lEndStr, true)
	if lEndErr != nil {
		log.Printf("[ProductHandler][GetSalesAnalyticsHandler][ERROR] [HNSA001] Invalid endDate: %v", lEndErr)
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNSA001] Invalid endDate format (expected RFC3339 or YYYY-MM-DD)", false)
		return
	}

	if lStartDate != nil && lEndDate != nil && lStartDate.After(*lEndDate) {
		log.Println("[ProductHandler][GetSalesAnalyticsHandler][ERROR] [HNSA002] startDate cannot be after endDate")
		pHandler.sendJSON(pWriter, http.StatusBadRequest, nil, "[HNSA002] startDate cannot be after endDate", false)
		return
	}

	lAnalytics, lErr := pHandler.store.GetSalesAnalytics(lStartDate, lEndDate)
	if lErr != nil {
		log.Printf("[ProductHandler][GetSalesAnalyticsHandler][ERROR] [STSA001] Failed to compute analytics: %v", lErr)
		pHandler.sendJSON(pWriter, http.StatusInternalServerError, nil, "[STSA001] Failed to compute sales analytics", false)
		return
	}

	log.Printf("[ProductHandler][GetSalesAnalyticsHandler] Analytics retrieved successfully for period '%s'", lAnalytics.Period)
	pHandler.sendJSON(pWriter, http.StatusOK, lAnalytics, "Sales analytics retrieved successfully", true)
}
```

---

### `Backend/cmd/server/main.go`
```go
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
	pRouter.HandleFunc("/analytics/sales", pHandler.GetSalesAnalyticsHandler).Methods(http.MethodGet)
	pRouter.HandleFunc("/getSalesAnalytics", pHandler.GetSalesAnalyticsHandler).Methods(http.MethodGet)
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
```

---

### `frontend/src/services/apiClient.js`
```javascript
import axios from 'axios';

const gBaseURL = 'http://localhost:8080/api';

const gApiClient = axios.create({
  baseURL: gBaseURL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
});

export default gApiClient;
```

---

### `frontend/src/services/productService.js`
```javascript
import gApiClient from './apiClient';

const gProductService = {
  fetchProducts() {
    return gApiClient.get('/getProducts');
  },
  createProduct(pData) {
    return gApiClient.post('/getProducts', pData);
  },
  updateStock(pId, pQuantity) {
    const lPayload = { stockQuantity: pQuantity };
    return gApiClient.patch(`/getProducts/${pId}`, lPayload);
  }
};

export default gProductService;
```

---

### `frontend/src/services/orderService.js`
```javascript
import gApiClient from './apiClient';

const gOrderService = {
  fetchOrders() {
    return gApiClient.get('/orders');
  },
  checkoutOrder(pPayload) {
    const lData = pPayload;
    return gApiClient.post('/orders', lData);
  },
  fetchSalesAnalytics(pParams) {
    return gApiClient.get('/analytics/sales', { params: pParams });
  }
};

export default gOrderService;
```

---

### `frontend/src/services/index.js`
```javascript
import gApiClient from './apiClient';
import gProductService from './productService';
import gOrderService from './orderService';

export const EventService = {
  // Domain modules
  products: gProductService,
  productService: gProductService,
  orders: gOrderService,
  orderService: gOrderService,

  // Direct convenience methods
  fetchProducts() {
    return gProductService.fetchProducts();
  },
  createProduct(pData) {
    return gProductService.createProduct(pData);
  },
  updateStock(pId, pQuantity) {
    return gProductService.updateStock(pId, pQuantity);
  },
  fetchOrders() {
    return gOrderService.fetchOrders();
  },
  checkoutOrder(pPayload) {
    return gOrderService.checkoutOrder(pPayload);
  },
  fetchSalesAnalytics(pParams) {
    return gOrderService.fetchSalesAnalytics(pParams);
  },
  checkHealth() {
    return gApiClient.get('/getHealth');
  },

  // Underlying configured API client
  apiClient: gApiClient
};

export { gApiClient as apiClient, gProductService as productService, gOrderService as orderService };
export default EventService;
```

---

### `frontend/src/views/SalesView.vue`
```vue
<template>
  <v-container fluid>
    <!-- Top Summary KPI Cards -->
    <v-row class="mb-4">
      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Total Revenue</div>
              <v-list-item-title class="text-h5 font-weight-bold success--text">
                ₹{{ totalRevenue.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="success lighten-4" size="52">
              <v-icon color="success" large>mdi-currency-inr</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Completed Orders</div>
              <v-list-item-title class="text-h5 font-weight-bold primary--text">
                {{ completedOrdersCount }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="primary lighten-4" size="52">
              <v-icon color="primary" large>mdi-receipt-text-check</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Average Order Value</div>
              <v-list-item-title class="text-h5 font-weight-bold info--text">
                ₹{{ averageOrderValue.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="info lighten-4" size="52">
              <v-icon color="info" large>mdi-chart-line</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="mx-auto" elevation="3" outlined>
          <v-list-item two-line>
            <v-list-item-content>
              <div class="text-overline mb-1">Units Sold</div>
              <v-list-item-title class="text-h5 font-weight-bold deep-purple--text">
                {{ totalUnitsSold }}
              </v-list-item-title>
            </v-list-item-content>
            <v-avatar color="deep-purple lighten-4" size="52">
              <v-icon color="deep-purple" large>mdi-package-variant-closed-check</v-icon>
            </v-avatar>
          </v-list-item>
        </v-card>
      </v-col>
    </v-row>

    <!-- Analytics Date Filter & Reporting Toolbar Card -->
    <v-card elevation="2" class="rounded-lg mb-4 pa-4">
      <v-row align="center" dense>
        <!-- Period Filter Button Toggle -->
        <v-col cols="12" md="8" class="d-flex flex-wrap align-center">
          <span class="caption font-weight-bold text-uppercase mr-3 text--secondary">
            <v-icon small left color="primary">mdi-calendar-range</v-icon>
            Period:
          </span>
          <v-btn-toggle
            v-model="selectedPeriod"
            mandatory
            dense
            color="primary"
            class="elevation-0"
            @change="onPeriodChange"
          >
            <v-btn small value="all" class="font-weight-medium text-capitalize">All Time</v-btn>
            <v-btn small value="today" class="font-weight-medium text-capitalize">Today</v-btn>
            <v-btn small value="last7" class="font-weight-medium text-capitalize">Last 7 Days</v-btn>
            <v-btn small value="last30" class="font-weight-medium text-capitalize">Last 30 Days</v-btn>
            <v-btn small value="custom" class="font-weight-medium text-capitalize">Custom Range</v-btn>
          </v-btn-toggle>
        </v-col>

        <!-- Actions: Refresh & CSV Export -->
        <v-col cols="12" md="4" class="d-flex justify-md-end align-center mt-2 mt-md-0">
          <v-btn
            color="success"
            outlined
            dense
            small
            class="font-weight-bold mr-2"
            @click="exportCSV"
          >
            <v-icon left small>mdi-file-delimited</v-icon>
            Export CSV
          </v-btn>
          <v-btn
            color="primary"
            class="font-weight-bold"
            small
            @click="openSalesDialog"
          >
            <v-icon left small>mdi-cart-plus</v-icon>
            Record Sale
          </v-btn>
        </v-col>
      </v-row>

      <!-- Custom Date Range Sub-Bar -->
      <v-expand-transition>
        <div v-if="selectedPeriod === 'custom'" class="pt-3 mt-2">
          <v-divider class="mb-3"></v-divider>
          <v-row dense align="center">
            <v-col cols="12" sm="4" md="3">
              <v-text-field
                v-model="startDate"
                label="Start Date"
                type="date"
                outlined
                dense
                hide-details
                prepend-inner-icon="mdi-calendar-start"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="4" md="3">
              <v-text-field
                v-model="endDate"
                label="End Date"
                type="date"
                outlined
                dense
                hide-details
                prepend-inner-icon="mdi-calendar-end"
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="4" md="3" class="d-flex align-center">
              <v-btn
                color="primary"
                small
                class="font-weight-bold mr-2"
                :loading="analyticsLoading"
                @click="applyDateFilter"
              >
                Apply Filter
              </v-btn>
              <v-btn
                text
                small
                @click="clearDateFilter"
              >
                Reset
              </v-btn>
            </v-col>
          </v-row>
        </div>
      </v-expand-transition>
    </v-card>

    <!-- Main Content: Sales History (Left) & Top Products Breakdown (Right) -->
    <v-row>
      <!-- Completed Sales History Table Column -->
      <v-col cols="12" lg="8">
        <v-card elevation="4" class="rounded-lg h-100">
          <v-card-title class="py-4 px-6">
            <v-icon left color="primary" medium>mdi-point-of-sale</v-icon>
            <span class="text-h6 font-weight-bold">Completed Sales History</span>
            <v-spacer></v-spacer>

            <!-- Search Input -->
            <v-text-field
              v-model="search"
              append-icon="mdi-magnify"
              label="Search orders or customers..."
              single-line
              hide-details
              outlined
              dense
              class="mr-2"
              style="max-width: 260px;"
            ></v-text-field>

            <!-- Refresh Button -->
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <v-btn
                  icon
                  color="primary"
                  v-bind="attrs"
                  v-on="on"
                  :loading="loading"
                  @click="refreshData"
                >
                  <v-icon>mdi-refresh</v-icon>
                </v-btn>
              </template>
              <span>Refresh Sales</span>
            </v-tooltip>
          </v-card-title>

          <v-divider></v-divider>

          <!-- Error Alert -->
          <v-alert
            v-if="errorMessage"
            type="error"
            dismissible
            dense
            class="mx-6 mt-4 mb-0"
            @input="errorMessage = ''"
          >
            {{ errorMessage }}
          </v-alert>

          <!-- Orders Data Table -->
          <v-data-table
            :headers="headers"
            :items="filteredOrders"
            :search="search"
            :loading="loading"
            loading-text="Fetching sales transactions via EventService..."
            no-data-text="No sales records available in selected period"
            class="elevation-0"
          >
            <!-- Order ID Slot -->
            <template v-slot:item.id="{ item }">
              <span class="font-weight-medium primary--text">#{{ item.id }}</span>
            </template>

            <!-- Total Amount Slot -->
            <template v-slot:item.totalAmount="{ item }">
              <span class="font-weight-bold success--text">
                ₹{{ Number(item.totalAmount).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </span>
            </template>

            <!-- Items Quantity Slot -->
            <template v-slot:item.itemsCount="{ item }">
              <span>{{ item.itemsCount }} item{{ item.itemsCount === 1 ? '' : 's' }}</span>
            </template>

            <!-- Date Slot -->
            <template v-slot:item.createdAt="{ item }">
              <span>{{ formatDate(item.createdAt) }}</span>
            </template>

            <!-- Status Chip Slot -->
            <template v-slot:item.status="{ item }">
              <v-chip
                small
                color="success lighten-1"
                class="white--text font-weight-bold"
              >
                <v-icon left x-small>mdi-check-circle</v-icon>
                {{ item.status || 'Completed' }}
              </v-chip>
            </template>

            <!-- Actions Slot -->
            <template v-slot:item.actions="{ item }">
              <v-btn
                icon
                small
                color="primary"
                @click="showOrderDetails(item)"
              >
                <v-icon small>mdi-receipt-text</v-icon>
              </v-btn>
            </template>
          </v-data-table>
        </v-card>
      </v-col>

      <!-- Top 5 Best-Selling Products Breakdown Widget -->
      <v-col cols="12" lg="4">
        <v-card elevation="4" class="rounded-lg h-100">
          <v-card-title class="py-4 px-6 d-flex align-center">
            <v-icon left color="amber darken-2" medium>mdi-podium-gold</v-icon>
            <span class="text-h6 font-weight-bold">Top 5 Products</span>
            <v-spacer></v-spacer>
            <v-chip x-small color="primary lighten-4" class="primary--text font-weight-bold">
              Volume
            </v-chip>
          </v-card-title>
          <v-divider></v-divider>

          <v-card-text class="pa-4">
            <div v-if="analyticsLoading" class="text-center py-8">
              <v-progress-circular indeterminate color="primary" size="32"></v-progress-circular>
              <div class="caption mt-2 text--secondary">Calculating sales rankings...</div>
            </div>
            <div v-else-if="topSellingProducts.length === 0" class="text-center py-8 text--secondary">
              <v-icon large color="grey lighten-1">mdi-package-variant</v-icon>
              <div class="mt-2 caption font-weight-medium">No sales recorded in this period</div>
            </div>
            <div v-else>
              <div
                v-for="(lProd, lIdx) in topSellingProducts"
                :key="lProd.product_id"
                class="mb-3 pa-3 rounded-lg"
                :class="$vuetify.theme.dark ? 'grey darken-3' : 'grey lighten-4'"
              >
                <div class="d-flex align-center justify-space-between mb-1">
                  <div class="d-flex align-center">
                    <v-chip
                      x-small
                      :color="lIdx === 0 ? 'amber darken-2' : lIdx === 1 ? 'blue-grey lighten-1' : lIdx === 2 ? 'brown lighten-1' : 'primary'"
                      class="white--text font-weight-bold mr-2"
                    >
                      #{{ lIdx + 1 }}
                    </v-chip>
                    <span class="font-weight-bold text-body-2 text-truncate" style="max-width: 170px;">
                      {{ lProd.name }}
                    </span>
                  </div>
                  <span class="font-weight-bold text-caption success--text">
                    ₹{{ Number(lProd.total_earned).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
                  </span>
                </div>
                <div class="d-flex align-center justify-space-between caption text--secondary mb-1">
                  <span>{{ lProd.units_sold }} unit{{ lProd.units_sold === 1 ? '' : 's' }} sold</span>
                  <span>{{ Math.round((lProd.units_sold / topProductMaxUnits) * 100) }}%</span>
                </div>
                <v-progress-linear
                  :value="(lProd.units_sold / topProductMaxUnits) * 100"
                  height="6"
                  rounded
                  :color="lIdx === 0 ? 'amber darken-2' : 'primary'"
                ></v-progress-linear>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Record Sale Modal Dialog -->
    <v-dialog v-model="lSalesDialog" max-width="720px" persistent>
      <v-card class="rounded-lg">
        <v-card-title class="primary white--text py-3 px-5">
          <v-icon left color="white">mdi-cash-register</v-icon>
          <span>Record New Sale</span>
          <v-spacer></v-spacer>
          <v-btn icon dark small @click="closeSalesDialog">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text class="pt-5 px-6">
          <!-- Structured Error Alert -->
          <v-alert
            v-if="formError"
            type="error"
            dense
            outlined
            dismissible
            class="mb-4 font-weight-medium"
            @input="formError = ''"
          >
            {{ formError }}
          </v-alert>

          <!-- Customer Name Input -->
          <v-row dense class="mb-2">
            <v-col cols="12" sm="8">
              <v-text-field
                v-model="lCustomerName"
                label="Customer Name"
                placeholder="Walk-in Customer"
                outlined
                dense
                prepend-inner-icon="mdi-account"
              ></v-text-field>
            </v-col>
          </v-row>

          <div class="text-subtitle-2 font-weight-bold mb-2 primary--text">
            Line Items
          </div>

          <!-- Line Items List -->
          <v-card
            outlined
            class="pa-3 mb-3"
            v-for="(lItem, lIndex) in lLineItems"
            :key="lIndex"
          >
            <v-row dense align="center">
              <!-- Product Autocomplete -->
              <v-col cols="12" sm="5">
                <v-autocomplete
                  v-model="lItem.product"
                  :items="availableProducts"
                  item-text="name"
                  return-object
                  label="Select Product *"
                  placeholder="Type to search..."
                  outlined
                  dense
                  hide-details="auto"
                  prepend-inner-icon="mdi-cube-outline"
                  @change="onProductChange(lItem)"
                >
                  <template v-slot:item="{ item }">
                    <v-list-item-content>
                      <v-list-item-title class="font-weight-medium">{{ item.name }}</v-list-item-title>
                      <v-list-item-subtitle class="caption">
                        SKU: {{ item.sku }} | Price: ₹{{ Number(item.price).toFixed(2) }}
                      </v-list-item-subtitle>
                    </v-list-item-content>
                    <v-list-item-action>
                      <v-chip
                        x-small
                        :color="item.stockQuantity > 0 ? 'success' : 'error'"
                        outlined
                        class="font-weight-bold"
                      >
                        {{ item.stockQuantity }} in stock
                      </v-chip>
                    </v-list-item-action>
                  </template>
                </v-autocomplete>
              </v-col>

              <!-- Unit Price (read-only indicator) -->
              <v-col cols="6" sm="2">
                <v-text-field
                  :value="lItem.unitPrice ? '₹' + Number(lItem.unitPrice).toFixed(2) : '-'"
                  label="Unit Price"
                  outlined
                  dense
                  readonly
                  hide-details
                ></v-text-field>
              </v-col>

              <!-- Quantity Input with Stock Validation -->
              <v-col cols="6" sm="2">
                <v-text-field
                  v-model.number="lItem.quantity"
                  type="number"
                  min="1"
                  :max="lItem.maxStock"
                  label="Qty *"
                  outlined
                  dense
                  hide-details="auto"
                  :rules="[
                    v => !!v && v > 0 || '[SVSO001] Min 1',
                    v => !lItem.product || v <= lItem.maxStock || `[SVSO002] Max ${lItem.maxStock}`
                  ]"
                ></v-text-field>
              </v-col>

              <!-- Line Item Total -->
              <v-col cols="8" sm="2" class="text-right">
                <div class="caption text--secondary">Line Total</div>
                <div class="font-weight-bold success--text">
                  ₹{{ ((lItem.quantity || 0) * (lItem.unitPrice || 0)).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
                </div>
              </v-col>

              <!-- Remove Item Button -->
              <v-col cols="4" sm="1" class="text-right">
                <v-btn
                  icon
                  small
                  color="error"
                  :disabled="lLineItems.length <= 1"
                  @click="removeLineItem(lIndex)"
                >
                  <v-icon small>mdi-delete-outline</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </v-card>

          <!-- Add Line Item Button -->
          <v-btn
            text
            small
            color="primary"
            class="font-weight-bold px-0 mb-3"
            @click="addLineItem"
          >
            <v-icon left small>mdi-plus</v-icon>
            + Add Another Item
          </v-btn>

          <v-divider class="my-3"></v-divider>

          <!-- Grand Total Summary in Footer -->
          <v-row align="center" justify="end" class="px-2">
            <div class="text-h6 font-weight-medium mr-4">Total Amount:</div>
            <div class="text-h5 font-weight-bold success--text">
              ₹{{ lDialogGrandTotal.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
            </div>
          </v-row>
        </v-card-text>

        <v-card-actions class="pb-4 px-6">
          <v-spacer></v-spacer>
          <v-btn text @click="closeSalesDialog" :disabled="submitLoading">
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            class="font-weight-bold px-6"
            :loading="submitLoading"
            @click="submitSale"
          >
            <v-icon left>mdi-check-circle-outline</v-icon>
            Complete Sale
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Order Receipt Dialog -->
    <v-dialog v-model="isDetailsDialogVisible" max-width="480px">
      <v-card v-if="selectedOrder" class="rounded-lg">
        <v-card-title class="primary white--text py-3 px-5">
          <v-icon left color="white">mdi-receipt</v-icon>
          <span>Order Details: #{{ selectedOrder.id }}</span>
          <v-spacer></v-spacer>
          <v-btn icon dark small @click="isDetailsDialogVisible = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text class="pt-4 px-6">
          <v-list dense>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Customer Name</v-list-item-subtitle>
                <v-list-item-title class="font-weight-bold">{{ selectedOrder.customer }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Transaction Date</v-list-item-subtitle>
                <v-list-item-title>{{ formatDate(selectedOrder.createdAt) }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Total Items</v-list-item-subtitle>
                <v-list-item-title class="font-weight-medium">{{ selectedOrder.itemsCount }} units</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Total Amount</v-list-item-subtitle>
                <v-list-item-title class="text-h6 font-weight-bold success--text">
                  ₹{{ Number(selectedOrder.totalAmount).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-content>
                <v-list-item-subtitle>Payment Status</v-list-item-subtitle>
                <v-list-item-title>
                  <v-chip small color="success" text-color="white">{{ selectedOrder.status || 'Completed' }}</v-chip>
                </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions class="pb-3 px-6">
          <v-spacer></v-spacer>
          <v-btn text color="primary" @click="isDetailsDialogVisible = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Feedback Snackbar -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      timeout="4000"
      bottom
      right
    >
      {{ snackbar.text }}
      <template v-slot:action="{ attrs }">
        <v-btn text v-bind="attrs" @click="snackbar.show = false">
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-container>
</template>

<script>
import EventService from '../services';

export default {
  name: 'SalesView',
  data() {
    return {
      search: '',
      loading: false,
      analyticsLoading: false,
      serverAnalytics: null,
      selectedPeriod: 'all',
      startDate: '',
      endDate: '',
      orders: [],
      availableProducts: [],
      selectedOrder: null,
      isDetailsDialogVisible: false,
      lSalesDialog: false,
      submitLoading: false,
      formError: '',
      errorMessage: '',
      lCustomerName: 'Walk-in Customer',
      lLineItems: [
        { product: null, quantity: 1, maxStock: 0, unitPrice: 0 }
      ],
      headers: [
        { text: 'Order ID', value: 'id', align: 'start', sortable: true },
        { text: 'Customer', value: 'customer', align: 'start', sortable: true },
        { text: 'Date', value: 'createdAt', align: 'start', sortable: true },
        { text: 'Items', value: 'itemsCount', align: 'right', sortable: true },
        { text: 'Total Amount (₹)', value: 'totalAmount', align: 'right', sortable: true },
        { text: 'Status', value: 'status', align: 'center', sortable: false },
        { text: 'Actions', value: 'actions', align: 'center', sortable: false }
      ],
      snackbar: {
        show: false,
        text: '',
        color: 'success'
      }
    };
  },
  computed: {
    filteredOrders() {
      if (!this.selectedPeriod || this.selectedPeriod === 'all') {
        return this.orders;
      }
      let lStart = null;
      let lEnd = null;
      const lNow = new Date();

      if (this.selectedPeriod === 'today') {
        lStart = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate(), 0, 0, 0);
        lEnd = new Date(lNow.getFullYear(), lNow.getMonth(), lNow.getDate(), 23, 59, 59, 999);
      } else if (this.selectedPeriod === 'last7') {
        lStart = new Date(Date.now() - 7 * 24 * 3600 * 1000);
        lEnd = lNow;
      } else if (this.selectedPeriod === 'last30') {
        lStart = new Date(Date.now() - 30 * 24 * 3600 * 1000);
        lEnd = lNow;
      } else if (this.selectedPeriod === 'custom') {
        if (this.startDate) {
          lStart = new Date(this.startDate + 'T00:00:00');
        }
        if (this.endDate) {
          lEnd = new Date(this.endDate + 'T23:59:59.999');
        }
      }

      return this.orders.filter(pOrder => {
        if (!pOrder.createdAt) return true;
        const lOrderDate = new Date(pOrder.createdAt);
        if (isNaN(lOrderDate.getTime())) return true;
        if (lStart && lOrderDate < lStart) return false;
        if (lEnd && lOrderDate > lEnd) return false;
        return true;
      });
    },
    totalRevenue() {
      if (this.serverAnalytics && typeof this.serverAnalytics.total_revenue === 'number') {
        return this.serverAnalytics.total_revenue;
      }
      return this.filteredOrders.reduce((lSum, lOrder) => lSum + (Number(lOrder.totalAmount) || 0), 0);
    },
    completedOrdersCount() {
      if (this.serverAnalytics && typeof this.serverAnalytics.order_count === 'number') {
        return this.serverAnalytics.order_count;
      }
      return this.filteredOrders.length;
    },
    averageOrderValue() {
      if (this.serverAnalytics && typeof this.serverAnalytics.average_order_value === 'number') {
        return this.serverAnalytics.average_order_value;
      }
      if (this.completedOrdersCount === 0) return 0;
      return this.totalRevenue / this.completedOrdersCount;
    },
    totalUnitsSold() {
      if (this.serverAnalytics && typeof this.serverAnalytics.total_units_sold === 'number') {
        return this.serverAnalytics.total_units_sold;
      }
      return this.filteredOrders.reduce((lSum, lOrder) => lSum + (Number(lOrder.itemsCount) || 0), 0);
    },
    topSellingProducts() {
      if (this.serverAnalytics && Array.isArray(this.serverAnalytics.top_products)) {
        return this.serverAnalytics.top_products;
      }
      // Fallback client-side calculation from filteredOrders
      const lMap = {};
      for (const lOrder of this.filteredOrders) {
        if (Array.isArray(lOrder.items)) {
          for (const lItem of lOrder.items) {
            if (!lMap[lItem.product_id]) {
              const lProd = this.availableProducts.find(p => p.id === lItem.product_id);
              lMap[lItem.product_id] = {
                product_id: lItem.product_id,
                name: lProd ? lProd.name : lItem.product_id,
                units_sold: 0,
                total_earned: 0
              };
            }
            lMap[lItem.product_id].units_sold += Number(lItem.quantity) || 0;
            lMap[lItem.product_id].total_earned += (Number(lItem.quantity) || 0) * (Number(lItem.unit_price) || 0);
          }
        }
      }
      return Object.values(lMap)
        .sort((pA, pB) => pB.units_sold - pA.units_sold)
        .slice(0, 5);
    },
    topProductMaxUnits() {
      if (!this.topSellingProducts || this.topSellingProducts.length === 0) return 1;
      const lMax = Math.max(...this.topSellingProducts.map(p => p.units_sold));
      return lMax > 0 ? lMax : 1;
    },
    lDialogGrandTotal() {
      return this.lLineItems.reduce((lSum, lItem) => {
        const lQty = Number(lItem.quantity) || 0;
        const lPrice = Number(lItem.unitPrice) || 0;
        return lSum + (lQty * lPrice);
      }, 0);
    }
  },
  mounted() {
    this.refreshData();
    this.fetchProducts();
  },
  methods: {
    async refreshData() {
      await Promise.all([
        this.fetchOrders(),
        this.fetchAnalytics()
      ]);
    },
    async fetchOrders() {
      this.loading = true;
      this.errorMessage = '';
      try {
        const lResponse = await EventService.orders.fetchOrders();
        if (lResponse.data && lResponse.data.success) {
          this.orders = lResponse.data.data || [];
          this.$emit('api-status-change', true);
        } else if (Array.isArray(lResponse.data)) {
          this.orders = lResponse.data;
          this.$emit('api-status-change', true);
        } else {
          const lMsg = 'Failed to load orders: ' + (lResponse.data.message || 'Unknown error');
          this.orders = [];
          this.errorMessage = lMsg;
          this.showSnackbar(lMsg, 'error');
        }
      } catch (lError) {
        const lMsg = 'Failed to fetch sales transactions: ' + (lError.response?.data?.message || lError.message || 'Network error');
        this.orders = [];
        this.errorMessage = lMsg;
        this.showSnackbar(lMsg, 'error');
        this.$emit('api-status-change', false);
      } finally {
        this.loading = false;
      }
    },
    async fetchAnalytics() {
      this.analyticsLoading = true;
      try {
        const lParams = {};
        if (this.selectedPeriod === 'today') {
          const lNow = new Date();
          const lTodayStr = lNow.toISOString().slice(0, 10);
          lParams.startDate = lTodayStr;
          lParams.endDate = lTodayStr;
        } else if (this.selectedPeriod === 'last7') {
          const lPast = new Date(Date.now() - 7 * 24 * 3600 * 1000);
          lParams.startDate = lPast.toISOString().slice(0, 10);
          lParams.endDate = new Date().toISOString().slice(0, 10);
        } else if (this.selectedPeriod === 'last30') {
          const lPast = new Date(Date.now() - 30 * 24 * 3600 * 1000);
          lParams.startDate = lPast.toISOString().slice(0, 10);
          lParams.endDate = new Date().toISOString().slice(0, 10);
        } else if (this.selectedPeriod === 'custom') {
          if (this.startDate) lParams.startDate = this.startDate;
          if (this.endDate) lParams.endDate = this.endDate;
        }

        const lResponse = await EventService.orders.fetchSalesAnalytics(lParams);
        if (lResponse.data && lResponse.data.success && lResponse.data.data) {
          this.serverAnalytics = lResponse.data.data;
        }
      } catch (lError) {
        this.serverAnalytics = null;
      } finally {
        this.analyticsLoading = false;
      }
    },
    onPeriodChange(pPeriod) {
      if (pPeriod !== 'custom') {
        this.startDate = '';
        this.endDate = '';
        this.fetchAnalytics();
      }
    },
    applyDateFilter() {
      if (this.startDate && this.endDate && this.startDate > this.endDate) {
        this.showSnackbar('[HNSA002] Start date cannot be after end date.', 'error');
        return;
      }
      this.fetchAnalytics();
    },
    clearDateFilter() {
      this.selectedPeriod = 'all';
      this.startDate = '';
      this.endDate = '';
      this.fetchAnalytics();
    },
    exportCSV() {
      try {
        if (!this.filteredOrders || this.filteredOrders.length === 0) {
          this.showSnackbar('No sales records available to export.', 'warning');
          return;
        }

        const lHeaders = ['Order ID', 'Customer', 'Date', 'Items Count', 'Total Amount (₹)', 'Status'];
        const lRows = this.filteredOrders.map(pOrder => [
          `"${pOrder.id}"`,
          `"${(pOrder.customer || '').replace(/"/g, '""')}"`,
          `"${pOrder.createdAt || ''}"`,
          pOrder.itemsCount || 0,
          (Number(pOrder.totalAmount) || 0).toFixed(2),
          `"${pOrder.status || 'Completed'}"`
        ]);

        const lCsvContent = [lHeaders.join(','), ...lRows.map(pRow => pRow.join(','))].join('\r\n');
        const lBlob = new Blob([lCsvContent], { type: 'text/csv;charset=utf-8;' });
        const lUrl = URL.createObjectURL(lBlob);
        const lLink = document.createElement('a');
        lLink.setAttribute('href', lUrl);
        lLink.setAttribute('download', `shoptrack_sales_${new Date().toISOString().slice(0, 10)}.csv`);
        document.body.appendChild(lLink);
        lLink.click();
        document.body.removeChild(lLink);
        URL.revokeObjectURL(lUrl);

        this.showSnackbar('Sales report exported to CSV successfully!', 'success');
      } catch (lError) {
        this.showSnackbar('[SVSA001] Failed to export sales CSV: ' + (lError.message || 'Unknown error'), 'error');
      }
    },
    async fetchProducts() {
      try {
        const lResponse = await EventService.products.fetchProducts();
        if (lResponse.data && lResponse.data.success) {
          this.availableProducts = lResponse.data.data || [];
        } else if (Array.isArray(lResponse.data)) {
          this.availableProducts = lResponse.data;
        } else {
          this.availableProducts = [];
        }
      } catch (lError) {
        this.availableProducts = [];
      }
    },
    openSalesDialog() {
      this.formError = '';
      this.lCustomerName = 'Walk-in Customer';
      this.lLineItems = [
        { product: null, quantity: 1, maxStock: 0, unitPrice: 0 }
      ];
      this.fetchProducts();
      this.lSalesDialog = true;
    },
    closeSalesDialog() {
      this.lSalesDialog = false;
      this.formError = '';
    },
    onProductChange(pItem) {
      if (pItem.product) {
        pItem.unitPrice = Number(pItem.product.price) || 0;
        pItem.maxStock = Number(pItem.product.stockQuantity) || 0;
        if (pItem.quantity > pItem.maxStock && pItem.maxStock > 0) {
          pItem.quantity = pItem.maxStock;
        }
      } else {
        pItem.unitPrice = 0;
        pItem.maxStock = 0;
      }
    },
    addLineItem() {
      this.lLineItems.push({
        product: null,
        quantity: 1,
        maxStock: 0,
        unitPrice: 0
      });
    },
    removeLineItem(pIndex) {
      if (this.lLineItems.length > 1) {
        this.lLineItems.splice(pIndex, 1);
      }
    },
    async submitSale() {
      this.formError = '';

      // Validation 1: Empty line items or missing products
      if (!this.lLineItems || this.lLineItems.length === 0) {
        this.formError = '[SVSO001] Order must contain at least one item.';
        return;
      }

      for (let lIndex = 0; lIndex < this.lLineItems.length; lIndex++) {
        const lItem = this.lLineItems[lIndex];
        if (!lItem.product || !lItem.product.id) {
          this.formError = '[SVSO001] Please select a product for all line items.';
          return;
        }
        const lQty = Number(lItem.quantity);
        if (!lQty || lQty <= 0) {
          this.formError = `[SVSO001] Quantity for item #${lIndex + 1} must be at least 1.`;
          return;
        }
        if (lQty > lItem.maxStock) {
          this.formError = `[SVSO002] Quantity for '${lItem.product.name}' (${lQty}) exceeds available stock (${lItem.maxStock}).`;
          return;
        }
      }

      this.submitLoading = true;
      try {
        const lOrderPayload = {
          customer: (this.lCustomerName || 'Walk-in Customer').trim(),
          items: this.lLineItems.map(pItem => ({
            product_id: pItem.product.id,
            quantity: Number(pItem.quantity),
            unit_price: Number(pItem.unitPrice)
          }))
        };

        const lResponse = await EventService.orders.checkoutOrder(lOrderPayload);
        if (lResponse.data && lResponse.data.success) {
          const lCreatedOrder = lResponse.data.data;
          this.showSnackbar(`Sale recorded successfully! Total: ₹${Number(lCreatedOrder.totalAmount).toFixed(2)}`, 'success');
          this.closeSalesDialog();
          this.refreshData();
          this.fetchProducts();
        } else {
          this.formError = lResponse.data.message || '[HNCO004] Failed to record sale.';
        }
      } catch (lError) {
        if (lError.response && lError.response.data && lError.response.data.message) {
          this.formError = lError.response.data.message;
        } else {
          this.formError = '[HNCO001] Network error or backend unavailable.';
        }
      } finally {
        this.submitLoading = false;
      }
    },
    formatDate(pDateStr) {
      if (!pDateStr) return '-';
      const lDate = new Date(pDateStr);
      if (isNaN(lDate.getTime())) return pDateStr;
      return lDate.toLocaleDateString('en-IN', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    },
    showOrderDetails(pOrder) {
      this.selectedOrder = pOrder;
      this.isDetailsDialogVisible = true;
    },
    showSnackbar(pText, pColor = 'success') {
      this.snackbar.text = pText;
      this.snackbar.color = pColor;
      this.snackbar.show = true;
    }
  }
};
</script>

<style scoped>
</style>
```
