FROM golang:1.26@sha256:313faae491b410a35402c05d35e7518ae99103d957308e940e1ae2cfa0aac29b AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services ./services
COPY libs ./libs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bot ./services/bot/cmd

FROM debian:bookworm-slim@sha256:67b30a61dc87758f0caf819646104f29ecbda97d920aaf5edc834128ac8493d3

WORKDIR /app

COPY --from=builder --chown=root:root --chmod=555 /app/bot /app/bot

USER 65532:65532

EXPOSE 8085

CMD ["/app/bot"]
