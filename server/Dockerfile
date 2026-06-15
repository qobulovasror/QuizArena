# --- build ---
FROM golang:1.26-alpine AS build
WORKDIR /src
# go.sum hali yo'q (deps qo'shilgach paydo bo'ladi) — shuning uchun faqat go.mod
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

# --- runtime ---
FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
