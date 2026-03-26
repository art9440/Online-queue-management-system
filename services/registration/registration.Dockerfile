FROM golang:1.26-alpine AS builder

WORKDIR /app

# зависимости
COPY go.mod go.sum ./
RUN go mod download

# код
COPY services ./services
COPY libs ./libs

# билд
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o registration ./services/registration/cmd

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/registration .

# безопасность
RUN adduser -D appuser
USER appuser

EXPOSE 8081

CMD ["./registration"]