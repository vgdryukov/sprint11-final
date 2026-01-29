# Этап 1: Сборка
FROM golang:1.22-alpine AS builder

# Минимальный набор зависимостей
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY *.go ./

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -o tracker .

# Этап 2: Запуск
FROM alpine:latest

# Устанавливаем только sqlite (утилита командной строки)
RUN apk add --no-cache sqlite

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/tracker /app/tracker

# Создаем таблицу и запускаем приложение
CMD ["sh", "-c", "sqlite3 tracker.db 'CREATE TABLE IF NOT EXISTS parcel (number INTEGER PRIMARY KEY AUTOINCREMENT, client INTEGER NOT NULL, status TEXT NOT NULL, address TEXT NOT NULL, created_at TEXT NOT NULL);' && ./tracker"]