# ─────────────────────────────────────────────────────────────────────────────
# Wealth Management – Makefile
#
# Targets
#   Docker (full stack via docker-compose)
#     make up          – build images and start all services in the background
#     make down        – stop and remove containers (volumes preserved)
#     make restart     – down + up
#     make build       – (re)build Docker images without starting containers
#     make logs        – tail logs for all services
#     make ps          – show running service status
#     make clean       – stop containers AND remove the postgres data volume
#
#   Local development (without Docker)
#     make backend     – build the Go binary to ./backend/wealth-management-backend
#     make backend-run – build and run the backend (reads DB_* from environment)
#     make frontend    – install npm dependencies and build the React app
#     make frontend-dev – start the Vite dev server (hot-reload on :5173)
#
#   Utilities
#     make help        – print this help
# ─────────────────────────────────────────────────────────────────────────────

# Load .env when it exists so variables are available to make targets that call
# docker-compose directly (docker-compose itself also reads .env automatically).
-include .env
export

COMPOSE         := docker compose
BACKEND_DIR     := backend
FRONTEND_DIR    := frontend
BINARY          := $(BACKEND_DIR)/wealth-management-backend

.DEFAULT_GOAL := help

# ── Docker targets ────────────────────────────────────────────────────────────

.PHONY: up
up: ## Build images (if needed) and start all services in the background
	$(COMPOSE) up --build -d

.PHONY: down
down: ## Stop and remove containers (data volume is preserved)
	$(COMPOSE) down

.PHONY: restart
restart: down up ## Stop, rebuild, and restart all services

.PHONY: build
build: ## Build (or rebuild) all Docker images
	$(COMPOSE) build

.PHONY: logs
logs: ## Tail logs for all services (Ctrl-C to stop)
	$(COMPOSE) logs -f

.PHONY: ps
ps: ## Show status of all services
	$(COMPOSE) ps

.PHONY: clean
clean: ## Stop containers and DELETE the postgres data volume (destructive!)
	$(COMPOSE) down -v

# ── Local development targets ─────────────────────────────────────────────────

.PHONY: backend
backend: ## Compile the Go backend binary (output: backend/wealth-management-backend)
	cd $(BACKEND_DIR) && go build -o wealth-management-backend .

.PHONY: backend-run
backend-run: backend ## Build and run the backend locally (uses DB_* env vars)
	cd $(BACKEND_DIR) && ./wealth-management-backend

.PHONY: frontend
frontend: ## Install npm dependencies and build the React app for production
	cd $(FRONTEND_DIR) && npm install && npm run build

.PHONY: frontend-dev
frontend-dev: ## Start the Vite development server with hot-reload on :5173
	cd $(FRONTEND_DIR) && npm install && npm run start

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Print available targets and their descriptions
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
