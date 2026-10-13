# Week 07 · Backup dan disaster recovery

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai backup dan disaster recovery sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Logical vs physical backup
- WAL archiving & PITR
- RPO & RTO
- Retention & off-site
- Restore drill

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Restore benar-benar berhasil ke instance/database baru yang kosong.
2. Data hasil restore lolos verifikasi (jumlah row + constraint + checksum tabel).
3. Waktu backup dan restore diukur, lalu dibandingkan dengan target RPO/RTO.
4. Backup terenkripsi dan tersimpan terpisah dari volume database.

## Struktur folder

```
labs/week-07-backup-recovery/
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

- [PostgreSQL — Backup and Restore](https://www.postgresql.org/docs/current/backup.html) — logical vs physical backup.
- [PostgreSQL — Continuous Archiving & PITR](https://www.postgresql.org/docs/current/continuous-archiving.html) — WAL archiving untuk PITR.
- [pgBackRest documentation](https://pgbackrest.org/) — standar de-facto backup Postgres (full/incr/diff, S3, enkripsi).
- [Barman (EnterpriseDB)](https://www.pgbarman.org/) — alternatif enterprise.
- [WAL-G (GitHub)](https://github.com/wal-g/wal-g) — backup cloud-native ringan.

