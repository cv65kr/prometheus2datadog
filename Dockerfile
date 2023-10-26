FROM golang:1.21.3-alpine as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /app/prometheus2datadog ./cmd/main.go

FROM alpine:3.18

RUN addgroup -S app && adduser -S -G app app

WORKDIR /home/app

COPY --from=builder /app/prometheus2datadog .

RUN chown -R app:app ./

USER app

CMD ["./prometheus2datadog"]