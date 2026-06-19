.PHONY: tools run build tidy sqlc migrate-up migrate-down up down

# Kerakli CLI vositalarini o'rnatish (sqlc, goose)
tools:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

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

# Boshlang'ich ma'lumot (english soha/kategoriya) — DATABASE_URL kerak
seed:
	cd server && go run ./cmd/seed

# Foydalanuvchini admin qilish: make admin EMAIL=ali@example.com
admin:
	cd server && go run ./cmd/setadmin -email=$(EMAIL)

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
