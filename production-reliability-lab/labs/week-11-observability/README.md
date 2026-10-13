# Week 11 · Logs dan distributed tracing

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai logs dan distributed tracing sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Structured logging (slog)
- Correlation ID & W3C traceparent
- Trace/span & context propagation
- Sampling
- Metrics vs logs vs traces

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Log terstruktur JSON dan bisa difilter berdasarkan trace/correlation ID.
2. Trace memperlihatkan span aplikasi DAN dependency (Postgres/Redis) beserta durasinya.
3. Satu request lambat sengaja dibuat dan berhasil ditelusuri tanpa menebak.
4. Konteks (context.Context) diteruskan, sehingga trace tidak putus.

## Struktur folder

```
labs/week-11-observability/
├── README.md        <- file ini
├── notes.md         <- catatan konsep Hari 1
├── commands.md      <- semua perintah yang dijalankan + output yang disanitasi
├── runbooks/        <- cara memperbaiki kegagalan (Hari 4)
└── evidence/        <- bukti: screenshot, log, hasil pengukuran
```

## Aturan bukti

1. Setiap acceptance criteria di atas harus punya **bukti yang bisa direproduksi** (perintah + output).
2. Output yang mengandung IP internal, token, password, atau data pribadi **wajib disanitasi** sebelum di-commit.
3. Jangan tandai minggu ini selesai hanya karena tutorialnya sudah ditonton.

## Referensi

1. OpenTelemetry — Documentation (traces, logs, metrics): https://opentelemetry.io/docs/
2. OpenTelemetry — Trace Context (W3C `traceparent`): https://opentelemetry.io/docs/specs/otel/trace/trace-context/
3. OpenTelemetry — Sampling concepts: https://opentelemetry.io/docs/concepts/sampling/
4. Grafana — Loki LogQL query language: https://grafana.com/docs/loki/latest/query/logql/
5. Grafana — Alloy (pengganti Promtail): https://grafana.com/docs/alloy/latest/
6. Grafana — Tempo (distributed tracing backend): https://grafana.com/docs/tempo/latest/

