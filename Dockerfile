FROM golang:1.26.10-bookworm

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

ENTRYPOINT ["/app/main"]