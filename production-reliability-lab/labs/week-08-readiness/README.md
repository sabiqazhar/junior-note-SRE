# Week 08 · Production readiness & reliability baseline

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai production readiness & reliability baseline sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Production Readiness Review
- Dependency mapping
- Failure mode (FMEA)
- Single point of failure
- Operational checklist

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Sistem bisa dibangun ulang dari nol hanya dari repository.
2. Semua langkah pemulihan terdokumentasi DAN sudah diuji (bukan sekadar ditulis).
3. Risiko yang belum ditangani ditulis eksplisit beserta alasannya.
4. Environment pulih otomatis setelah reboot/restart.

## Struktur folder

```
labs/week-08-readiness/
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

- [Google SRE — Evolving SRE Engagement Model (PRR)](https://sre.google/sre-book/evolving-sre-engagement-model/) — konsep Production Readiness Review.
- [Launch Coordination Checklist](https://sre.google/sre-book/launch-checklist/) — checklist peluncuran produksi.
- [Google SRE — Failure Modes and Root Cause Analysis](https://sre.google/sre-book/failure-modes/) — FMEA & analisis kegagalan.
- [Docker Compose startup order (`depends_on`)](https://docs.docker.com/compose/how-tos/startup-order/) — `condition: service_healthy`.
- `docs/production-readiness.md` (di repo ini) — template 5 kategori + dependency map.

