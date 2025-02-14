FROM golang:1.24-alpine

# Install curl, ca-certificates, and openssl
RUN apk add --no-cache curl ca-certificates openssl

# Download your root CA PEM file
RUN curl -o /usr/local/share/ca-certificates/root-ca.pem http://pki.jhofer.lan/certs/root-ca.pem

# Convert PEM to CRT and update the CA certificates
RUN openssl x509 -outform der -in /usr/local/share/ca-certificates/root-ca.pem -out /usr/local/share/ca-certificates/your-root-ca.crt && \
    update-ca-certificates

WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -o /usr/local/bin/deploy ./deploy/main.go

ENTRYPOINT ["deploy"]
