.PHONY: run build tidy sqlc migrate-up migrate-down up down

# Go serverni lokal ishga tushirish (skelet)
run:
	cd server && go run ./cmd/server

build:
	cd server && go build -o bin/server ./cmd/server

tidy:
	cd server && go mod tidy

# DB-kod generatsiya (sqlc kerak)
sqlc:
	cd server && sqlc generate

# Migratsiya (goose + DATABASE_URL env kerak)
migrate-up:
	cd server && goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	cd server && goose -dir migrations postgres "$$DATABASE_URL" down

# Docker Compose
up:
	docker compose up -d --build

down:
	docker compose down
