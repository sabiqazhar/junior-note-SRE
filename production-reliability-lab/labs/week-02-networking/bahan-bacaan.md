# Minggu 02 · Networking untuk backend engineer — Bahan Bacaan

> Dokumen ini berdiri sendiri. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham konsepnya. Baru setelah itu kerjakan lab.md.

---

## 1. Kenapa minggu ini penting

Bayangin lo dapet tiket: *"API kita lemot, tapi cuma buat user di region tertentu."* Lo buka Grafana, lihat latency p99 naik. Terus lo bingung: ini gara-gara kode lo, gara-gara database, atau gara-gara jaringan? Tanpa paham networking, debatnya cuma nebak. Dan nebak di produksi itu mahal: lo bisa salah naikin instance padahal masalahnya di DNS resolver yang timeout, atau salah nyalahin "jaringan lab" padahal sebenarnya aplikasi lo yang balas 500.

Minggu ini ngajarin satu skill: **bisa memecah satu request jadi fase-fase, dan nunjukin angka di tiap fase**. Request itu kayak kurir yang harus lewat beberapa pos: cari alamat (DNS), ketuk pintu (TCP), saling kenal & bikin kunci (TLS), baru ngobrol (HTTP), lalu antar barangnya (transfer body). Kalau lo bisa bilang "DNS 40ms, TCP 2ms, TLS 30ms, TTFB 200ms", debat "lelet di server atau jaringan?" selesai dalam 10 detik.

Ini penting karena hampir semua layer di atas (load balancer, service mesh, CDN) cuma menambah fase ke perjalanan kurir ini. Kalau lo nggak paham fase dasarnya, lo nggak akan paham fase tambahannya.

---

## 2. Konsep inti

### 2.1 DNS: dari nama jadi alamat IP

DNS itu buku telepon internet. `api.example.com` nggak bisa dikirimin paket langsung — harus jadi `93.184.216.34`. Resolusi itu sendiri butuh waktu dan bisa gagal.

Ada dua cara ngecek yang sering disamain padahal beda:
- `dig example.com` → nanya resolver publik, jujur soal apa kata DNS authority.
- `getent hosts redis` → nanya resolver **lokal** aplikasi lo (bisa `systemd-resolved`, `/etc/hosts`, stub `127.0.0.53`). Ini yang dipakai program lo beneran.

Di Docker Compose, `api` bisa resolve `redis` lewat DNS internal — tapi laptop lo nggak bisa. Ini sumber 90% kebingungan pemula.

### 2.2 TCP handshake: ketuk pintu

Sebelum ngobrol, harus ada koneksi. TCP pakai 3-way handshake: SYN → SYN-ACK → ACK. Kalau `ss -lntp` nggak nemu yang listen di port tujuan, hasilnya **connection refused** (pintu dikunci). Kalau paket nyampe tapi nggak ada balasan (firewall DROP), hasilnya **timeout** (nggak ada yang nyahut).

### 2.3 TLS handshake: saling kenal & bikin kunci

Buat HTTPS, setelah TCP nyambung, client & server nego versi TLS, cipher, dan tukar sertifikat. Di sini SNI (`-servername`) penting: server bisa punya banyak sertifikat, dan客家 client harus bilang "aku mau yang buat domain apa".

Fakta 2025: **OpenSSL 3.5 LTS** rilis 8 April 2025 dan didukung sampai 2030. Versi ini yang bakal lo temuin di distro baru. Make sure lo ngetes dengan `openssl` yang support TLS 1.3.

### 2.4 HTTP/1.1 vs HTTP/2 vs HTTP/3

- **HTTP/1.1** — satu request per koneksi (keep-alive bikin beberapa, tapi serial).
- **HTTP/2** — multiplexing di satu koneksi, ~50% traffic global 2025 (Cloudflare Radar).
- **HTTP/3** — di atas QUIC (UDP), RFC 9114. Bukan eksperimen lagi: ~21% traffic global 2025. Butuh port UDP/443 terbuka.

Lo nggak wajib kuasai QUIC minggu ini, tapi lo harus tau HTTP/3 itu ada dan beda transport.

### 2.5 Lima fase latency satu request

Ini yang paling berguna. `curl -w` bisa cetak tiap fase:

