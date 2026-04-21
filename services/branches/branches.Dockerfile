FROM golang:1.26 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/branches ./services/branches/cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/branches /app/branches

RUN chmod +x /app/branches

EXPOSE 8083

CMD ["/app/branches"]