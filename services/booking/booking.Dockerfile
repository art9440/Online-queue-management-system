FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services
COPY libs ./libs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/booking ./services/booking/cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/booking /app/booking

EXPOSE 8083

CMD ["/app/booking"]