| Variabel curl | Fase | Arti |
| --- | --- | --- |
| `time_namelookup` | DNS | Waktu sampai dapet IP |
| `time_connect` | TCP | + waktu sampai TCP handshake selesai |
| `time_appconnect` | TLS | + waktu sampai TLS handshake selesai |
| `time_starttransfer` (TTFB) | Server mikir + kirim header | Waktu sampai byte pertama balik |
| `time_total` | Transfer body | Waktu total sampai selesai |

Rumus praktis: `ttfb - appconnect` = waktu server mikir. `total - starttransfer` = waktu kirim body. Jangan campur.

### 2.6 Listening socket & state

`ss -lntp` nunjukin port yang **listen** + proses pemiliknya. Tanpa `-p` lo cuma lihat port, bukan siapa. Ini hakim waktu debat "port bentrok atau nggak".

### 2.7 Connection timeout vs application timeout

Dua timeout beda:
- **Connection timeout** — "berapa lama nunggu pintu kebuka". Di `curl` itu `--connect-timeout`.
- **Read/operation timeout** — "pintu udah kebuka, tapi nggak ada yang ngomong". Di `curl` itu `-m`/`--max-time`.

Di nginx nanti (minggu 4) ini jadi `proxy_connect_timeout` vs `proxy_read_timeout`. Bedain sejak sekarang.

### 2.8 DROP vs REJECT (firewall)

- **REJECT** → balikin "port unreachable" → client langsung dapet refused.
- **DROP** → diam aja → client nunggu sampai timeout.

Ini kenapa "timeout" sering kali bukan aplikasi lemot, tapi firewall/sg yang ngeblokir silently.

---

## 3. Mekanisme di balik layar: satu request dari client ke DB

Urutan sebenarnya (buat lo jelasin insiden dengan percaya diri):

1. **DNS** — client resolve `api.example.com`. Kalau gagal di sini, nggak ada paket sama sekali; `curl` exit 6 (`Could not resolve host`).
2. **Routing** — kernel cari jalan ke IP tujuan lewat `ip route`. Salah route = paket nyasar.
3. **TCP handshake** — SYN ke port 8080. Kalau nggak ada yang listen → `connection refused` (exit 7). Kalau di-DROP → tunggu sampai `--connect-timeout` habis (exit 28).
4. **TLS** (kalau HTTPS) — nego versi & cipher, verifikasi sertifikat. Gagal → `SSL_ERROR` / `curl` exit 60.
5. **HTTP** — client kirim request. Server (nginx → Go API) proses, tanya Postgres/Redis.
6. **Transfer body** — response dikirim balik. Di sinilah `total - starttransfer` dihabiskan.
7. **Tutup koneksi** — keep-alive bisa dipakai ulang (mulai nginx 1.29.7 default ON, lihat minggu 4).

Tiap nomor di atas punya angka sendiri di `curl -w`. Kalau step 1 lambat, langkah 2–7 nggak akan dibuka.

---

## 4. Glosarium mini

| Istilah | Arti dengan bahasa sederhana | Contoh konkret |
| --- | --- | --- |
| **DNS** | Buku telepon yang ubah nama jadi IP | `dig api.example.com` → `93.184.216.34` |
| **TCP handshake** | Ketuk pintu 3 kali sebelum ngobrol | SYN → SYN-ACK → ACK |
| **TLS** | Protokol bikin koneksi terenkripsi & terverifikasi | HTTPS, butuh sertifikat valid |
| **SNI** | "Aku mau sertifikat buat domain apa" | `openssl s_client -servername api.example.com` |
| **TTFB** | Waktu sampai byte pertama balik (termasuk server mikir) | `time_starttransfer` di curl |
| **HTTP/3** | HTTP di atas QUIC (UDP), RFC 9114 | ~21% traffic global 2025 |
| **Connection refused** | Port nggak ada yang listen | `curl` exit 7 |
| **Timeout** | Nggak ada balasan (biasanya DROP/lemot) | `curl` exit 28 |
| **getent hosts** | Resolusi ala aplikasi, bukan ala dig | `getent hosts redis` di container |
| **DROP vs REJECT** | Diam vs tolak tegas di firewall | DROP → timeout; REJECT → refused |
| **OpenSSL 3.5** | LTS rilis 8 Apr 2025, support sampai 2030 | `openssl version` → 3.5.x |

