package models

// Product represents an inventory item in the store.
type Product struct {
	ID            string  `json:"id"`
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	StockQuantity int     `json:"stockQuantity"`
}

// StandardResponse defines a unified API response structure.
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}
