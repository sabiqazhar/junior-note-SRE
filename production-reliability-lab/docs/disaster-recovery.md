# Disaster Recovery

> Backup yang belum pernah di-restore bukan backup — cuma harapan.

## Target lab

| Metrik | Target | Alasan |
| --- | --- | --- |
| **RPO** | ≤ 5 menit | WAL diarsipkan dengan `archive_timeout=60s`; maksimal 1 file WAL hilang |
| **RTO** | ≤ 30 menit | Restore full backup + replay WAL + verifikasi, diukur saat drill |

Angka ini target latihan. Ukur yang sebenarnya, lalu bandingkan.

## Strategi backup

| Lapisan | Tool | Frekuensi | Retensi | Lokasi |
| --- | --- | --- | --- | --- |
| Full backup | pgBackRest | Mingguan | 7 | Object storage (MinIO/S3), terenkripsi |
| Incremental | pgBackRest | Harian | 14 | Sama |
| WAL archive | pgBackRest | Kontinu (PITR) | Mengikuti retensi full | Sama, beda prefix |
| Logical dump | `pg_dump -Fc` | Sebelum perubahan skema besar | 3 | Lokal, buat jaga-jaga |

`pg_dump` **tidak bisa PITR**. PITR butuh base backup + WAL. Jangan tertukar.

## Perintah

```bash
# Backup
pgbackrest --stanza=app --type=full backup
pgbackrest --stanza=app --type=incr backup
pgbackrest --stanza=app info
pgbackrest --stanza=app verify          # cek integritas repo

# Restore ke instance BARU yang kosong
pgbackrest --stanza=app --type=time \
  --target="2026-10-12 10:00:00+00" --target-action=promote restore

# Logical (cara cepat buat sanity check)
docker compose exec -T postgres pg_dump -U app -d sre_lab -Fc > backup.dump
pg_restore --list backup.dump
```

## Prosedur restore (runbook)

1. **Isolasi.** Pastikan instance target tidak menerima traffic.
2. **Siapkan instance kosong.** Jangan restore menimpa data yang masih bisa diselamatkan.
3. **Restore** dengan target waktu yang jelas.
4. **Verifikasi** — jangan berhenti di "exit code 0":
   ```sql
   SELECT count(*) FROM orders;
   SELECT md5(string_agg(id::text, ',' ORDER BY id)) FROM orders;  -- checksum
   \d+ orders                                                      -- constraint & index
   ```
5. **Catat waktu.** Berapa menit dari mulai restore sampai verifikasi selesai.
6. **Bandingkan dengan RTO.** Kalau lebir, tulis kenapa.

## Drill restore

Jadwalkan **minimal sebulan sekali**. Tanpa drill, lo tidak tahu backup-nya rusak sampai hari lo benar-benar butuh.

| Tanggal drill | Durasi restore | Verifikasi lolos | RTO terpenuhi | Catatan |
| --- | --- | --- | --- | --- |
| | | | | |

## Yang sengaja belum ditangani

- Belum ada restore otomatis berkala (drill masih manual).
- Belum ada off-site ke region berbeda — backup dan WAL masih satu bucket.
- Belum diuji skenario "object storage-nya sendiri hilang".
- Belum ada object lock / immutability terhadap penghapusan tidak sengaja.
