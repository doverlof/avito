FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git make

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o backend-app \
    ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Устанавливаем необходимые утилиты
RUN apk add --no-cache curl ca-certificates tzdata

# Копируем собранное приложение
COPY --from=builder /app/backend-app .

# Создаем непривилегированного пользователя
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=40s \
    CMD curl -f http://localhost:8080/health || exit 1

# Запуск приложения
CMD ["./backend-app"]