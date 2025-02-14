FROM golang:1.24-alpine

# Install curl and ca-certificates
RUN apk add --no-cache curl ca-certificates

# Download your root CA PEM file
RUN curl -o /usr/local/share/ca-certificates/root-ca.crt http://pki.jhofer.lan/certs/root-ca.pem

# Update CA certificates
RUN update-ca-certificates

WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -o /usr/local/bin/deploy ./deploy/main.go
