FROM golang:1.24-alpine

# Install curl and ca-certificates
RUN apk add --no-cache curl ca-certificates

# Download your root CA PEM file
RUN curl -o /usr/local/share/ca-certificates/root-ca.pem http://pki.jhofer.lan/certs/root-ca.pem

# Update the CA certificates
RUN mv /usr/local/share/ca-certificates/root-ca.pem /usr/local/share/ca-certificates/root-ca.crt && \
    update-ca-certificates

WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -o /usr/local/bin/deploy ./deploy/main.go

ENTRYPOINT ["deploy"]
