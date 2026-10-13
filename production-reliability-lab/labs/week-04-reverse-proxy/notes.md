# Catatan konsep — Week 04 · DNS, TLS, dan Reverse Proxy

> Tulis ulang dengan bahasa lo sendiri di bawah tiap poin. Kalau nggak bisa jelasin, berarti belum paham.

## Reverse proxy vs forward proxy
- Forward proxy = perantara buat **client** (akses internet lewat dia).
- Reverse proxy = perantara buat **server** (client nggak lihat backend).
- Di lab: Nginx di depan Go API, cuma port 80 yang kebuka ke host.

## Proxy headers wajib
- `Host $host` — biar backend nggak salah pilih virtual host.
- `X-Real-IP` / `X-Forwarded-For` — jejak IP asli client.
- `X-Forwarded-Proto` — http/https asli; lupa ini = redirect loop pas TLS nyala.

## Timeout (default nginx = 60s, kepanjangan)
- `proxy_connect_timeout 5s` — waktu nyambung ke backend.
- `proxy_read_timeout 30s` — waktu nunggu backend balas.
- `proxy_send_timeout 30s` — waktu nunggu backend terima body.
- Aturan: client timeout < `proxy_read_timeout`, kalau nggak → 499.

## Keep-alive upstream
- `keepalive 32` + `proxy_http_version 1.1` + `proxy_set_header Connection ""`.
- Default ON sejak nginx 1.29.7, tapi tulis eksplisit.

## TLS 2025
- Let's Encrypt umur sertifikat **6 hari** → renewal otomatis harian wajib.
- Lab: `mkcert` (CA lokal) atau `step-ca`. Jangan pakai domain publik.

## Failure matrix (hafal!)
| Kode | Penyebab | Cek di |
| --- | --- | --- |
| 502 | upstream nolak/reset | `error.log`: `connect() failed (111...)` |
| 504 | `proxy_read_timeout` habis | `error.log`: `upstream timed out` |
| 499 | client tutup duluan (bukan standar HTTP) | `access.log` status |

## Self-check (jawab tanpa Google)
1. Apa bedanya reverse proxy dan forward proxy?
2. Kenapa `X-Forwarded-Proto` penting saat TLS nyala?
3. Kapan 502, kapan 504, kapan 499?
4. Kenapa timeout default 60s kepanjangan?
5. Fungsi `keepalive` di upstream?
6. Kenapa sertifikat 6 hari butuh renewal otomatis harian?
