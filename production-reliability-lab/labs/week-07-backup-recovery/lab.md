# Lab — Minggu 07 · Backup dan disaster recovery

> Ikuti BERURUTAN. Setiap langkah: perintah → Yang harus terjadi → Kalau gagal → Catat notes.
> **Aturan mutlak:** simulasi kehilangan data HANYA di database disposable. Jangan `DROP TABLE` di produksi/lab yang jadi dependency.

## Persiapan (10 menit)

- [ ] `docker compose version` → v2.x.
- [ ] Stack jalan: `make up`. `curl -fsS http://localhost/readyz` → 200.
- [ ] Ada data: pastikan tabel `orders` punya beberapa row (jalankan load test `make loadtest` sebentar kalau kosong).
- [ ] MinIO (atau S3-compatible) jalan sebagai repo pgBackRest — cek `docker compose ps minio`.

```bash
docker compose ps postgres minio
psql "$(grep DATABASE_URL .env | cut -d= -f2)" -c "SELECT count(*) FROM orders;"
```

**Yang harus terjadi:** kedua service up, `orders` punya row.
**Kalau gagal:** `make logs postgres` buat lihat kenapa DB nggak sehat.
**Catat di notes.md:** jumlah row awal sebagai baseline verifikasi nanti.

## Hari 1 · Teori (2 jam)

### Langkah 1.1 — Baca bahan bacaan

Buka `bahan-bacaan.md`, baca sampai selesai.

**Yang harus terjadi:** lo bisa jawab 6 self-check tanpa Google.
**Kalau gagal:** ulangi bagian 2 (logical vs physical, PITR).
**Catat di notes.md:** 1 kalimat beda `pg_dump` vs base backup + WAL.

### Langkah 1.2 — Catat RPO/RTO lab

Tulis di `notes.md`: RPO ≤ 5 menit (karena `archive_timeout=60s`), RTO ≤ 30 menit (target, akan diukur).

**Yang harus terjadi:** ada dua angka tertulis + alasan.
**Kalau gagal:** baca `docs/disaster-recovery.md` bagian "Target lab".
**Catat di notes.md:** angka + alasan.

### Langkah 1.3 — Baca `docs/disaster-recovery.md`

Fokus ke strategi backup (full/incr/WAL) dan runbook restore.

**Yang harus terjadi:** lo paham urutan restore: isolate → instance kosong → restore → verifikasi → catat waktu.
**Kalau gagal:** baca lagi bagian 3 bahan-bacaan.
**Catat di notes.md:** 3 hal yang diverifikasi setelah restore.

## Hari 2 · Praktik (2 jam)

### Langkah 2.1 — Logical dump (sanity check cepat)

```bash
docker compose exec -T postgres pg_dump -U app -d sre_lab -Fc > /tmp/backup.dump
pg_restore --list /tmp/backup.dump | head -5
ls -lh /tmp/backup.dump
```

**Yang harus terjadi:** file dump kebentuk; `--list` keluar daftar objek (tabel, index).
**Kalau gagal:** `role app does not exist` → pastikan `-U app` sesuai user DB.
**Catat di notes.md:** ukuran dump sebagai pembanding.

### Langkah 2.2 — Cek konfigurasi PITR (archive_mode)

```bash
docker compose exec postgres psql -U app -d sre_lab -c "SHOW archive_mode;"
docker compose exec postgres psql -U app -d sre_lab -c "SHOW archive_command;"
```

**Yang harus terjadi:** `archive_mode = on`, `archive_command` arahkan ke pgBackRest.
**Kalau gagal:** kalau `off` → PITR mustahil; catat sebagai temuan & perbaiki `postgresql.conf`.
**Catat di notes.md:** nilai keduanya.

### Langkah 2.3 — Full backup (pgBackRest)

```bash
time docker compose exec pgbackrest pgbackrest --stanza=app --type=full backup
docker compose exec pgbackrest pgbackrest --stanza=app info
```

