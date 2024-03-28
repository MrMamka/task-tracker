FROM golang:1.22-alpine

WORKDIR /tracker/user_service
COPY go.* ./

RUN go mod download -x

COPY . .

RUN go build -o user_service ./cmd/main.go

ENTRYPOINT ["./user_service"]