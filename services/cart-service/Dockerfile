FROM golang:1.23.2-alpine3.20 AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cart-service ./cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/cart-service .

EXPOSE 50053

CMD ["./cart-service"]
