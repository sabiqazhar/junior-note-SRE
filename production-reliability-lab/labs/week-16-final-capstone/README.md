# Week 16 · Load testing, chaos, dan final incident simulation

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai load testing, chaos, dan final incident simulation sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Capacity planning & Little's Law
- Queueing & saturation
- Retry storm, backoff, jitter
- Timeout budget & circuit breaker
- Chaos experiment & game day

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Load test bisa dijalankan ulang dan punya threshold yang menggagalkan CI.
2. Bottleneck didukung data (saturasi, ρ, koneksi DB), bukan perasaan.
3. Insiden terdeteksi, didiagnosis, dimitigasi, dan dipulihkan dengan timeline tercatat.
4. Hasil sebelum vs sesudah perbaikan bisa dibandingkan kuantitatif.

## Struktur folder

```
labs/week-16-final-capstone/
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

- [SRE Workbook](https://sre.google/workbook/) — capacity planning, error budget, postmortem
- [k6 Documentation (Grafana)](https://grafana.com/docs/k6/latest/) — threshold, scenarios, v2.x
- [k6 Docs](https://k6.io/docs/) — panduan praktis load testing
- [Chaos Mesh](https://chaos-mesh.org/) — chaos experiment di Kubernetes
- [toxiproxy](https://github.com/Shopify/toxiproxy) — inject latency/timeout di level proxy
- [SRE Book: Postmortem Culture](https://sre.google/sre-book/postmortem-culture/) — format insiden

