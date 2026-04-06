# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o opencentral .

# ── Stage 2: Run ────────────────────────────────────────────────────────────
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/opencentral .

# Storage and DB live outside the container via a mounted volume
VOLUME ["/storage"]

EXPOSE 8080

CMD ["./opencentral"]
