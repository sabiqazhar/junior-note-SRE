# Catatan konsep — Week 08 · Production Readiness & Reliability Baseline

> Tulis ulang dengan bahasa lo sendiri di bawah tiap poin. Kalau nggak bisa jelasin, berarti belum paham.

## PRR = audit 5 kategori
1. Security — secret, non-root, scan image, port internal tertutup, TLS + auto-renew.
2. Availability — >1 instance / SPOF dinyatakan, health+readiness terpisah, graceful shutdown, rollback.
3. Observability — metrics RED, log terstruktur + correlation ID, trace, alert service down, dashboard.
4. Recovery — backup otomatis + restore teruji, RPO/RTO tertulis, runbook 3 skenario.
5. Deployment — pipeline otomatis, artifact immutable (tag SHA), rollback teruji, nggak ada langkah manual.

## Aturan status
- `TODO` = belum. `PARTIAL` = sebagian. `DONE` = **teruji ada bukti**.
- Jangan tulis `DONE` kalau belum pernah diuji.

## FMEA (wajib ada kolom "deteksi")
| Komponen | Gagal | Dampak | Deteksi | Mitigasi | SPOF? |
| --- | --- | --- | --- | --- | --- |
| PostgreSQL | mati/lambat | tulis gagal | readiness + latency | pool+timeout+backup | YA |
| Redis | mati | cache miss → DB banjir | hit ratio | fallback kosong | ya (ada fallback) |
| Nginx | mati | akses gagal | healthcheck eksternal | — | YA |

## SPOF yang diterima (tulis eksplisit)
- Single instance DB (belum ada replica).
- No multi-AZ.
- Kapasitas belum diuji produksi.

## Syarat tertinggi
- Dari repo kosong → SATU perintah sampai jalan.
- Reboot → pulih sendiri, catat RTO.

## Self-check (jawab tanpa Google)
1. Apa bedanya PRR dan "deploy sukses"?
2. Kenapa FMEA butuh kolom deteksi?
3. Sebutkan 3 SPOF di lab.
4. Kenapa "satu perintah dari repo kosong" penting?
5. Kenapa risiko harus ditulis eksplisit?
