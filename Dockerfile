FROM golang:1.24-alpine

# Install curl and ca-certificates
RUN apk add --no-cache curl ca-certificates

# Download and set up the CA certificate
RUN curl -o /usr/local/share/ca-certificates/root-ca.crt http://pki.jhofer.lan/certs/root-ca.pem \
    && cp /usr/local/share/ca-certificates/root-ca.crt /etc/ssl/certs/root-ca.crt \
    && update-ca-certificates

WORKDIR /app
COPY . .
RUN go mod download
RUN go mod tidy
RUN go build -o /usr/local/bin/deploy ./deploy/main.go

ENTRYPOINT ["deploy"]
