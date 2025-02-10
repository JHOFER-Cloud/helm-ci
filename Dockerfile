FROM golang:1.23-alpine

# Install curl
RUN apk add --no-cache curl

WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -o /usr/local/bin/deploy ./deploy/main.go

ENTRYPOINT ["deploy"]
