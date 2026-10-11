# ShopTrack - Store & Inventory Management System

A full-stack store and inventory management service built with Go (Gorilla Mux) and Vue 2 / Vuetify, designed with modular service layers, clean concurrency patterns, robust error handling, and structured startup logging.

## Tech Stack
- **Backend**: Go (Gorilla Mux, In-Memory Thread-Safe Store with `sync.RWMutex`)
- **Frontend**: Vue 2, Vuetify 2, Axios, Vue Router

---

## Key Architectural Additions

### 1. Robust Middleware & CORS Preflight
- **CORS Preflight Middleware (`pkg/middleware/cors.go`)**:
  - Implements standalone `CORSMiddleware(pNext http.Handler) http.Handler`.
  - Injects headers:
    - `Access-Control-Allow-Origin: *`
    - `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
    - `Access-Control-Allow-Headers: Content-Type, Authorization`
  - Intercepts `OPTIONS` preflight requests and returns HTTP `200 OK` immediately without passing down to unmatched routes.
- **Panic Recovery Middleware (`pkg/middleware/recovery.go`)**:
  - Implements standalone `RecoveryMiddleware(pNext http.Handler) http.Handler`.
  - Captures runtime panics with `defer ... recover()`, logs the stack trace using standard `log.Printf`, and responds with HTTP `500` JSON: `{"error": "internal server error"}` to prevent process termination.

### 2. Microsecond-Timestamped Startup Logging
- **Logger Package (`pkg/logger/logger.go`)**:
  - Generates a fresh log file on every server startup under `logs/app_YYYY-MM-DD_HH-mm-ss.000000.log` (microsecond layout: `2006-01-02_15-04-05.000000`).
  - Uses `io.MultiWriter(os.Stdout, lFile)` for simultaneous console and file logging.
  - Sets `log.SetFlags(log.Ldate | log.Ltime)` for second-precision timestamps across log entries.
  - Provides `CloseLogger()` for flushing and clean file descriptor release upon shutdown.

### 3. Modular Frontend `EventService` Layer
- **Centralized API Client (`frontend/src/services/apiClient.js`)**:
  - Preconfigured Axios instance targeting `http://localhost:8080/api`.
- **Domain Services**:
  - `productService.js`: Encapsulates `fetchProducts()`, `createProduct(pData)`, and `updateStock(pId, pQuantity)`.
  - `orderService.js`: Encapsulates `fetchOrders()` and `checkoutOrder(pPayload)`.
- **Unified Export (`frontend/src/services/index.js`)**:
  - Exposes `EventService` consolidating domain services, top-level convenience methods, and health checking so components import from a single location.
  - All views and dialogs (`ProductsView.vue`, `ProductDialog.vue`, `SalesView.vue`, `App.vue`) route their HTTP communication through `EventService`.

### 4. Currency Localization (`₹`)
- All price representations, table column headers, summary cards, and product creation dialog input indicators are standardized to Indian Rupees (`₹`).

### 5. Sales & Orders Module (`/sales`)
- **Sales View (`frontend/src/views/SalesView.vue`)**:
  - Executive KPI cards for Total Revenue (₹), Completed Orders, Average Order Value (AOV in ₹), and Total Units Sold.
  - Interactive data table listing completed orders with item details, formatted dates, status badges, and receipt detail modal.
  - Registered route at `/sales` with navigation buttons directly in the top application bar.

### 6. Sales Reporting Analytics & Export
- **Server-Side Analytics Engine (`/api/analytics/sales`)**:
  - Aggregates total revenue, order volume, average order value, and total units sold.
  - Computes top 5 best-selling products by units sold and total earnings.
  - Supports ISO 8601 / RFC3339 and `YYYY-MM-DD` date range filters (`startDate`, `endDate`) with validation error codes (`[HNSA001]`, `[HNSA002]`).
- **Client-Side Reporting Enhancements**:
  - Date filter toolbar with quick presets (All Time, Today, Last 7 Days, Last 30 Days, Custom Range).
  - Top 5 Best-Selling Products widget with ranking chips (#1 to #5) and volume progress bars.
  - Synchronized real-time KPI metrics.
  - Client-side CSV transaction export with RFC-compliant escaping and structured error handling (`[SVSA001]`).

---

## API Endpoints Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/getHealth` | Health check endpoint |
| `GET` | `/api/getProducts` | Retrieve all inventory products |
| `GET` | `/api/getProducts/{id}` | Retrieve specific product by ID |
| `POST` | `/api/getProducts` | Create a new inventory product |
| `GET` | `/api/orders` | Retrieve completed sales orders |
| `POST` | `/api/orders` | Checkout and record a new sales order |
| `GET` | `/api/analytics/sales` | Aggregated sales metrics & top products with optional date range |
| `OPTIONS` | Any route | CORS preflight returning HTTP 200 OK |
| `GET` | `/panic` | Panic simulation route for recovery verification |

*(Note: Routes are also mirrored on `/api/v1/*` for backward compatibility).*

---

## Getting Started

### Backend Setup
```bash
cd Backend
go test -v ./...
go run cmd/server/main.go
# Server starts on http://localhost:8080 and writes logs to logs/app_*.log
```

### Frontend Setup
```bash
cd frontend
npm install
npm run serve
# Dashboard starts on http://localhost:8081
```