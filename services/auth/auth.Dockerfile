FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services
COPY libs ./libs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth ./services/auth/cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/auth /app/auth
RUN chmod +x /app/auth

EXPOSE 8082

CMD ["/app/auth"]