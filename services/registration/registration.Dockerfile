FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services
COPY libs ./libs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/registration ./services/registration/cmd

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/registration /app/registration

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

#RUN chmod +x /app/registration

EXPOSE 8081

CMD ["/app/registration"]