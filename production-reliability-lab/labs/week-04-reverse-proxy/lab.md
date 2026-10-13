# Lab — Minggu 4 · DNS, TLS, dan Reverse Proxy

> Ikuti **berurutan**, jangan lompat. Setiap langkah: perintah → **Yang harus terjadi** → **Kalau gagal**.

## Persiapan (10 menit)
- [ ] Docker + Docker Compose v2 jalan (`docker compose version`).
- [ ] Stack minggu 3 jalan: `docker compose up -d --build && docker compose ps` → `api`, `postgres`, `redis` status `healthy`/`running`.
- [ ] `nginx/default.conf` sudah ada (lihat repo).

## Hari 1 · Teori (2 jam)
### Langkah 1.1 — Pahami posisi proxy
Baca `nginx/default.conf` di repo.
**Yang harus terjadi:** lo paham tiap blok: `upstream api`, `location /`, `proxy_pass`, dan 3 timeout.
**Kalau gagal:** balik ke `bahan-bacaan.md` bagian 2.1–2.4.
**Catat di notes.md:** bedanya forward vs reverse proxy.

### Langkah 1.2 — Baca dokumentasi primer
Buka [NGINX Beginner's Guide](https://nginx.org/en/docs/beginners_guide.html) & [ngx_http_proxy_module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html).
**Yang harus terjadi:** lo tahu fungsi `proxy_set_header`, `proxy_next_upstream`, dan 3 timeout.
**Catat di notes.md:** tabel timeout default (60s) vs nilai lab (5s/30s/30s).

## Hari 2 · Praktik (2 jam)
### Langkah 2.1 — Tambah Nginx ke compose & jalankan
`docker compose up -d && curl -v http://localhost/health`
**Yang harus terjadi:** balasan `200 OK` dari proxy di port 80.
**Kalau gagal:** `docker compose exec nginx nginx -t` → perbaiki syntax; `docker compose logs nginx`.

### Langkah 2.2 — Verifikasi header forwarding
`curl -s http://localhost/health -i | head` dan cek log aplikasi.
**Yang harus terjadi:** aplikasi lihat `X-Forwarded-Proto`, `X-Forwarded-For`, `Host`.
**Kalau gagal:** header nggak muncul → cek `proxy_set_header` di `default.conf`.

### Langkah 2.3 — Buktikan port DB tertutup
`ss -lntp | grep -E '5432|6379'` (di host) → harus **kosong**.
`nmap localhost` → cuma port 80 (dan 22) yang kebuka.
**Yang harus terjadi:** 5432/6379 nggak listen di host.
**Kalau gagal:** ada `ports:` di `postgres`/`redis` di compose → hapus.

## Hari 3 · Eksperimen kegagalan (2 jam) — FAILURE MATRIX
### Langkah 3.1 — 502 (upstream mati)
`docker compose stop api && curl -sS -o /dev/null -w '%{http_code}\n' http://localhost/health`
**Yang harus terjadi:** `502`. `docker compose exec nginx tail -5 /var/log/nginx/error.log` → `connect() failed (111: Connection refused)`.
**Kembali normal:** `docker compose start api`.

### Langkah 3.2 — 504 (upstream lambat)
Buat endpoint lambat di app (atau `docker compose exec api sh -c '...'`), lalu:
`curl -sS -o /dev/null -w '%{http_code}\n' 'http://localhost/slow?ms=60000'`
**Yang harus terjadi:** `504` setelah 30 detik. `error.log` → `upstream timed out`.
**Kembali normal:** hentikan request lambat.

### Langkah 3.3 — 499 (client kabur duluan)
`curl --max-time 2 -sS -o /dev/null -w '%{http_code}\n' 'http://localhost/slow?ms=60000'`
**Yang harus terjadi:** curl berhenti di 2 detik; `access.log` Nginx catat status `499`.
**Kembali normal:** selesai.

### Langkah 3.4 — TLS handshake gagal (bonus)
`curl --tls-max 1.1 https://localhost` (kalau TLS sudah dipasang).
**Yang harus terjadi:** gagal handshake karena server cuma terima TLS 1.2/1.3.

## Hari 4 · Review & dokumentasi (2 jam)
### Langkah 4.1 — Buat sertifikat lokal
`mkcert -install && mkcert app.local 127.0.0.1 ::1`
**Yang harus terjadi:** file `.pem` + `.key` dibuat; CA lokal masuk trust store.
**Kalau gagal:** `mkcert` belum install → `brew install mkcert` / `apt install mkcert`.

### Langkah 4.2 — Update diagram & failure matrix
Salin tabel failure matrix (502/504/499/TLS) ke `evidence/` lengkap dengan log asli.
**Yang harus terjadi:** tiap kode punya bukti log.

### Langkah 4.3 — Commit
`git add labs/week-04-reverse-proxy && git commit -m "week4: reverse proxy + failure matrix"`

## Daftar periksa "minggu ini selesai"
- [ ] API cuma bisa diakses lewat proxy (port 80).
- [ ] 5432/6379 nggak kebuka di host.
- [ ] 502/504/499 direproduksi & dijelaskan dari log.
- [ ] TLS lokal (`mkcert`) jalan & bisa diverifikasi.
- [ ] Header forwarding terbukti sampai ke aplikasi.

## Cara membuktikan ke reviewer
```bash
docker compose exec nginx nginx -t
curl -sS -o /dev/null -w 'proxy=%{http_code}\n' http://localhost/health
docker compose stop api && curl -sS -o /dev/null -w '502=%{http_code}\n' http://localhost/health
docker compose start api
ss -lntp | grep -E '5432|6379'   # harus kosong
```
