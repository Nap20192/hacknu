# Загружаем .env если есть
-include .env
export

DB_URL ?= $(DATABASE_URL)

.PHONY: generate migrate-up migrate-down migrate-drop run build

## sqlc: генерация Go-кода из SQL запросов
generate:
	sqlc generate

## migrate-up: применить все миграции
migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

## migrate-down: откатить последнюю миграцию
migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

## migrate-drop: удалить все таблицы (осторожно!)
migrate-drop:
	migrate -path db/migrations -database "$(DB_URL)" drop -f

## build: собрать бинарники
build:
	go build -o bin/server    ./cmd/server
	go build -o bin/simulator ./cmd/simulator

## run: запустить сервер локально
run:
	go run ./cmd/server

## tidy: обновить зависимости
tidy:
	go mod tidy
