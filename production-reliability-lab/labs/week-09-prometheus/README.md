# Week 09 · Metrics dengan Prometheus

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai metrics dengan prometheus sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Counter/gauge/histogram
- Scrape interval & label
- Kardinalitas
- PromQL: rate, increase, histogram_quantile
- Native histogram

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Prometheus berhasil scrape `/metrics` (target UP).
2. Metrics berubah sesuai traffic dan error yang dikirim (bukti: sebelum vs sesudah).
3. Query request rate, error ratio, dan p95 latency bisa dijalankan dan hasilnya masuk akal.
4. Tidak ada label ber-kardinalitas tinggi (user_id, path mentah) di metrics.

## Struktur folder

```
labs/week-09-prometheus/
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

1. Prometheus — Getting Started: https://prometheus.io/docs/tutorials/getting_started/
2. Prometheus — Querying basics (`rate`, `increase`, range selectors): https://prometheus.io/docs/prometheus/latest/querying/basics/
3. Prometheus — Histograms and summaries: https://prometheus.io/docs/practices/histograms/
4. Prometheus — Metric and label naming (hindari kardinalitas): https://prometheus.io/docs/practices/naming/
5. Prometheus 3.0 blog (native histogram, OTLP): https://prometheus.io/blog/2024/11/14/prometheus-3-0/
6. Prometheus client_golang (instrumentasi Go): https://github.com/prometheus/client_golang

