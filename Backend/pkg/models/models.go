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

