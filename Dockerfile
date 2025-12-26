# --- Stage 1: build ---
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Вытаскиваем модули отдельно (кэш быстрее)
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o analytics-service .

# --- Stage 2: run ---
FROM alpine:latest

WORKDIR /app

# Для HTTPS-запросов (если вдруг пригодится)
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/analytics-service .

# PORT и REDIS_ADDR можно переопределить через env
ENV PORT=8080
ENV REDIS_ADDR=redis:6379

EXPOSE 8080

CMD ["./analytics-service"]
