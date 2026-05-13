FROM golang:1.22.4-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o app ./cmd

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/app .

CMD ["./app"]