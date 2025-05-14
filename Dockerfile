FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o loadbalancer ./cmd/loadbalancer

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/loadbalancer .

COPY --from=builder /app/migrations ./migrations

CMD ["./loadbalancer"]
