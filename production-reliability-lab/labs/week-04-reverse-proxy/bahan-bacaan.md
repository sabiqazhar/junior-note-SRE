# Minggu 4 · DNS, TLS, dan Reverse Proxy — Bahan Bacaan

> Dokumen ini **berdiri sendiri**. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham konsepnya. Baru setelah itu kerjakan `lab.md`.

## 1. Kenapa minggu ini penting

Bayangin lo bikin API yang jalan sempurna, tapi besok pagi user komplain "web-nya error semua". Lo cek aplikasi — normal. Lo cek database — normal. Ternyata yang rusak adalah Nginx di depannya yang nolak koneksi karena sertifikat TLS kedaluwarsa. Atau proxy yang nggak bisa nyambung ke backend karena `proxy_read_timeout` kepanjangan.

Di minggu 2 lo belajar request sampai ke aplikasi. Di minggu 4 lo belajar apa yang terjadi **tepat di depan** aplikasi: lapisan yang paling sering jadi penyebab "tapi di lokal kan jalan" — padahal di lokal lo nggak pakai proxy. Reverse proxy, TLS, dan timeout adalah tempat lahirnya 80% keluhan "lambat" dan "error aneh", dan kebanyakan orang salah diagnosa karena nggak tahu bedanya **502** sama **504**. Minggu ini lo belajar membedakannya pakai bukti dari log, bukan tebakan.

## 2. Konsep inti

### 2.1 Reverse proxy bedanya sama forward proxy
**Forward proxy** = perantara buat *client* (lo akses internet lewat dia). **Reverse proxy** = perantara buat *server*: client nggak tahu ada banyak backend di belakangnya. Nginx di sini nerima request dari publik, lalu neruskannya ke `api:8080` (container internal). Keuntungannya: lo cuma buka **SATU** port ke publik, backend nggak kelihatan ke luar, dan lo bisa pasang TLS, limit, dan log di satu titik.

### 2.2 Upstream dan `proxy_next_upstream`
`upstream` adalah daftar backend yang dilayani proxy. Nginx bisa "pindah" ke backend lain kalau satu backend balas error/timeout lewat direktif `proxy_next_upstream error timeout http_502`. Di lab kita cuma punya satu backend, tapi polanya identik kalau nanti lo punya 5 replica.

### 2.3 Proxy headers — jangan lupa ini
Waktu Nginx neruskkan request, backend butuh tahu:
- client aslinya dari IP mana → `X-Real-IP` dan `X-Forwarded-For`
- pakai http atau https → `X-Forwarded-Proto`
- nama host asli → `Host`

Lupa set `Host $host` → backend salah milih virtual host. Lupa `X-Forwarded-Proto` → aplikasi ngerasa di http terus, bisa bikin redirect loop ke https. Ini jebakan klasik yang nggak kelihatan sampai lo pasang TLS.

### 2.4 Timeout — sumber 502 / 504 / 499
Nginx punya tiga timeout penting:
- `proxy_connect_timeout` — waktu buat **nyambung** ke backend
- `proxy_read_timeout` — waktu **nunggu** backend balas
- `proxy_send_timeout` — waktu nunggu backend nerima body

Default ketiganya **60 detik** — kepanjangan buat API modern. Turunin. **Aturan main:** timeout di *client* (mis. `curl --max-time`) harus **lebih kecil** dari `proxy_read_timeout`. Kalau sebaliknya, lo dapet **499** (client kabur duluan), bukan 504.

### 2.5 Keep-alive upstream
Nginx bisa pakai ulang koneksi TCP ke backend biar nggak buka baru tiap request. Caranya: `keepalive 32` di blok `upstream` + `proxy_http_version 1.1` + `proxy_set_header Connection ""`. Mulai **nginx 1.29.7** ini default ON, tapi tetap tulis eksplisit biar aman di versi lama dan gampang dibaca orang lain.

