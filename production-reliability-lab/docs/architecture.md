# Arsitektur Lab

## Diagram alur request

```
                    Client
                      │
                      │ 1. DNS
                      │ 2. TCP handshake
                      │ 3. TLS 1.3 (ALPN: h2)
                      ▼
              ┌───────────────┐
              │  Nginx :80    │   ← satu-satunya yang terekspos ke host
              │  (reverse     │     proxy_read_timeout 30s
              │   proxy)      │     X-Forwarded-For / Proto
              └───────┬───────┘
                      │ 4. HTTP/1.1 keep-alive ke upstream
                      ▼
              ┌───────────────┐
              │  Go API :8080 │
              │               │
              │  /healthz  ← liveness  (cek proses saja)
              │  /readyz   ← readiness (cek DB + Redis)
              │  /metrics  ← Prometheus (jangan publik)
              │  /orders   ← contoh endpoint
              │  /slow     ← buat nguji timeout
              └───┬───────┬───┘
                  │       │
        ┌─────────▼─┐   ┌─▼─────────┐
        │ PostgreSQL│   │   Redis   │
        │    :5432  │   │   :6379   │
        │ (volume)  │   │ (volume)  │
        └───────────┘   └───────────┘
             │
             ▼
      Object storage (backup terenkripsi)

─────────── Observability ───────────
API  ──/metrics──▶ Prometheus ──▶ Grafana (dashboard + alert)
Log  ──stdout────▶ Alloy ───────▶ Loki ──▶ Grafana
Trace──OTLP──────▶ Alloy ───────▶ Tempo ──▶ Grafana
```

## Batas kepercayaan

| Zona | Yang boleh diakses | Yang tidak boleh |
| --- | --- | --- |
| Internet / host | Nginx :80 | Semua yang lain |
| Jaringan internal Compose | API → Postgres, API → Redis | — |
| Host | `-` | Postgres, Redis, Prometheus, Loki, Grafana langsung |

Port internal (9090, 3000, 3100, 4317) **sengaja tidak di-publish**. Akses pakai `docker compose exec` atau port-forward, bukan `ports:`.

## Keputusan dan alasannya

| Keputusan | Alasan |
| --- | --- |
| Go untuk service utama | Sudah dikuasai; fokus ke reliability, bukan ke bahasa baru |
| Compose, bukan Kubernetes, sampai minggu 15 | Overhead orchestration belum sepadan sebelum observability beres |
| Nginx di depan, bukan langsung expose API | Satu titik buat TLS, timeout, buffering, dan log akses |
| Alloy, bukan Promtail | Promtail EOL 2 Maret 2026 |
| Image ber-tag commit SHA | Rollback harus deterministik |
| `/healthz` ≠ `/readyz` | Liveness yang cek dependency memicu restart massal saat DB down |

## Single point of failure yang diketahui

1. **PostgreSQL** — belum ada replika. Mitigasi sementara: backup + PITR.
2. **Nginx** — satu instance. Di produksi butuh load balancer dengan health check.
3. **Satu node/host** — kegagalan mesin = layanan mati. Keterbatasan lab, dicatat eksplisit.
