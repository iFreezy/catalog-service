OUTPUT:=./bin/app
GO_LINT_VERSION=1.64.8
MOCKERY_VERSION=2.53.3

GO_FILE:=./main.go

.PHONY: up
up: ## Поднимает (запускает) окружение для работы приложения
	docker compose up -d

.PHONY: down
down: ## Отключает окружение для работы приложения
	docker compose down --remove-orphans

.PHONY: lint
lint: ## Запуск линтера
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v${GO_LINT_VERSION} run

.PHONY: lint-fix
lint-fix: ## Запуск линтера с фиксом
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v${GO_LINT_VERSION} run --fix

.PHONY: build
build: ## Сборка приложения
	go build -o ${OUTPUT} ${GO_FILE}

.PHONY: test
test: ## Запуск тестов
	go test -count=1 -v ./...

.PHONY: cover
cover: ## Запуск тестов с отчётом о покрытии
	go test -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: mocks
mocks: ## Генерация моков через mockery
	go run github.com/vektra/mockery/v2@v${MOCKERY_VERSION} --config .mockery.yaml

.PHONY: run
run: ## Запуск приложения
	go run ${GO_FILE}

.PHONY: clean
clean: ## Очистка кэша приложения
	go clean