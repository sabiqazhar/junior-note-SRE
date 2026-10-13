# Lab — Week 09 · Metrics dengan Prometheus

> Target: tahu kondisi service dari metrics, bukan dari chat user.
> Waktu: 4 hari × 2 jam. Stack: `api` (Go) + `prometheus` (3.13.x) + `postgres`.

## Persiapan (wajib sebelum Hari 1)
```
docker compose ps
```
Pastikan `sre-lab-api-1` dan `sre-lab-prometheus-1` status `Up`.
Kalau `prometheus` `Exit`, cek `observability/prometheus/prometheus.yml` —
job `api` menunjuk ke `api:8080` dan `scrape_native_histograms: true` sudah ada.
Buka `http://localhost:9090/targets` → job `api` harus `UP`.

---

## Hari 1 · Teori (2 jam)
- Baca `bahan-bacaan.md` sampai selesai (seksi 1–7).
- Catat di `notes.md`: beda Counter vs Gauge vs Histogram vs Summary pakai
  kata lo sendiri + analogi odometer/speedometer/ember.
- Pahami `rate()` vs nilai mentah, dan kenapa `histogram_quantile` itu aproksimasi.

**Yang harus terjadi:** lo bisa jelasin kenapa `rate(http_server_requests_total[5m])`
lebih benar dari nilai mentah, tanpa buka dokumen.

**Kalau gagal:** balik ke `bahan-bacaan.md` seksi 2.1–2.3. Jangan lanjut praktik
kalau 4 tipe metrik masih baur.

**Catat notes:** tulis 1 contoh metrik di `app/main.go` untuk tiap tipe.

---

## Hari 2 · Praktik (2 jam)
1. Cek raw metrics:
   ```
   curl -fsS http://localhost:8080/metrics | grep http_server
   ```
   **Yang harus terjadi:** muncul `http_server_requests_total`,
   `http_server_request_duration_seconds_bucket`, `http_server_requests_in_flight`.
   **Kalau gagal:** cek app jalan (`docker compose ps`), port 8080 diekspos.

2. Jalankan query di Prometheus UI (`http://localhost:9090`):
   ```
   sum by (route) (rate(http_server_requests_total[5m]))
   ```
   **Yang harus terjadi:** nilai > 0 setelah lo kirim traffic.
   **Kalau gagal (hasil kosong):** target DOWN → lihat `/targets`, cek `lastError`.

3. Simpan 4 query wajib ke `labs/week-09-prometheus/queries.promql`:
   request rate, error ratio, p95 latency, in-flight.

**Catat notes:** screenshot hasil query sebelum ada traffic (nol) dan sesudah.

---

## Hari 3 · Eksperimen (2 jam) — sengaja rusak
1. Kirim 100 request normal:
   ```
   for i in $(seq 1 100); do curl -sS -o /dev/null http://localhost/orders; done
   ```
   **Yang harus terjadi:** `status="200"` naik ~100.

2. Kirim 20 request error:
   ```
   for i in $(seq 1 20); do curl -sS -o /dev/null -X POST -d 'bad' http://localhost/orders; done
   curl -sS http://localhost:8080/metrics | grep 'http_server_requests_total{.*status="500"'
   ```
   **Yang harus terjadi:** tepat 20 di `status="500"`.
   **Kalau gagal:** pastikan `POST /orders` memang return 500 (lihat `handleOrders`).

3. Bikin lambat & cek histogram:
   ```
   curl -sS "http://localhost/slow?ms=2000" &
   ```
   Lalu query p95:
   ```
   histogram_quantile(0.95, sum by (le,route)(rate(http_server_request_duration_seconds_bucket[5m])))
   ```
   **Yang harus terjadi:** `{route="/slow"}` mendekati 2s.
   **Kalau gagal:** request terlalu cepat selesai — naikin `ms` atau kirim berulang.

4. **CARA KEMBALI NORMAL (wajib):**
   ```
   docker compose restart api        # reset counter & state eksperimen
   # atau cukup berhenti kirim request /slow; in-flight akan turun sendiri
   curl -fsS http://localhost:8080/metrics | grep http_server_requests_in_flight
   ```
   **Yang harus terjadi:** `in_flight` kembali ke angka kecil (0–3),
   error ratio kembali mendekati baseline. Bukan rollback image — cukup stop
   traffic eksperimen & restart container kalau perlu.

**Catat notes:** selisih error ratio sebelum (#2 awal) vs sesudah (#2 akhir).

---

## Hari 4 · Review (2 jam)
1. Uji kardinalitas:
   ```
   topk(10, count by (__name__)({__name__=~".+"}))
   ```
   **Yang harus terjadi:** jumlah seri rendah, TIDAK ada `user_id`/`request_id`.
2. Verifikasi `route` dipakai, bukan path mentah:
   ```
   count by (route) (http_server_requests_total)
   ```
3. Commit `queries.promql` + `commands.md` + `evidence/`.

**Catat notes:** simpulkan 1 kesalahan paling umum yang lo sendiri suka lakuin.

---

## Daftar periksa (checklist)
- [ ] `docker compose ps` → api & prometheus Up
- [ ] `/targets` → job `api` UP
- [ ] 4 query wajib tersimpan & jalan
- [ ] 100 req normal → counter 200 naik
- [ ] 20 req error → counter 500 naik tepat 20
- [ ] p95 `/slow` ≈ 2s
- [ ] `topk` kardinalitas rendah, tanpa label high-card
- [ ] Cara kembali normal sudah dicoba & dokumentasikan

## Cara buktikan (acceptance)
1. Buka `/targets`, screenshot job `api` = UP → masukkan ke `evidence/`.
2. Sebelum/sesudah kirim 20 error: `grep status="500"` naik 20 → `evidence/`.
3. Query p95 jalan & `{route="/slow"}` ≈ 2s → `evidence/`.
4. `topk` kardinalitas: daftar metrik & jumlah seri → `evidence/`.
