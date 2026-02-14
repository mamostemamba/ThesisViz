.PHONY: infra infra-down api render render-setup web dev

# Proxy for accessing Google API (comment out if not needed)
export https_proxy := http://127.0.0.1:7897
export http_proxy  := http://127.0.0.1:7897
export all_proxy   := socks5://127.0.0.1:7897

# Start Docker infrastructure (PostgreSQL, Redis, MinIO)
infra:
	docker compose up -d

# Stop Docker infrastructure
infra-down:
	docker compose down

# Start Go API server (port 8080)
api:
	cd go-api && go run ./cmd/server

# Create Python venv and install dependencies (run once)
render-setup:
	cd py-render && python3.13 -m venv .venv && .venv/bin/pip install -r requirements.txt

# Start Python render sidecar (port 8081)
render:
	cd py-render && .venv/bin/uvicorn main:app --host 0.0.0.0 --port 8081 --reload

# Start Next.js frontend (port 3000)
web:
	cd web && npm run dev

# Start all services (infrastructure + all apps)
dev:
	@echo "Starting infrastructure..."
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Starting Go API, Python Render, and Next.js..."
	@trap 'kill 0' EXIT; \
		(cd go-api && go run ./cmd/server) & \
		(cd py-render && .venv/bin/uvicorn main:app --host 0.0.0.0 --port 8081 --reload) & \
		(cd web && npm run dev) & \
		wait
