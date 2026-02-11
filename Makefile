.PHONY: infra infra-down api render web dev

# Start Docker infrastructure (PostgreSQL, Redis, MinIO)
infra:
	docker compose up -d

# Stop Docker infrastructure
infra-down:
	docker compose down

# Start Go API server (port 8080)
api:
	cd go-api && go run ./cmd/server

# Start Python render sidecar (port 8081)
render:
	cd py-render && python3.12 -m uvicorn main:app --host 0.0.0.0 --port 8081 --reload

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
		(cd py-render && python3.12 -m uvicorn main:app --host 0.0.0.0 --port 8081 --reload) & \
		(cd web && npm run dev) & \
		wait
