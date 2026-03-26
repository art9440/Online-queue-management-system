FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./services/migrate/cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations

CMD ["/app/migrate"]