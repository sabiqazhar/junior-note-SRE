# Week 10 · Grafana dan alerting

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai grafana dan alerting sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- RED & USE metrics
- Dashboard design
- Alert threshold & burn rate
- Alert fatigue
- Notification policy & silence

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Dashboard menampilkan RED + USE secara real time dan bisa di-reprovision dari Git.
2. Alert service-down terpicu dalam target waktu yang lo tentukan (misal <2 menit).
3. Setiap alert punya severity, runbook, dan langkah diagnosis yang jelas.
4. Alert tidak flapping (pakai `for:` + group interval yang benar).

## Struktur folder

```
labs/week-10-grafana/
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

1. Grafana — Alerting overview: https://grafana.com/docs/grafana/latest/alerting/
2. Grafana — Provisioning dashboards & alerts as code: https://grafana.com/docs/grafana/latest/administration/provisioning/
3. Grafana — Alloy (pengganti Promtail): https://grafana.com/docs/alloy/latest/
4. Prometheus — Alerting best practices (`for:`, severity): https://prometheus.io/docs/practices/alerting/
5. Google SRE Book — Monitoring distributed systems (symptom-based): https://sre.google/sre-book/monitoring-distributed-systems/
6. Prometheus — Querying functions: https://prometheus.io/docs/prometheus/latest/querying/functions/

