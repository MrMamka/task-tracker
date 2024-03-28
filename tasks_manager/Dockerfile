FROM golang:1.22-alpine

WORKDIR /tracker/tasks_manager
COPY go.* ./

RUN go mod download -x

COPY . .

RUN go build -o tasks_manager ./cmd/main.go

ENTRYPOINT ["./tasks_manager"]