.PHONY: run build test swagger migrate docker-up docker-down lint tidy clean mock

APP_NAME := ops-server
BINARY   := ./bin/$(APP_NAME)
MAIN     := ./cmd/main.go

# ── Build ──────────────────────────────────────────────────────────────────────
build:
	@echo ">> Building $(APP_NAME)..."
	@mkdir -p bin
	go build -ldflags="-w -s" -o $(BINARY) $(MAIN)

# ── Run ────────────────────────────────────────────────────────────────────────
run:
	go run $(MAIN)

# ── Tests ──────────────────────────────────────────────────────────────────────
test:
	go test ./... -v -cover -race

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo ">> Report: coverage.html"

# ── Swagger ────────────────────────────────────────────────────────────────────
swagger:
	@which swag > /dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	swag init \
		-g docs/swagger.go \
		--dir .,./internal/domain/user/controller,./internal/domain/notification/controller,./internal/domain/metrics/controller,./internal/domain/audit/controller \
		--output ./docs \
		--parseDependency \
		--parseInternal \
		--generatedTime
	@echo "✅  Swagger: http://localhost:8080/swagger/index.html"

# ── Lint ───────────────────────────────────────────────────────────────────────
lint:
	@which golangci-lint > /dev/null 2>&1 || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	golangci-lint run ./...

# ── Mocks ──────────────────────────────────────────────────────────────────────
mock:
	@which mockgen > /dev/null 2>&1 || go install github.com/golang/mock/mockgen@latest
	mockgen -source=internal/domain/user/repository/user_repository.go \
		-destination=internal/domain/user/repository/mocks/user_repository_mock.go -package=mocks
	mockgen -source=internal/domain/notification/repository/notification_repository.go \
		-destination=internal/domain/notification/repository/mocks/notification_repository_mock.go -package=mocks
	mockgen -source=internal/domain/audit/repository/audit_repository.go \
		-destination=internal/domain/audit/repository/mocks/audit_repository_mock.go -package=mocks
	mockgen -source=internal/domain/metrics/repository/metrics_repository.go \
		-destination=internal/domain/metrics/repository/mocks/metrics_repository_mock.go -package=mocks

# ── Docker ─────────────────────────────────────────────────────────────────────
docker-up:
	docker compose -f deployments/docker-compose.yml up -d
	@echo ">> Kafka UI: http://localhost:8090"

docker-down:
	docker compose -f deployments/docker-compose.yml down

docker-logs:
	docker compose -f deployments/docker-compose.yml logs -f

docker-build:
	docker build -f deployments/Dockerfile -t $(APP_NAME):latest .

# ── Migrations ─────────────────────────────────────────────────────────────────
migrate-up:
	bash scripts/migrations/migrate.sh up

migrate-down:
	bash scripts/migrations/migrate.sh down

# ── Tidy ───────────────────────────────────────────────────────────────────────
tidy:
	go mod tidy

# ── Clean ──────────────────────────────────────────────────────────────────────
clean:
	rm -rf bin/ coverage.out coverage.html

# ── Dev bootstrap (tout en un) ──────────────────────────────────────────────
dev: docker-up tidy swagger
	@echo ">> Stack prête. Lancement de l'API..."
	go run $(MAIN)
