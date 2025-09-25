FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN go mod download && go build -o confession-wall .

FROM debian:bookworm-slim

RUN apt update && apt install -y --no-install-recommends \
    openssl bash ca-certificates \
    && rm -rf /var/lib/apt/lists/*

VOLUME /data
WORKDIR /app
COPY --from=builder /app/confession-wall .
COPY entrypoint.sh /entrypoint.sh
RUN mkdir -p /app/uploads && mkdir -p /app/config && chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 8080
