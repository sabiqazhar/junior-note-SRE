# Lab — Minggu 06 · Secrets, konfigurasi, dan deployment

> Ikuti BERURUTAN. Setiap langkah: perintah → Yang harus terjadi → Kalau gagal → Catat notes.

## Persiapan (10 menit)

- [ ] `docker compose version` → `Docker Compose version v2.x`.
- [ ] `go version` → 1.22+.
- [ ] `sops --version` terpasang (atau `brew install sops` / `go install github.com/getsops/sops/v3/cmd/sops@latest`).
- [ ] Punya age key: `age-keygen -o ~/.config/sops/age/keys.txt` → simpan public key.
- [ ] Repo jalan: `make up` lalu `make health` → `/health` & `/readyz` 200.

```bash
docker compose -f docker/compose.yaml config >/dev/null && echo "compose valid"
curl -fsS http://localhost/health && echo " health ok"
curl -fsS http://localhost/readyz && echo " readyz ok"
```

**Yang harus terjadi:** ketiga perintah keluar tanpa error.
**Kalau gagal:** `make logs api` buat lihat kenapa container nggak sehat.
**Catat di notes.md:** cara lo bikin age key dan public key-nya.

## Hari 1 · Teori (2 jam)

### Langkah 1.1 — Baca bahan bacaan

Buka `bahan-bacaan.md`, baca sampai selesai.

**Yang harus terjadi:** lo bisa jawab 6 self-check tanpa Google.
**Kalau gagal:** ulangi bagian 2 (konsep inti), khususnya tabel probe.
**Catat di notes.md:** 1 kalimat beda liveness vs readiness dengan bahasa lo.

### Langkah 1.2 — Bedah `app/main.go`

```bash
sed -n '94,119p' app/main.go
```

**Yang harus terjadi:** lo lihat `signal.NotifyContext`, `srv.Shutdown(c)`, `db.Close()`.
**Kalau gagal:** buka `app/main.go` langsung, cari fungsi `main`.
**Catat di notes.md:** baris mana yang ngejamin timeout shutdown < grace period.

### Langkah 1.3 — Baca `docs/deployment.md`

Fokus ke tabel "Kriteria rollback".

**Yang harus terjadi:** lo tahu ambang error rate (2%) dan p95 (2× baseline).
**Kalau gagal:** baca lagi bagian 4 bahan-bacaan.
**Catat di notes.md:** 3 kondisi yang wajib rollback.

## Hari 2 · Praktik (2 jam)

### Langkah 2.1 — Enkripsi file env pakai SOPS + age

```bash
age-keygen -o ~/.config/sops/age/keys.txt
export AGE_PUBLIC=$(grep -o 'age1[^ ]*' ~/.config/sops/age/keys.txt)
sops --encrypt --age "$AGE_PUBLIC" .env.example > .env.enc
file .env.enc
head -c 80 .env.enc; echo
```

**Yang harus terjadi:** `.env.enc` isinya terenkripsi (bukan plaintext).
**Kalau gagal:** `age-keygen` error → cek folder `~/.config/sops` writable.
**Catat di notes.md:** perintah dekripsi yang akan lo pakai di CI.

### Langkah 2.2 — Dekripsi balik

```bash
sops --decrypt .env.enc > /tmp/.env.dec
diff <(grep -v '^#' .env.example) <(grep -v '^#' /tmp/.env.dec) && echo "isi sama"
rm /tmp/.env.dec
```

**Yang harus terjadi:** `isi sama` → dekripsi menghasilkan isi identik.
**Kalau gagal:** `AGE_SECRET_KEY` belum diset → `export AGE_SECRET_KEY=$(grep -o 'AGE-SECRET-KEY[^ ]*' ~/.config/sops/age/keys.txt)`.
**Catat di notes.md:** bukti round-trip enkripsi berhasil.

### Langkah 2.3 — Jalankan & cek kedua endpoint

```bash
make up
curl -s -o /dev/null -w "healthz=%{http_code}\n" http://localhost/health
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
```

**Yang harus terjadi:** keduanya `200`.
**Kalau gagal:** `make logs api` → cek apakah `DATABASE_URL`/`REDIS_URL` salah.
**Catat di notes.md:** behaviour normal sebagai baseline.

### Langkah 2.4 — Lihat exec form di Dockerfile

```bash
grep -n 'ENTRYPOINT\|CMD' docker/Dockerfile
```

**Yang harus terjadi:** baris `["/app/sre-api"]` (exec form, bukan shell).
**Kalau gagal:** kalau pakai `/app/sre-api` tanpa array → sinyal nggak nyampe; ubah ke exec form.
**Catat di notes.md:** kenapa exec form penting buat graceful shutdown.

## Hari 3 · Eksperimen kegagalan (2 jam) — WAJIB cara kembali normal

> Aturan: tiap eksperimen punya langkah pemulihan. Jangan tinggalin sistem dalam kondisi rusak.

### Langkah 3.1 — Graceful shutdown: 0 request gagal saat SIGTERM

```bash
# Sesi A: bombardir endpoint
hey -z 30s -c 50 http://localhost/orders &
# Sesi B: kirim SIGTERM
docker compose stop api
# hitung 5xx
docker compose logs api 2>&1 | grep -c 'level":"error"' || true
```

