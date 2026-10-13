# Ringkasan — Minggu 02 · Networking untuk backend engineer

## Inti minggu ini dalam 5 kalimat

1. Satu HTTP request itu perjalanan berfase: **DNS → TCP → TLS → HTTP → transfer body**, dan tiap fase punya angka sendiri.
2. `curl -w` bisa cetak semua fase itu, jadi lo bisa nunjuk persis di fase mana request lemot — bukan nebak.
3. Tiga kegagalan klasik punya "sidik jari" beda: **exit 6** (DNS), **exit 7** (refused/port), **exit 28** (timeout/firewall).
4. `dig` nggak selalu sama dengan apa yang dipakai aplikasi — `getent hosts` dan resolver internal Docker lebih jujur.
5. OpenSSL 3.5 LTS (rilis 8 Apr 2025, support sampai 2030) adalah versi yang lo temuin di distro baru; TLS 1.3 + SNI wajib lo pahami buat debug sertifikat.

## Peta konsep (daftar bertingkat)

- Request flow
  - DNS (nama → IP)
    - resolver publik: `dig`
    - resolver lokal/aplikasi: `getent hosts`, `systemd-resolved`, stub `127.0.0.53`
    - DNS internal Docker Compose: `api` ↔ `redis`
  - TCP handshake (SYN/SYN-ACK/ACK)
    - listen socket: `ss -lntp`
    - refused vs timeout
  - TLS handshake
    - versi (TLS 1.3), cipher, SNI
    - `openssl s_client -servername`
  - HTTP (1.1 / 2 / 3)
    - HTTP/3 = QUIC di UDP/443 (RFC 9114, ~21% traffic 2025)
  - Transfer body
- Latency phases (`curl -w`)
  - `time_namelookup` (DNS)
  - `time_connect` (TCP)
  - `time_appconnect` (TLS)
  - `time_starttransfer` / TTFB (server mikir + header)
  - `time_total` (transfer body)
- Failure signature
  - exit 6 = DNS failure
  - exit 7 = connection refused
  - exit 28 = timeout (DROP / overload)
  - exit 0 + 5xx = aplikasi respond tapi jelek
- Firewall
  - REJECT → refused
  - DROP → timeout

## Perintah kunci + yang harus lo lihat di output

| Perintah | Output kunci | Interpretasi |
| --- | --- | --- |
| `ip addr` | `inet 192.168.x.x` per interface | Alamat IP & interface aktif |
| `ip route` | `default via 192.168.x.1` | Ke mana traffic dilempar |
| `ss -lntp` | `LISTEN ... :8080 users:(("sre-api",pid=9001))` | Port listen + pemilik proses |
| `dig example.com +trace` | baris `A` dari root→TLD→authoritative | Jalur resolusi DNS lengkap |
| `dig +short example.com` | `93.184.216.34` | Jawaban DNS mentah |
| `getent hosts redis` | `172.20.0.3 redis` | Resolusi ala aplikasi (Docker) |
| `curl -w 'dns=%{time_namelookup} ...' https://...` | `dns=0.04 tcp=0.06 tls=0.09 ttfb=0.25 total=0.26` | Fase latency per request |
| `curl --connect-timeout 3 http://localhost:9999` | `curl: (7) Connection refused` | Port nggak listen |
| `openssl s_client -connect host:443 -servername host -tls1_3 -alpn h2 -brief` | `Verify return code: 0 (ok)` + `TLSv1.3` | Versi TLS & ALPN (h2) yang dipakai |
| `curl -sS -o /dev/null -w '%{http_code} %{exitcode}\n' URL` | `200 0` / `000 7` / `000 28` | Kode HTTP + exit code curl |

## Failure matrix / decision tree

```
Request gagal
├─ curl exit 6   → DNS. Cek: dig, getent hosts, /etc/resolv.conf, resolver internal Docker
├─ curl exit 7   → Nggak ada yang listen. Cek: ss -lntp, proses hidup?, port bener?
├─ curl exit 28  → Timeout. Cek: firewall/SecurityGroup (DROP?), routing, service overload
└─ exit 0 tapi 5xx → Aplikasi. Cek: log aplikasi, dependency health, trace
```

| Kondisi | Error | exit code | Cara deteksi |
| --- | --- | --- | --- |
| DNS gagal | `Could not resolve host` | 6 | `dig`, `getent hosts` |
| Port nggak listen | `Connection refused` | 7 | `ss -lntp` |
| Firewall drop / routing | hang lalu `Timeout` | 28 | `mtr`, `tcpdump`, cek SG |
| Aplikasi error | HTTP 500 | 0 | log aplikasi |

## Jebakan umum

- **TTFB itu bukan waktu jaringan** — udah termasuk server mikir. `total - starttransfer` baru waktu kirim body.
- **Hasil `dig` belum tentu yang dipakai aplikasi.** Ada cache + `systemd-resolved` (stub `127.0.0.53`). Pakai `getent hosts`.
- **Ping diblok ≠ service mati.** ICMP ≠ TCP.
- **`openssl s_client` tanpa `-servername`** bisa dapet sertifikat salah (SNI).
- **DNS Docker itu internal.** `api` resolve `redis`, laptop lo nggak.
- **Lupa cek `$?`** setelah curl — exit code itu informasi, bukan gangguan.

## Acceptance criteria

1. Bisa jelasin beda **DNS failure, TCP failure, HTTP error, timeout** — dan mereproduksi keempatnya.
2. Bisa nemuin port yang listen **dan** proses pemiliknya (`ss -lntp`).
3. Bisa ukur dan jelasin 5 fase latency satu request.
4. Bisa buktiin versi TLS & protokol (h2/h3) dengan `openssl s_client`.

## Satu kalimat yang harus lo ingat

> Sebelum nyalahin server, pecah request jadi DNS→TCP→TLS→TTFB→total; angkanya bakal nunjukin di fase mana masalahnya sebenarnya.
