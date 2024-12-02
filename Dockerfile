FROM golang:1.21-alpine

WORKDIR /app
COPY deploy/main.go .
RUN go build -o /usr/local/bin/deploy main.go

ENTRYPOINT ["deploy"]
