# ShopTrack - Store & Inventory Management System

A full-stack store and inventory management service built with Go and Vue 2 / Vuetify, designed to demonstrate modular architecture, clean concurrency patterns, and automated code review workflows.

## Tech Stack
- **Backend**: Go (Gorilla Mux, In-Memory Thread-Safe Store with `sync.RWMutex`)
- **Frontend**: Vue 2, Vuetify 2, Axios

## Architecture Overview
- `backend/`: Go REST API exposing health check, product catalog, and order processing endpoints.
- `frontend/`: Single-page dashboard displaying live inventory status, low-stock alerts, and item registration modals.

## Getting Started

### Backend Setup
```bash
cd backend
go run cmd/server/main.go
# Server starts on http://localhost:8080