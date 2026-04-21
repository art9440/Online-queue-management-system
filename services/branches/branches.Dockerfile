FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services
COPY libs ./libs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/branches ./services/branches/cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/branches /app/branches

EXPOSE 8083

CMD ["/app/branches"]