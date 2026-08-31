FROM golang:1.25.0-alpine AS builder

WORKDIR /app

RUN apk add --no-cache \
    curl \
    gcc \
    musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files
RUN templ generate

# Install Tailwind CSS CLI
RUN curl -sL \
    https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.13/tailwindcss-linux-arm64 \
    -o /usr/local/bin/tailwindcss \
    && chmod +x /usr/local/bin/tailwindcss

# Build Tailwind CSS
RUN tailwindcss \
    -i css/input.css \
    -o css/output.css \
    --minify

# Build application
RUN CGO_ENABLED=1 GOOS=linux \
    go build -o app ./cmd


FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/app .

COPY --from=builder /app/css/output.css ./css/output.css

CMD ["./app"]