---

## 5. Bahan bacaan wajib

| # | Sumber (link) | Bagian yang dibaca & apa yang dicari | Estimasi |
| --- | --- | --- | --- |
| 1 | [Cloudflare Learning Center — Network Layer](https://www.cloudflare.com/learning/network-layer/) | Baca bagian DNS, TCP/IP, dan OSI vs TCP/IP model. Cari pemahaman tiap lapisan tempat fase request berjalan. | 30 menit |
| 2 | [curl manpage — -w/--write-out](https://curl.se/docs/manpage.html) | Cari bagian `write-out` dan daftar `%{time_*}`, `%{http_version}`, `%{exitcode}`. Ini perintah paling sering lo pakai minggu ini. | 25 menit |
| 3 | [OpenSSL 3.5 Final Release](https://openssl-library.org/post/2025-04-08-openssl-35-final-release/) | Baca pengumuman rilis: kenapa 3.5 itu LTS dan didukung sampai 2030. Lo bakal pakai `openssl s_client` dari versi ini. | 15 menit |
| 4 | [RFC 9114 — HTTP/3](https://www.rfc-editor.org/rfc/rfc9114.html) | Cukup baca Introduction & "HTTP/3 over QUIC". Pahami beda transport UDP/QUIC vs TCP. | 30 menit |
| 5 | [Cloudflare Radar 2025 Year in Review](https://blog.cloudflare.com/radar-2025-year-in-review/) | Cari bagian protokol: HTTP/2 ~50%, HTTP/3 ~21% traffic global 2025. Bukti HTTP/3 bukan eksperimen. | 15 menit |
| 6 | [curl write-out explained (everything.curl.dev)](https://everything.curl.dev/usingcurl/verbose/writeout) | Penjelasan tiap variabel write-out dengan contoh. Melengkapi manpage di baris 2. | 20 menit |

---

## 6. Self-check (harus bisa dijawab tanpa buka Google)

1. `curl` ke domain mati kasih exit code berapa, dan di fase mana request gagal? Gimana cara lo bedain itu dari "port nggak listen"?
2. Kenapa `dig` bisa bilang IP benar, tapi aplikasi lo tetep "could not resolve host"? Perintah apa buat ngecek resolusi yang dipakai aplikasi beneran?
3. Apa bedanya `time_connect` dan `time_appconnect`? Di HTTPS, mana yang lebih gede secara umum?
4. `TTFB` tinggi (mis. 800ms) tapi `total` cuma 805ms. Apa kesimpulan lo soal server vs jaringan?
5. `curl` exit 28 (timeout) ke suatu host. Sebutkan 3 kemungkinan penyebab di layer jaringan (bukan aplikasi).
6. Kenapa `openssl s_client` tanpa `-servername` bisa dapet sertifikat yang salah?
7. Apa bedanya firewall REJECT dan DROP dari sudut pandang client, dan kenapa itu penting buat troubleshooting?

---

## 7. Yang sering salah paham

| Mitos | Kenyataannya |
| --- | --- |
| "Ping sukses berarti service hidup." | ICMP dan TCP jalan tol beda. Bisa aja ping dibalas tapi port 8080 nggak listen (atau sebaliknya). |
| "`dig` benar berarti aplikasi bisa connect." | Ada cache, `systemd-resolved`, stub `127.0.0.53`. `getent hosts` lebih jujur soal resolusi nyata aplikasi. |
| "TTFB itu waktu jaringan." | TTFB udah termasuk waktu server mikir. `total - starttransfer` baru waktu kirim body. |
| "Timeout pasti gara-gara server lemot." | Sering kali firewall DROP atau routing salah — client nunggu sampai timeout tanpa ada balasan. |
| "DNS di Docker sama dengan DNS di laptop." | DNS Compose itu internal; `api` resolve `redis`, tapi laptop lo nggak. Ini 90% kebingungan pemula. |
| "HTTP/3 masih eksperimen." | RFC 9114 stabil; ~21% traffic global 2025 menurut Cloudflare Radar. Sudah produksi. |
