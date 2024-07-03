FROM golang:1.22
LABEL authors="aminrezaei"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .env ./
COPY . .

RUN GOOS=linux go build -o /indexer-go

CMD ["/indexer-go"]