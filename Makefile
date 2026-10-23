APP_NAME=shieldgate
CMD_PATH=./cmd/shieldgate/main.go
MIGRATOR_PATH=./cmd/migrator/main.go
DOCKER_COMPOSE=docker-compose

.PHONY: help up down infra-up infra-down run dev restart-app logs logs-all test test-race lint build clean

help:
	@echo "Доступные команды:"
	@echo "  make infra-up     - Запустить ТОЛЬКО Postgres и Redis (для быстрой локальной разработки)"
	@echo "  make infra-down   - Остановить базы данных"
	@echo "  make run          - Запустить Go-сервер локально"
	@echo "  make migrate	   - Запустить миграции к БД (требуется поднятая БД)"
	@echo "  make dev          - Запустить с автоперезагрузкой (hot-reload через утилиту air)"
	@echo "  make up           - Запустить ВСЁ окружение в Docker (App + DB + Redis + Mock)"
	@echo "  make restart-app  - Пересобрать и перезапустить ТОЛЬКО Go-приложение в Docker (без рестарта БД!)"
	@echo "  make down         - Остановить всё Docker окружение"
	@echo "  make logs         - Смотреть логи Go-приложения в реальном времени"
	@echo "  make test         - Запустить Unit-тесты с детектором гонок (-race)"
	@echo "  make lint         - Прогнать линтер golangci-lint"


infra-up:
	$(DOCKER_COMPOSE) up -d postgres redis mock-upstream

infra-down:
	$(DOCKER_COMPOSE) stop postgres redis mock-upstream

run:
	go run $(CMD_PATH)

migrate:
	go run $(MIGRATE_PATH)
dev:
	air

up:
	$(DOCKER_COMPOSE) up --build -d

down:
	$(DOCKER_COMPOSE) down

restart-app:
	$(DOCKER_COMPOSE) up -d --build shieldgate

logs:
	$(DOCKER_COMPOSE) logs -f shieldgate

logs-all:
	$(DOCKER_COMPOSE) logs -f

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/$(APP_NAME) $(CMD_PATH)

clean:
	rm -rf bin/