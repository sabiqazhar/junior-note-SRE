# syntax=docker/dockerfile:1.7
# Multi-stage: builder gede, image runtime kecil + non-root.

# ---------- Stage 1: build ----------
FROM golang:1.25-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

# Cache dependency dulu: ubah kode nggak bikin download ulang.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath -ldflags="-s -w" \
      -o /out/sre-api ./app

# ---------- Stage 2: runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

# healthcheck internal butuh binary ini sendiri (distroless nggak ada curl/wget).
COPY --from=builder /out/sre-api /app/sre-api

USER 65532:65532
EXPOSE 8080

# exec form: PID 1 = binary, supaya SIGTERM nyampe langsung (minggu 6).
ENTRYPOINT ["/app/sre-api"]
