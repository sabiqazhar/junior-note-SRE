# Ringkasan — Minggu 4 · DNS, TLS, dan Reverse Proxy

> Satu halaman buat dibaca ulang sebelum wawancara/ujian. Bukan pengganti `bahan-bacaan.md`.

## Inti minggu ini dalam 5 kalimat
1. Reverse proxy (Nginx) duduk di depan aplikasi; cuma dia yang boleh kebuka ke publik.
2. Tiga timeout (`connect`/`read`/`send`) default 60s — turunin, dan pastikan client timeout < `proxy_read_timeout`.
3. Header `Host` + `X-Forwarded-*` wajib diset biar backend nggak salah behavior.
4. **502** = upstream nolak, **504** = upstream kelamaan, **499** = client kabur duluan (bukan standar HTTP).
5. TLS 2025 = sertifikat umur 6 hari → renewal otomatis harian wajib (lab pakai `mkcert`).

## Peta konsep
- Reverse proxy
  - upstream (daftar backend) + `proxy_next_upstream`
  - proxy headers: `Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`
  - timeout: `connect` / `read` / `send`
  - keep-alive: `keepalive` + `proxy_http_version 1.1` + `Connection ""`
- TLS
  - sertifikat CA, ACME, auto-renewal
  - lab: `mkcert` / `step-ca` (jangan dipakai domain publik)
- Failure matrix: 502 / 504 / 499

## Perintah kunci + yang harus lo lihat di output
| Perintah | Yang dicari di output | Artinya kalau beda |
| --- | --- | --- |
| `docker compose exec nginx nginx -t` | `configuration test is successful` | Config salah → reload gagal |
| `curl -sS -o /dev/null -w '%{http_code}\n' http://localhost/health` | `200` | Bukan 200 → cek proxy/upstream |
| `docker compose logs nginx` | akses/error log | Error handshake/timeout |
| `openssl s_client -connect localhost:443` | `Verify return code: 0` | Sertifikat nggak tepercaya |

## Failure matrix
| Kode | Arti | Cara bikin di lab | Di mana cek |
| --- | --- | --- | --- |
| 502 | Upstream nolak/reset | `docker compose stop api` | `error.log`: `connect() failed (111...)` |
| 504 | `proxy_read_timeout` habis | endpoint `/slow?ms=60000` | `error.log`: `upstream timed out` |
| 499 | Client tutup duluan | `curl --max-time 2` ke endpoint lambat | `access.log` kolom status |
| TLS fail | Versi/cipher nggak cocok | `curl --tls-max 1.1` | `error.log`: `SSL_do_handshake() failed` |

## Jebakan umum
- 502 dan 504 sering ketuker.
- Lupa `proxy_set_header Host $host` → backend salah vhost.
- `proxy_read_timeout` > client timeout → hasilnya 499, bukan 504.
- Sertifikat 6 hari tanpa auto-renewal = outage berkala.

## Acceptance criteria
1. Client akses API cuma lewat reverse proxy.
2. Port Postgres & Redis nggak terekspos ke host (`ss -lntp`, `nmap`).
3. Kegagalan upstream menghasilkan error yang bisa dilacak dari log.
4. Bisa jelasin & reproduksi 502 / 504 / 499.

## Satu kalimat yang harus lo ingat
> 502 = upstream nolak, 504 = upstream kelamaan, 499 = client yang kabur — baca `error.log` buat membedakannya.
