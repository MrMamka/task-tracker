FROM golang:1.22-alpine

WORKDIR /tracker/statics_service
COPY go.* ./

RUN go mod download -x

COPY . .

RUN go build -o statics_service ./cmd/main.go

ENTRYPOINT ["./statics_service"]