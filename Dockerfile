FROM golang:1.22.4 AS builder

WORKDIR /app

RUN go install github.com/air-verse/air@v1.52.2
RUN go install github.com/a-h/templ/cmd/templ@v0.2.707


FROM golang:1.22.4

COPY --from=builder /go/bin/air /usr/local/bin/air
COPY --from=builder /go/bin/templ /usr/local/bin/templ

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

CMD ["air"]
