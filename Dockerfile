FROM golang:1.20-bullseye

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

ENTRYPOINT ["/app/main"]