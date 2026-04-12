#!/usr/bin/env bash

set -e

APP_NAME="ops-server"
BINARY="./bin/$APP_NAME"
MAIN="./cmd/main.go"

# ─────────────────────────────────────────────
# Utils
# ─────────────────────────────────────────────
log() {
  echo -e "\033[1;34m>> $1\033[0m"
}

ensure_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    log "$1 not found, installing..."
    eval "$2"
  }
}

# ─────────────────────────────────────────────
# Commands
# ─────────────────────────────────────────────

build() {
  log "Building $APP_NAME..."
  mkdir -p bin
  go build -ldflags="-w -s" -o "$BINARY" "$MAIN"
}

run() {
  go run "$MAIN"
}

test() {
  go test ./... -v -cover -race
}

test_coverage() {
  go test ./... -coverprofile=coverage.out
  go tool cover -html=coverage.out -o coverage.html
  log "Report: coverage.html"
}

swagger() {
  ensure_cmd "swag" "go install github.com/swaggo/swag/cmd/swag@latest"

  swag init \
    -g docs/swagger.go \
    --dir .,./internal/domain/user/controller,./internal/domain/notification/controller,./internal/domain/metrics/controller,./internal/domain/audit/controller \
    --output ./docs \
    --parseDependency \
    --parseInternal \
    --generatedTime

  echo "✅ Swagger: http://localhost:8080/swagger/index.html"
}

lint() {
  ensure_cmd "golangci-lint" \
    "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"

  golangci-lint run ./...
}

mock() {
  ensure_cmd "mockgen" "go install github.com/golang/mock/mockgen@latest"

  mockgen -source=internal/domain/user/repository/user_repository.go \
    -destination=internal/domain/user/repository/mocks/user_repository_mock.go -package=mocks

  mockgen -source=internal/domain/notification/repository/notification_repository.go \
    -destination=internal/domain/notification/repository/mocks/notification_repository_mock.go -package=mocks

  mockgen -source=internal/domain/audit/repository/audit_repository.go \
    -destination=internal/domain/audit/repository/mocks/audit_repository_mock.go -package=mocks

  mockgen -source=internal/domain/metrics/repository/metrics_repository.go \
    -destination=internal/domain/metrics/repository/mocks/metrics_repository_mock.go -package=mocks
}

docker_up() {
  docker compose -f deployments/docker-compose.yml up -d
  echo ">> Kafka UI: http://localhost:8090"
}

docker_down() {
  docker compose -f deployments/docker-compose.yml down
}

docker_logs() {
  docker compose -f deployments/docker-compose.yml logs -f
}

docker_build() {
  docker build -f deployments/Dockerfile -t "$APP_NAME:latest" .
}

migrate_up() {
  bash scripts/migrations/migrate.sh up
}

migrate_down() {
  bash scripts/migrations/migrate.sh down
}

tidy() {
  go mod tidy
}

clean() {
  rm -rf bin/ coverage.out coverage.html
}

dev() {
  docker_up
  tidy
  swagger
  log "Stack prête. Lancement de l'API..."
  go run "$MAIN"
}

# ─────────────────────────────────────────────
# Dispatcher (équivalent make target)
# ─────────────────────────────────────────────

case "$1" in
  build) build ;;
  run) run ;;
  test) test ;;
  test-coverage) test_coverage ;;
  swagger) swagger ;;
  lint) lint ;;
  mock) mock ;;
  docker-up) docker_up ;;
  docker-down) docker_down ;;
  docker-logs) docker_logs ;;
  docker-build) docker_build ;;
  migrate-up) migrate_up ;;
  migrate-down) migrate_down ;;
  tidy) tidy ;;
  clean) clean ;;
  dev) dev ;;
  *)
    echo "Usage: $0 {build|run|test|test-coverage|swagger|lint|mock|docker-up|docker-down|docker-logs|docker-build|migrate-up|migrate-down|tidy|clean|dev}"
    exit 1
    ;;
esac