# Week 04 · DNS, TLS, dan reverse proxy

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai dns, tls, dan reverse proxy sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Reverse proxy & upstream
- Proxy header & timeout
- Upstream keep-alive
- TLS & sertifikat lokal
- Failure matrix 502/504/499

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Client hanya bisa mengakses API lewat reverse proxy; port Postgres/Redis tidak terbuka ke host.
2. 502/504/499 bisa direproduksi sengaja dan dijelaskan penyebabnya dari log.
3. TLS lokal terpercaya jalan (mkcert/step-ca) dan bisa diverifikasi.
4. Header forwarding (Host, X-Forwarded-For, X-Forwarded-Proto) benar dan terbukti di log aplikasi.

## Struktur folder

```
labs/week-04-reverse-proxy/
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

- [NGINX Beginner's Guide](https://nginx.org/en/docs/beginners_guide.html) — konsep `server`, `location`, `proxy_pass`.
- [ngx_http_proxy_module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html) — timeout, `proxy_set_header`, `proxy_next_upstream`.
- [mkcert (GitHub)](https://github.com/FiloSottile/mkcert) — sertifikat lokal tepercaya buat lab.
- [step-ca (smallstep)](https://github.com/smallstep/certificates) — CA lokal yang lebih serius.
- [Let's Encrypt](https://letsencrypt.org/) — kenapa sertifikat sekarang umur 6 hari.
- [Caddy docs](https://caddyserver.com/docs/) — alternatif reverse proxy "HTTPS otomatis".
- [Traefik docs](https://doc.traefik.io/) — alternatif proxy buat Docker/Kubernetes.

