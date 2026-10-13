# Week 12 · SLI, SLO, error budget, dan incident response

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai sli, slo, error budget, dan incident response sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- SLI vs SLO vs SLA
- Error budget & policy
- Multiwindow multi-burn-rate
- Incident lifecycle & severity
- Blameless postmortem

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. SLI bisa dihitung dari metrics yang sudah dikumpulkan (sumber data + window jelas).
2. SLO punya periode evaluasi dan aturan eksklusi request tidak valid.
3. Alert burn rate terpasang dan diuji dengan `promtool test rules`.
4. Satu postmortem blameless lengkap: impact, timeline, root cause, 5-whys, action item ber-owner.

## Struktur folder

```
labs/week-12-slo/
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

1. Google SRE Workbook — Implementing SLOs: https://sre.google/workbook/implementing-slos/
2. Google SRE Workbook — Alerting on SLOs (multiwindow burn-rate): https://sre.google/workbook/alerting-on-slos/
3. Google SRE Book — Embracing risk (error budgets): https://sre.google/sre-book/embracing-risk/
4. Prometheus — Recording rules: https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
5. Prometheus — promtool check/test rules: https://prometheus.io/docs/prometheus/latest/command-line/promtool/
6. Grafana — On-call & incident response: https://grafana.com/docs/oncall/latest/

