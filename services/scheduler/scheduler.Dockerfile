FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services
COPY libs ./libs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/scheduler ./services/scheduler/cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/scheduler /app/scheduler

CMD ["/app/scheduler"]