**Yang harus terjadi:** backup selesai; `info` keluar daftar backup + ukuran + waktu.
**Kalau gagal:** `repo not found` → cek koneksi MinIO & `repo1-s3-*` di config pgBackRest.
**Catat di notes.md:** durasi full backup + ukuran.

### Langkah 2.4 — Incremental backup

```bash
docker compose exec pgbackrest pgbackrest --stanza=app --type=incr backup
docker compose exec pgbackrest pgbackrest --stanza=app verify
```

**Yang harus terjadi:** incr selesai lebih cepat; `verify` keluar `status: ok`.
**Kalau gagal:** verify gagal → jangan lanjut ke restore, periksa integritas repo.
**Catat di notes.md:** durasi incr + hasil verify.

## Hari 3 · Eksperimen kegagalan (2 jam) — RESTORE DRILL DI DB DISPOSABLE

> WAJIB: restore ke instance BARU yang kosong (disposable). Jangan restore menimpa `postgres` yang jalan.

### Langkah 3.1 — Siapkan instance disposable (terpisah)

```bash
# container Postgres "rongsokan" terpisah, volume sendiri, akan dibuang habis
docker run -d --name pg-restore \
  -e POSTGRES_USER=app -e POSTGRES_PASSWORD=labpassword -e POSTGRES_DB=sre_lab \
  -v pgrestore_vol:/var/lib/postgresql/data postgres:18
sleep 5
docker ps --filter name=pg-restore
```

**Yang harus terjadi:** `pg-restore` status `Up`, database KOSONG (belum ada tabel `orders`).
**Kalau gagal:** port bentrok → tambah `-p 5433:5432`; volume lama → pakai nama volume baru.
**Catat di notes.md:** ini instance disposable, akan di-`rm -f` setelah drill.

### Langkah 3.2 — Restore ke instance disposable (PITR titik waktu)

```bash
# hentikan server disposable dulu, lalu restore dari repo pgBackRest
docker stop pg-restore
docker run --rm --name pg-restore-restore \
  -e PGBACKREST_STANZA=app -e PGBACKREST_REPO1_... \
  pgbackrest --stanza=app --type=time \
  --target="2026-10-12 10:00:00+00" --target-action=promote restore
```

**Yang harus terjadi:** pgBackRest keluarkan base backup lalu replay WAL sampai titik target, lalu promote.
**Kalau gagal:** `target time not found` → kurangi `--target` ke waktu yang pasti ada (pakai `info` buat lihat range).
**Catat di notes.md:** titik waktu target + durasi restore (ukur dengan `time`).

> Catatan praktis: di lab ini pgBackRest jalan sebagai container terpisah (`pgbackrest`) yang share repo MinIO. Sesuaikan env/volume agar container restore bisa baca repo yang sama. Intinya: **restore ke direktori data container disposable, bukan ke `postgres` produksi lab.**

### Langkah 3.3 — Verifikasi data (jangan berhenti di exit code 0)

```bash
docker start pg-restore
sleep 5
psql -h localhost -p 5433 -U app -d sre_lab -c "SELECT count(*) FROM orders;"
psql -h localhost -p 5433 -U app -d sre_lab -c "SELECT md5(string_agg(id::text, ',' ORDER BY id)) FROM orders;"
psql -h localhost -p 5433 -U app -d sre_lab -c "\d+ orders"
```

**Yang harus terjadi:** jumlah row == baseline (atau sesuai titik waktu target); checksum konsisten; constraint + index + FK ikut kebawa.
**Kalau gagal:** row 0 / constraint hilang → restore salah stanza atau target waktu kelewat; ulangi 3.2.
**Catat di notes.md:** row count, checksum, status constraint — ini bukti sejati.

### Langkah 3.4 — Simulasi "kehilangan data" (HANYA di disposable)

```bash
# ini aman: di container pg-restore, BUKAN di postgres produksi lab
psql -h localhost -p 5433 -U app -d sre_lab -c "DROP TABLE orders;"
psql -h localhost -p 5433 -U app -d sre_lab -c "SELECT count(*) FROM orders;"   # ERROR
```