**Yang harus terjadi:** hampir nol 5xx; container exit bersih setelah request selesai.
**Kalau gagal:** banyak 5xx → timeout `srv.Shutdown` > `stop_grace_period`; kecilkan.
**Catat di notes.md:** berapa lama dari SIGTERM sampai proses exit (lihat log `shutdown selesai`).

### Langkah 3.2 — Simulasi dependency mati (Redis down)

```bash
docker compose stop redis
sleep 3
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
curl -s -o /dev/null -w "healthz=%{http_code}\n" http://localhost/health
```

**Yang harus terjadi:** `readyz=503`, tapi `healthz=200` (proses tetap hidup, cuma lepas dari LB).
**Kalau gagal:** kalau `healthz` ikut 503 → berarti lo salah taruh cek dependency di `/healthz`.
**Catat di notes.md:** beda respon ini penting — bukti probe terpisah.

**Cara kembali normal (wajib):**

```bash
docker compose start redis
sleep 3
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz   # harus 200 lagi
```

### Langkah 3.3 — Simulasi startup gagal (password DB salah)

```bash
# di .env, ubah POSTGRES_PASSWORD jadi salah, lalu restart api
sed -i.bak 's/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=salah123/' .env
docker compose up -d api
sleep 5
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
docker compose ps api
```

**Yang harus terjadi:** `readyz=503`, dan container TIDAK pernah dibilang healthy (status `unhealthy`/`starting`).
**Kalau gagal:** kalau malah `healthy` → aplikasi nggak ngecek DB beneran; periksa `handleReadyz`.
**Catat di notes.md:** bukti readiness nggak bohong.

**Cara kembali normal (wajib):**

```bash
mv .env.bak .env
docker compose up -d api
sleep 5
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz   # 200
```

### Langkah 3.4 — Rollback ke digest sebelumnya

```bash
# lihat digest yang lagi jalan
docker inspect --format='{{index .RepoDigests 0}}' production-reliability-lab-api
# asumsi versi sebelumnya ada di GHCR:
docker pull ghcr.io/<owner>/<repo>/sre-api@sha256:<digest-lama>
# ganti image di compose, lalu:
docker compose up -d --wait
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
```

**Yang harus terjadi:** sistem balik ke versi lama, `readyz=200`.
**Kalau gagal:** digest lama nggak ada → retensi image belum diatur (temuan).
**Catat di notes.md:** perintah rollback final.

**Cara kembali normal (wajib):** pastikan compose menunjuk digest/versi yang benar, `docker compose up -d --wait`, verifikasi `readyz=200`.

## Hari 4 · Review & dokumentasi (2 jam)

### Langkah 4.1 — Isi `docs/deployment.md` (sudah ada, lengkapi)

Pastikan checklist sebelum/sesudah deploy + tabel kriteria rollback lengkap.

**Yang harus terjadi:** dokumen menjawab "gimana deploy versi tertentu" + "kapan rollback".
**Kalau gagal:** lihat lagi `bahan-bacaan.md` bagian 2.6.

### Langkah 4.2 — Lengkapi `commands.md`

Salin perintah Hari 2 & 3 ke `commands.md` (format: perintah → output → interpretasi).

**Yang harus terjadi:** minimal 15 perintah.
**Kalau gagal:** lo kurang nyatet → balik ke Hari 2/3.

### Langkah 4.3 — Runbook kegagalan

Tulis di `runbooks/README.md`: (1) Redis mati → readyz 503; (2) password DB salah → startup gagal; (3) rollback digagalkan karena digest hilang.

**Yang harus terjadi:** tiap skenario punya gejala → bukti → pemulihan.
**Kalau gagal:** contohin dari `docs/deployment.md`.

### Langkah 4.4 — Acceptance test ulang & commit

```bash
make down && make up
curl -fsS http://localhost/health && curl -fsS http://localhost/readyz
git add labs/week-06-deployment docs/deployment.md .env.enc
git commit -m "week 06: deployment, graceful shutdown, SOPS"
```

**Yang harus terjadi:** health & readyz 200, commit masuk, README Referensi terisi.
**Kalau gagal:** jangan `--force` push.

## Daftar periksa "minggu ini selesai"

- [ ] `app/main.go` berhenti graceful saat SIGTERM; 0 (atau ~0) 5xx saat restart.
- [ ] `/readyz` 503 saat Redis/DB mati, `/healthz` tetap 200.
- [ ] Startup gagal (password salah) → container nggak pernah healthy.
- [ ] `.env.enc` terenkripsi; dekripsi round-trip sukses.
- [ ] Rollback by digest diuji (atau retensi image dicatat sebagai temuan).
- [ ] `docs/deployment.md` punya kriteria rollback objektif.
- [ ] `commands.md` ≥ 15 perintah; README Referensi terisi.

## Cara membuktikan ke reviewer

1. `docker compose logs api` saat SIGTERM → tunjuk `shutdown selesai` tanpa error 5xx.
2. `readyz=503` + `healthz=200` pas Redis mati (screenshot/log).
3. `.env.enc` isinya terenkripsi, bukan plaintext.
4. `docs/deployment.md` → tabel kriteria rollback.
