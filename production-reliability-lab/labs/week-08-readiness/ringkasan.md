# Ringkasan — Minggu 8 · Production Readiness & Reliability Baseline

> Satu halaman buat dibaca ulang sebelum wawancara/ujian. Bukan pengganti `bahan-bacaan.md`.

## Inti minggu ini dalam 5 kalimat
1. PRR = audit 5 kategori: Security, Availability, Observability, Recovery, Deployment.
2. `DONE` cuma boleh kalau **sudah diuji**, bukan cuma ditulis.
3. FMEA wajib punya kolom "gimana deteksinya" — tanpa itu, nggak bisa ditindaklanjuti.
4. SPOF harus **dinyatakan eksplisit** beserta alasan penerimaannya.
5. Syarat tertinggi: dari repo kosong, **satu perintah** sampai sistem jalan & pulih sendiri setelah reboot.

## Peta konsep
- Production Readiness Review
  - 5 kategori: Security / Availability / Observability / Recovery / Deployment
  - Status: `TODO` / `PARTIAL` / `DONE` (DONE = teruji)
- Dependency mapping → FMEA (gagal, dampak, deteksi, mitigasi, SPOF?)
- Single Point of Failure (nyatakan, jangan sembunyi)
- Recovery teruji: reboot → `/readyz` 200 tanpa campur tangan

## Perintah kunci + yang harus lo lihat
| Perintah | Yang dicari | Artinya kalau beda |
| --- | --- | --- |
| `docker compose config` | valid, no error | Compose salah → build gagal |
| `docker compose up -d --build --wait` | semua `healthy` | Ada service nggak naik |
| `curl -fsS http://localhost/readyz` | `200` | Readiness belum siap |
| `sudo reboot` lalu `curl /readyz` | balik 200 < 2 mnt | Nggak pulih otomatis |

## Failure matrix / decision tree
| Gejala | Kemungkinan penyebab | Cara buktikan | Mitigasi |
| --- | --- | --- | --- |
| Reboot → service nggak balik | volume hilang / restart policy salah | `docker inspect` RestartPolicy, `docker volume ls` | `restart: unless-stopped` + volume persist |
| `/readyz` 503 setelah reboot | dependency belum sehat | `docker compose ps` | `depends_on: service_healthy` |
| Data hilang | volume nggak persist | `docker volume inspect` | mount named volume |

## Jebakan umum
- PRR jadi dokumen sekali pakai yang langsung basi.
- FMEA tanpa kolom deteksi → nggak bisa ditindaklanjuti.
- Dependency pihak ketiga (auth/payment) lupa dicatat.
- Nggak ada pemilik on-call tertulis.

## Acceptance criteria
1. Sistem bisa dibangun & dijalankan ulang **hanya dari repository**.
2. Semua langkah pemulihan **terdokumentasi DAN teruji**.
3. Risiko yang belum ditangani **ditulis eksplisit**.

## Satu kalimat yang harus lo ingat
> PRR bukan formulir — dia jawab: "jam 3 pagi, bisa nggak orang jaga benerin sistem ini tanpa nanya?"