**Yang harus terjadi:** tabel hilang di disposable → bukti kita butuh restore.
**Kalau gagal:** (justru tabel masih ada) → perintah belum jalan di container yang benar.
**Catat di notes.md:** ini membuktikan mengapa drill penting.

**Cara kembali normal (wajib):**

```bash
# buang instance disposable beserta volumenya — produksi lab TIDAK tersentuh
docker rm -f pg-restore
docker volume rm pgrestore_vol
docker ps --filter name=pg-restore   # harus kosong
make health                         # produksi lab tetap 200
```

### Langkah 3.5 — Catat metrik recovery

Isi di `notes.md`:

| Metrik | Hasil | Target (RTO) |
| --- | --- | --- |
| Waktu full backup | ... | catat apa adanya |
| Waktu restore | ... | ≤ 30 menit |
| Waktu verifikasi | ... | bagian dari RTO |
| Row count pasca-restore | == baseline | ya |

**Yang harus terjadi:** tabel terisi, angka dibandingkan dengan RTO.
**Kalau gagal:** kalau lampaui RTO → tulis kenapa di Hari 4.

## Hari 4 · Review & dokumentasi (2 jam)

### Langkah 4.1 — Isi `docs/disaster-recovery.md` (lengkapi)

Pastikan: target RPO/RTO + alasan, strategi backup (full/incr/WAL), runbook restore, jadwal drill, dan "yang belum ditangani" (off-site region lain, object lock, dll).

**Yang harus terjadi:** dokumen menjawab "gimana restore & buktiin datanya benar".
**Kalau gagal:** lihat lagi `bahan-bacaan.md` bagian 2.7.

### Langkah 4.2 — Lengkapi `commands.md`

Salin perintah Hari 2 & 3 ke `commands.md` (format: perintah → output → interpretasi, min 15).

**Yang harus terjadi:** minimal 15 perintah, terkelompok Hari 2 / Hari 3.
**Kalau gagal:** lo kurang nyatet → balik ke Hari 2/3.

### Langkah 4.3 — Runbook kegagalan

Tulis di `runbooks/README.md`: (1) WAL nggak kearsip (`archive_mode=off`); (2) verify gagal; (3) restore target waktu nggak ketemu. Tiap skenario: gejala → bukti → pemulihan.

**Yang harus terjadi:** tiap skenario punya langkah pemulihan konkret.
**Kalau gagal:** contohin dari `docs/disaster-recovery.md`.

### Langkah 4.4 — Acceptance & commit

```bash
docker ps --filter name=pg-restore   # harus kosong (sudah dibuang)
make health                          # produksi lab tetap sehat
git add labs/week-07-backup-recovery docs/disaster-recovery.md
git commit -m "week 07: backup pgbackrest + restore drill"
```

**Yang harus terjadi:** disposable sudah bersih, produksi sehat, commit masuk, README Referensi terisi.
**Kalau gagal:** jangan `--force` push.

## Daftar periksa "minggu ini selesai"

- [ ] `pg_dump` jalan & `--list` keluar (logical sanity check).
- [ ] `archive_mode = on` terkonfirmasi.
- [ ] Full + incr backup pgBackRest sukses; `verify` ok.
- [ ] Restore drill ke instance DISPOSABLE berhasil.
- [ ] Verifikasi: row count + checksum + constraint lolos.
- [ ] Simulasi `DROP TABLE` cuma di disposable; produksi tidak tersentuh.
- [ ] Metrik recovery diukur & dibandingkan RTO.
- [ ] `commands.md` ≥ 15 perintah; README Referensi terisi.

## Cara membuktikan ke reviewer

1. `pgbackrest --stanza=app info` → daftar backup + ukuran.
2. Output verifikasi restore: `count(*)` == baseline + checksum + `\d+ orders`.
3. Log `time` restore vs RTO ≤ 30 menit.
4. Bukti disposable sudah dibuang (`docker ps` kosong) & produksi `make health` 200.
