# Этап сборки
FROM golang:1.23-alpine AS builder

# Установка необходимых зависимостей
RUN apk add --no-cache git

# Установка рабочей директории
WORKDIR /app

# Копирование файлов с зависимостями
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/http

# Финальный этап
FROM alpine:latest

# Установка часового пояса и сертификатов
RUN apk --no-cache add ca-certificates tzdata

# Создание непривилегированного пользователя
RUN adduser -D appuser

WORKDIR /app

# Копирование бинарного файла из этапа сборки
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Переключение на непривилегированного пользователя
USER appuser

# Определение точки входа
CMD ["./main"]