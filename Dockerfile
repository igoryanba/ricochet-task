# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Копируем go мод файлы отдельно для кеша зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка статически линкованного бинаря
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /ricochet-task .

# --- Runtime stage ---
FROM alpine:3.18

# Добавляем wget для healthcheck
RUN apk add --no-cache wget

COPY --from=builder /ricochet-task /usr/local/bin/ricochet-task
ENTRYPOINT ["ricochet-task"]
CMD ["-http"] 