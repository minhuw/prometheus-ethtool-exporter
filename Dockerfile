FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN apk add --no-cache gcc musl-dev linux-headers
RUN go build -o prometheus-ethtool-exporter

FROM alpine:latest

COPY --from=builder /app/prometheus-ethtool-exporter /usr/local/bin/

EXPOSE 9417
ENTRYPOINT ["/usr/local/bin/prometheus-ethtool-exporter"] 