### 2.6 TLS dan sertifikat (perubahan 2025 yang penting)
TLS = enskripsi lapisan transport antara client dan proxy. Di production, sertifikat datang dari CA (Let's Encrypt lewat protokol ACME). **Yang berubah 2025:** Let's Encrypt mulai terbitkan sertifikat umur **6 hari** (awal 2026 GA penuh). Artinya renewal **harus otomatis harian** — cron bulanan = outage berkala. Buat lab, pakai `mkcert` (bikin CA lokal di mesin lo sendiri) atau `step-ca`. **Jangan** pakai sertifikat lab buat domain publik — cuma dipercaya di mesin lo.

### 2.7 Failure matrix: 502 vs 504 vs 499
- **502 Bad Gateway** → Nginx nyambung tapi upstream *nolak/reset* (backend mati, port tutup).
- **504 Gateway Timeout** → `proxy_read_timeout` habis, backend *kelamaan* mikir.
- **499** → kode *internal* Nginx, **bukan** standar HTTP. Artinya client nutup koneksi duluan (biasanya client timeout < proxy timeout).

## 3. Mekanisme di balik layar

Waktu lo jalanin `curl http://localhost/health`:

1. Client buka koneksi TCP ke **port 80** (cuma port ini yang kebuka ke host).
2. Nginx terima, cocokkan `server_name`, masuk `location /`.
3. Nginx ambil koneksi dari pool keep-alive ke `api:8080`, tambahin header `X-Forwarded-*`.
4. Backend proses, balikin response.
5. Nginx tulis `access.log` (status, `request_time`, `upstream_response_time`).
6. Kalau backend **mati** → Nginx balik **502**, `error.log`: `connect() failed (111: Connection refused)`.
7. Kalau backend **lambat** lewat 30 detik → **504**, `error.log`: `upstream timed out`.

## 4. Glosarium mini

| Istilah | Arti dengan bahasa sederhana | Contoh konkret |
| --- | --- | --- |
| Reverse proxy | Penghubung di depan server; client nggak lihat backend | Nginx di depan Go API |
| Upstream | Daftar backend yang dilayani proxy | `upstream api { server api:8080; }` |
| 502 | Backend nolak/reset koneksi | Container `api` mati |
| 504 | Backend kelamaan sampai proxy timeout | Endpoint `/slow?ms=60000` |
| 499 | Client nutup koneksi duluan (kode Nginx) | `curl --max-time 2` ke endpoint lambat |
| TLS | Enskripsi transport (https) | Sertifikat Let's Encrypt |
| Keep-alive | Pakai ulang koneksi TCP biar nggak buka baru | `keepalive 32` |
| X-Forwarded-For | Jejak IP asli client lewat proxy | `203.0.113.5` di log backend |

## 5. Bahan bacaan wajib

| # | Sumber | Apa yang dibaca / dicari | Estimasi |
| --- | --- | --- | --- |
| 1 | [NGINX Beginner's Guide](https://nginx.org/en/docs/beginners_guide.html) | Konsep `server`, `location`, `proxy_pass` | 20 mnt |
| 2 | [ngx_http_proxy_module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html) | Timeout, `proxy_set_header`, `proxy_next_upstream` | 30 mnt |
| 3 | [mkcert (GitHub)](https://github.com/FiloSottile/mkcert) | Cara bikin sertifikat lokal tepercaya | 10 mnt |
| 4 | [step-ca (smallstep)](https://github.com/smallstep/certificates) | CA lokal yang lebih serius buat lab | 15 mnt |
| 5 | [Let's Encrypt](https://letsencrypt.org/) | Kenapa sertifikat sekarang umur 6 hari | 10 mnt |
| 6 | [Caddy docs](https://caddyserver.com/docs/) | Alternatif proxy "HTTPS otomatis" | 15 mnt |

## 6. Self-check (harus bisa dijawab tanpa buka Google)

1. Apa bedanya reverse proxy sama forward proxy?
2. Kenapa `X-Forwarded-Proto` penting kalau lo pasang TLS?
3. Kapan lo dapet 502, kapan 504, kapan 499?
4. Kenapa `proxy_read_timeout` 60 detik itu terlalu lama buat API?
5. Apa fungsi `keepalive` di blok `upstream`?
6. Kenapa sertifikat 6 hari butuh renewal otomatis harian?

## 7. Yang sering salah paham

| Mitos | Kenyataan |
| --- | --- |
| "502 sama 504 kan sama-sama gateway error" | 502 = upstream **nolak**, 504 = upstream **kelamaan**. Penanganannya beda. |
| "499 itu error server" | Bukan. 499 adalah kode internal Nginx: **client** yang nutup duluan. |
| "Set ProxyPass sudah cukup" | Tanpa `Host`/`X-Forwarded-*` backend bisa salah behavior, apalagi kalau ada redirect. |
| "Sertifikat cuma perlu dibikin sekali" | 2025: umur 6 hari. Tanpa auto-renewal harian = outage berkala. |
| "Timeout default aman" | Default 60s kepanjangan; client biasanya timeout duluan → 499 misterius. |
