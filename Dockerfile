FROM golang:1.22.4

WORKDIR /app

RUN go install github.com/air-verse/air@v1.52.2
RUN go install github.com/a-h/templ/cmd/templ@v0.2.707

COPY go.mod go.sum ./
RUN go mod download

CMD ["air"]