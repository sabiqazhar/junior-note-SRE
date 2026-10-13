# Week 03 · Docker production fundamentals

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai docker production fundamentals sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Image layer & multi-stage build
- Container lifecycle & restart policy
- Bridge network & DNS internal compose
- Volume & persistensi
- Healthcheck & depends_on condition
- cgroups v2 & resource limit

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Seluruh service hidup dari kondisi bersih dengan satu perintah (`docker compose up -d --build --wait`).
2. API gagal secara terkendali (bukan panic/hang) saat dependency mati.
3. Setiap container punya limit CPU/memory/pids yang masuk akal dan healthcheck yang benar-benar menguji kesiapan.
4. Tidak ada password hardcoded; rahasia lewat `.env` (di-.gitignore) + `.env.example`.

## Struktur folder

```
labs/week-03-docker/
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

- [Docker Get Started](https://docs.docker.com/get-started/) — image layer, container lifecycle, volume.
- [Docker Compose specification](https://docs.docker.com/compose/how-tos/) — `healthcheck`, `depends_on: condition: service_healthy`, `deploy.resources`.
- [Dockerfile best practices](https://docs.docker.com/build/building/best-practices/) — multi-stage, `.dockerignore`, BuildKit.
- [Chainguard / distroless images](https://www.chainguard.dev/) — image minimal tanpa shell (keamanan).
- [cgroups v2](https://docs.kernel.org/admin-guide/cgroup-v2.html) — resource limit (memory/CPU) di container.

