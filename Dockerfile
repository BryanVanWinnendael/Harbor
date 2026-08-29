FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o app ./cmd

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache libc6-compat

COPY --from=builder /app/app .

CMD ["./app"]
