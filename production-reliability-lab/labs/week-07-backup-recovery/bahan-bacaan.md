# Minggu 07 · Backup dan disaster recovery — Bahan Bacaan

> Dokumen ini berdiri sendiri. Baca dari atas ke bawah tanpa buka apa pun lain. Intinya satu: **backup yang belum pernah di-restore bukan backup, cuma harapan.** Baru setelah itu kerjakan lab.md.

## 1. Kenapa minggu ini penting

Ada tim yang tiap Jumat jalanin `pg_dump` lalu simpen ke folder `backups/`. Mereka merasa aman. Enam bulan kemudian, disk produksi rusak. Pas mau restore, ketemu: file dump-nya rusak (karena disk yang sama ikut rusak), dan `pg_dump` nggak bisa balikin data ke "10 menit sebelum ada DELETE salah" karena **`pg_dump` itu logical backup, nggak support Point-in-Time Recovery (PITR)**. Mereka punya backup, tapi nggak punya recovery.

Kejadian kayak gini bukan langka. Rata-rata insiden data hilang itu bukan karena nggak ada backup, tapi karena: (1) backup disimpen di disk yang sama dengan database (sekali disk mati, dua-duanya hilang), (2) nggak pernah diuji restore-nya, (3) nggak tahu berapa lama Recovery Time Objective (RTO)-nya — pas dibutuhin baru sadar butuh 4 jam. Minggu ini lo bikin supaya "data bisa dipulihkan" itu **terukur dan terbukti**, bukan cuma "pernah jalanin cron".

Analogi: backup itu kayak jaket pelampung di kapal. Kalau lo cuma taruh di gudang dan nggak pernah coba pakainya, pas kapal tenggelam lo baru tahu jaketnya bolong atau ukurannya salah. Drill restore = latihan pakai jaket pelampung **di kolam dangkal** (database disposable), bukan langsung di tengah laut.

## 2. Konsep inti

### 2.1 RPO dan RTO — dua angka yang wajib lo tahu

| Istilah | Arti | Contoh target lab |
| --- | --- | --- |
| **RPO** (Recovery Point Objective) | Maksimal data yang *boleh* hilang | "≤ 5 menit transaksi" |
| **RTO** (Recovery Time Objective) | Maksimal waktu sistem boleh mati | "≤ 30 menit" |

RPO ditentukan oleh seberapa sering WAL diarsipkan (`archive_timeout`). RTO ditentukan oleh seberapa cepat lo bisa restore + replay + verifikasi. Dua angka ini harus **tertulis dan diukur**, bukan tebakan.

### 2.2 Logical vs physical backup

- **Logical** (`pg_dump`/`pg_restore`): ekspor jadi SQL/biner. Gampang, portable, tapi lambat buat DB gede dan **nggak bisa PITR** — cuma bisa balik ke saat dump diambil.
- **Physical** (base backup + WAL archive): salinan blok disk Postgres. Cepat restore DB besar, dan **wajib buat PITR**.

```bash
# Logical (cepat buat sanity check)
docker compose exec -T postgres pg_dump -U app -d sre_lab -Fc > backup.dump
pg_restore --list backup.dump        # lihat isi tanpa restore

# Physical + PITR (pgBackRest)
pgbackrest --stanza=app --type=full backup
pgbackrest --stanza=app --type=incr backup
```

### 2.3 PITR butuh WAL archiving

PITR = balik ke titik waktu tertentu (mis. 10:47 sebelum ada DELETE salah). Syarat mutlak: **`archive_mode = on`** dan WAL diarsipkan ke repo terpisah. Kalau `archive_mode` mati, WAL nggak kearsip → PITR gagal total, meski full backup ada.

```ini
# postgresql.conf (inti PITR)
archive_mode = on
archive_command = 'pgbackrest --stanza=app archive-push %p'
archive_timeout = 60s        # WAL diarsipkan paling lambat tiap 60 detik -> RPO ~1 file
```

### 2.4 pgBackRest — standar de-facto

Tool yang dipakai lab ini. Fitur: full/diff/incr, kompresi, enkripsi, repo S3/MinIO, dan PITR.

```ini
# pgbackrest.conf (repo)
[global]
repo1-type=s3
repo1-s3-bucket=sre-lab-backup
repo1-s3-endpoint=minio:9000
repo1-cipher-type=aes-256-cbc     # backup terenkripsi
repo1-cipher-pass=<dari secret, BUKAN plaintext>
repo1-retention-full=7
repo1-retention-diff=14
compress-type=zst
```

```bash
pgbackrest --stanza=app info        # daftar backup + ukuran + waktu
pgbackrest --stanza=app verify      # cek integritas repo
```

### 2.5 WAL-G — alternatif ringan

WAL-G (v3.x) lebih minimalis & cloud-native: satu binary yang ngurus base backup + WAL push ke S3/GCS/Azure. Cocok kalau lo mau setup simpel. Di lab ini kita pakai pgBackRest karena fitur retention & verify-nya lebih lengkap, tapi tau WAL-G itu nilai plus buat wawancara.

### 2.6 Retensi & off-site

Backup dan WAL **harus** di bucket terpisah dari volume database. Lebih bagus lagi: beda region (off-site) dan pakai **object lock / versioning** supaya ransomware atau hapus tidak sengaja nggak ngerusak backup. Retensi WAL nggak boleh lebih pendek dari retensi full backup, kalau nggak backup jadi nggak berguna.

