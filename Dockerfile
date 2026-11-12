# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /dist/main .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /dist/main ./main
COPY --from=builder /build/config.yaml ./config.yaml
COPY --from=builder /build/internal/database/migrations ./internal/database/migrations

ENV TZ=Asia/Jakarta \
    MODE=dev

USER app

EXPOSE 8080

# Run migrations (PostgreSQL) then start the API. With DB_DRIVER=memory the
# migration command is skipped by docker-compose's command override.
CMD ["/bin/sh", "-c", "./main migrate --direction=up && ./main server"]
