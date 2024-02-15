FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod main.go ./

RUN go get
RUN go build -o app main.go

ENTRYPOINT ["/app/app"]