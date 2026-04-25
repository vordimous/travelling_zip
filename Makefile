.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: run-backend
run-backend: ## Run the Go API server (port 3001)
	cd backend && go run ./cmd/api

.PHONY: run-simulator
run-simulator: ## Run the Go CLI simulator
	cd backend && go run ./cmd/simulator

.PHONY: build-backend
build-backend: ## Compile Go binary → backend/bin/api
	cd backend && go build -o bin/api ./cmd/api

.PHONY: vet-backend
vet-backend: ## Run go vet on backend
	cd backend && go vet ./...

.PHONY: test-backend
test-backend: ## Run Go tests
	cd backend && go test ./...

.PHONY: test-frontend
test-frontend: ## Run frontend tests (Vitest)
	cd frontend && npm test

.PHONY: install-frontend
install-frontend: ## npm install in frontend/ (dev)
	cd frontend && npm install

.PHONY: ci-install-frontend
ci-install-frontend: ## npm ci in frontend/ (CI — clean reproducible install)
	cd frontend && npm ci

.PHONY: run-frontend
run-frontend: ## Start Vite dev server (port 5173)
	cd frontend && npm run dev

.PHONY: build-frontend
build-frontend: ## Vite production build → frontend/dist/
	cd frontend && npm run build

.PHONY: build
build: build-backend build-frontend ## Build both backend and frontend

.PHONY: test
test: test-backend test-frontend ## Run all tests
