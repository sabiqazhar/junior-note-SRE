# Week 02 · Networking untuk backend engineer

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai networking untuk backend engineer sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- DNS & resolusi
- TCP handshake & socket state
- TLS 1.3 & ALPN
- HTTP/1.1 vs HTTP/2 vs HTTP/3
- Fase latency request

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Bisa memecah latency jadi 5 fase pakai `curl -w` dan menjelaskan arti tiap angka.
2. Bisa membedakan DNS failure, connection refused, timeout, dan HTTP error — lengkap dengan cara mereproduksinya.
3. Bisa menemukan port yang listen dan proses pemiliknya (`ss -lntp`).
4. Bisa membuktikan versi TLS dan ALPN (h2/h3) dengan `openssl s_client`.

## Struktur folder

```
labs/week-02-networking/
├── README.md        <- file ini
├── notes.md         <- catatan konsep Hari 1
├── commands.md      <- semua perintah yang dijalankan + output yang disanitasi
├── runbooks/        <- cara memperbaiki kegagalan (Hari 4)
└── evidence/        <- bukti: screenshot, log, hasil pengukuran
```

## Aturan bukti

1. Setiap acceptance criteria di atas harus punya **bukti yang bisa direproduksi** (perintah + output).
2. Output yang mengandung IP internal, token, password, atau data pribadi **wajib disanitasi** sebelum di-commit.
3. Jangan tandai minggu ini selesai hanya karena tutorialnya sudah ditonton.

## Referensi

- [Cloudflare Learning Center — Network Layer](https://www.cloudflare.com/learning/network-layer/) — dasar DNS, TCP/IP, dan model OSI vs TCP/IP. Wajib sebelum paham fase request.
- [curl manpage — -w/--write-out](https://curl.se/docs/manpage.html) — daftar `%{time_*}`, `%{http_version}`, `%{exitcode}`. Perintah paling sering dipakai minggu ini.
- [curl write-out explained (everything.curl.dev)](https://everything.curl.dev/usingcurl/verbose/writeout) — penjelasan tiap variabel write-out lengkap dengan contoh.
- [OpenSSL 3.5 Final Release](https://openssl-library.org/post/2025-04-08-openssl-35-final-release/) — 3.5 itu LTS, rilis 8 Apr 2025, didukung sampai 2030. Versi `openssl` yang lo pakai buat `s_client`.
- [RFC 9114 — HTTP/3](https://www.rfc-editor.org/rfc/rfc9114.html) — HTTP di atas QUIC (UDP), sudah stabil dan produksi (~21% traffic global 2025).
- [Cloudflare Radar 2025 Year in Review](https://blog.cloudflare.com/radar-2025-year-in-review/) — data pangsa protokol: HTTP/2 ~50%, HTTP/3 ~21% traffic global 2025.

