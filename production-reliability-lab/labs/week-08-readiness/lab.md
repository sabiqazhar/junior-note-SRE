# Lab — Minggu 8 · Production Readiness & Reliability Baseline

> Ikuti **berurutan**, jangan lompat. Setiap langkah: perintah → **Yang harus terjadi** → **Kalau gagal**.

## Persiapan (10 menit)
- [ ] Repo punya `docker/compose.yaml`, `app/`, `nginx/`, `.env` (dari `.env.example`).
- [ ] Minggu 3–7 selesai: stack jalan, CI ada, backup pernah diuji.

## Hari 1 · Teori (2 jam)
### Langkah 1.1 — Pahami PRR
Baca [Evolving SRE Engagement Model](https://sre.google/sre-book/evolving-sre-engagement-model/) & [Launch Checklist](https://sre.google/sre-book/launch-checklist/).
**Yang harus terjadi:** lo paham 5 kategori PRR dan kenapa "jam 3 pagi" jadi standar.
**Catat di notes.md:** definisi PRR dengan bahasa lo.

### Langkah 1.2 — Pelajari FMEA & SPOF
Baca `docs/production-readiness.md` (template) dan bagian "Single Point of Failure".
**Yang harus terjadi:** lo bisa sebutkan 3 SPOF di lab.
**Catat di notes.md:** draf dependency map.

## Hari 2 · Praktik (2 jam)
### Langkah 2.1 — Build dari repo bersih
`docker compose config && docker compose up -d --build --wait && docker compose ps`
**Yang harus terjadi:** semua `healthy`/`running`. Satu rangkaian perintah, tanpa langkah manual.
**Kalau gagal:** `docker compose logs` — perbaiki sampai bisa dari nol.

### Langkah 2.2 — Verifikasi endpoint
`curl -fsS http://localhost/health && curl -fsS http://localhost/readyz`
**Yang harus terjadi:** dua-duanya `200`.
**Kalau gagal:** cek healthcheck & readiness di app.

## Hari 3 · Eksperimen (2 jam)
### Langkah 3.1 — Reboot & ukur recovery
`sudo reboot` (atau `docker compose down && docker compose up -d --wait`).
Lalu: `date +%s; curl -fsS http://localhost/readyz; date +%s`
**Yang harus terjadi:** sistem balik sendiri, `/readyz` 200 tanpa campur tangan. Catat selisih detik = RTO lab.
**Kalau gagal (nggak balik):** cek `restart: unless-stopped` di compose & volume persist.

### Langkah 3.2 — Buktikan data persist
Isi 1 baris di DB, reboot, cek baris masih ada.
**Yang harus terjadi:** data nggak hilang (volume persist).
**Kembali normal:** selesai.

## Hari 4 · Review & dokumentasi (2 jam)
### Langkah 4.1 — Isi PRR
Buka `docs/production-readiness.md`, isi 5 kategori dengan status + bukti. Tulis dependency map & FMEA.
**Yang harus terjadi:** tiap item `DONE` punya bukti; risiko diterima ditulis eksplisit.

### Langkah 4.2 — Tulis risiko & on-call
Di bagian "Risiko yang diterima": sebutkan single instance DB, no multi-AZ, kapasitas belum diuji produksi. Tulis pemilik on-call.

### Langkah 4.3 — Commit
`git add docs/production-readiness.md labs/week-08-readiness && git commit -m "week8: production readiness review"`

## Daftar periksa "minggu ini selesai"
- [ ] Sistem jalan dari repo bersih dengan satu perintah.
- [ ] Reboot → pulih sendiri, RTO tercatat.
- [ ] PRR 5 kategori terisi, `DONE` = teruji.
- [ ] Dependency map + FMEA ada.
- [ ] Risiko & on-call tertulis.

## Cara membuktikan ke reviewer
```bash
docker compose up -d --build --wait
curl -fsS http://localhost/readyz      # 200
sudo reboot
# setelah nyala:
curl -fsS http://localhost/readyz      # 200 tanpa campur tangan
```