### 2.7 Restore drill — inti minggu ini

Backup yang nggak pernah di-restore bukan bukti. Drill harus:
1. Ke **instance/database baru yang kosong** (disposable — jangan restore menimpa produksi).
2. Verifikasi datanya beneran benar: jumlah row, constraint, index, foreign key, dan checksum sederhana.
3. Ukur waktunya, bandingkan dengan RTO.

```bash
# restore ke instance BARU, ke titik waktu tertentu
pgbackrest --stanza=app --type=time \
  --target="2026-10-12 10:00:00+00" --target-action=promote restore

psql -c "SELECT count(*) FROM orders;"
psql -c "SELECT md5(string_agg(id::text, ',' ORDER BY id)) FROM orders;"  # checksum
psql -c "\d+ orders"        # constraint & index ikut kebawa?
```

> **Simulasikan kehilangan data Cuma di database yang disposable.** Jangan pernah latihan `DROP TABLE` di database pribadi atau produksi. Aturan nggak bisa ditawar.

## 3. Mekanisme di balik layar

Dari `pgbackrest backup` sampai data kembali, ini yang terjadi:

1. **Full backup** → pgBackRest salin semua file data Postgres ke repo (S3/MinIO), hitung checksum tiap file, tulis manifest.
2. **WAL streaming/archive** → tiap segmen WAL beres, `archive-push` kirim ke repo. Ini yang ngisi celah antar full backup (RPO kecil).
3. **Incremental** → cuma salin file yang berubah (atau blok yang berubah), hemat waktu & tempat.
4. **Restore** → pgBackRest keluarkan base backup ke direktori data baru, lalu **replay WAL** sampai titik target (`--type=time`).
5. **Promote** (`--target-action=promote`) → pas sampai titik target, database diangkat jadi bisa dibaca/tulis (berhenti replay).
6. **Verifikasi** → cek row count, checksum, constraint. Bukan cuma "exit code 0".
7. **Ukur** → catat durasi backup + restore + verifikasi, bandingkan dengan RTO.

## 4. Glosarium mini

| Istilah | Arti sederhana | Contoh |
| --- | --- | --- |
| **RPO** | Maksimal data yang boleh hilang | ≤ 5 menit |
| **RTO** | Maksimal waktu mati | ≤ 30 menit |
| **Logical backup** | Ekspor jadi SQL/biner | `pg_dump` |
| **Physical backup** | Salinan blok disk | base backup pgBackRest |
| **WAL** | Log perubahan berurutan | `archive-push` |
| **PITR** | Balik ke titik waktu tertentu | `--type=time` |
| **Stanza** | Config "set backup" di pgBackRest | `--stanza=app` |
| **Base backup** | Fondasi physical backup | full backup |
| **Archive mode** | Apakah WAL diarsipkan | `archive_mode = on` |
| **Retention** | Berapa lama backup disimpan | `retention-full=7` |
| **Off-site** | Backup di lokasi beda | bucket region lain |
| **Object lock** | Cegah hapus/ubah (ransomware) | S3 object lock |

## 5. Bahan bacaan wajib

| # | Sumber | Yang dicari |
| --- | --- | --- |
| 1 | [PostgreSQL: Backup and Restore](https://www.postgresql.org/docs/current/backup.html) | beda logical vs physical, kapan pakai masing-masing |
| 2 | [Continuous Archiving & PITR](https://www.postgresql.org/docs/current/continuous-archiving.html) | `archive_mode`, `restore_command`, `recovery_target_time` |
| 3 | [pgBackRest](https://pgbackrest.org/) | fitur, user guide, command reference |
| 4 | [pgBackRest User Guide](https://pgbackrest.org/user-guide.html) | setup stanza + S3 repo + restore |
| 5 | [WAL-G — GitHub](https://github.com/wal-g/wal-g) | alternatif ringan buat base backup + WAL |
| 6 | [`pg_dump`](https://www.postgresql.org/docs/current/app-pgdump.html) | opsi `-Fc`, `-a`, `--clean` |

## 6. Self-check

1. Kenapa `pg_dump` nggak bisa PITR, dan apa yang dibutuhkan buat PITR?
2. Apa beda RPO dan RTO, dan angka berapa yang lo pasang buat lab ini?
3. Kalau `archive_mode = off`, apa yang rusak di rencana recovery lo?
4. Kenapa restore harus ke instance BARU yang kosong, bukan menimpa produksi?
5. Kenapa "exit code 0 pas restore" bukan bukti data benar?
6. Apa bahaya kalau retensi WAL lebih pendek dari retensi full backup?

## 7. Yang sering salah paham

| Mitos | Kenyataan |
| --- | --- |
| "`pg_dump` tiap malam = aman." | Itu logical, nggak bisa PITR, dan belum tentu bisa di-restore. |
| "Backup ada di folder yang sama, lebih gampang." | Disk mati → backup ikut hilang. Harus terpisah + off-site. |
| "Restore sukses karena nggak ada error." | Exit 0 bukan bukti data benar; cek row/constraint/checksum. |
| "WAL nggak penting, full backup cukup." | Tanpa WAL, celah antar backup gede → RPO jelek, PITR gagal. |
| "Drill restore ribet, nanti kalau butuh baru." | Pas butuh baru tau rusak = terlambat. Drill minimal sebulan sekali. |
| "Enkripsi backup optional." | Backup sering jadi sumber bocor data; enkripsi wajib. |
