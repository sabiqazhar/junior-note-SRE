# Lab — Week 10 · Grafana dan Alerting

> Target: ubah metrics jadi informasi operasional.
> Waktu: 4 hari × 2 jam. Stack: `api`, `prometheus`, `grafana` (12.x), `alloy` (log pipeline).

## Persiapan (wajib)
```
docker compose ps
```
Pastikan `sre-lab-api-1`, `sre-lab-prometheus-1`, `sre-lab-grafana-1`,
dan `sre-lab-alloy-1` status `Up`. Buka `http://localhost:3000` (admin/admin —
ganti di lab nyata). Pastikan `observability/otel/config.alloy` aktif (bukan
Promtail). Di `/targets` Prometheus, job `api` = UP.

---

## Hari 1 · Teori (2 jam)
- Baca `bahan-bacaan.md` seksi 1–7.
- Catat di `notes.md`: tabel RED vs USE + 3 prinsip alerting (gejala, actionable, severity).
- Pahami `for:` cegah flapping dan `allowUiUpdates: false`.

**Yang harus terjadi:** lo bisa bedain alert gejala vs penyebab dengan contoh.

**Kalau gagal:** balik ke `bahan-bacaan.md` 2.3 & 2.6.

**Catat notes:** tulis 1 alert di `rules/slo.yml` dan sebutin itu gejala atau penyebab.

---

## Hari 2 · Praktik (2 jam) — bangun dashboard
1. Buat dashboard RED (provisioning as code, `allowUiUpdates: false`):
   | Panel | Query |
   | --- | --- |
   | Request rate | `sum by (route)(rate(http_server_requests_total[5m]))` |
   | Error rate % | error ratio × 100 |
   | p95/p99 | `histogram_quantile(0.95, sum by (le,route)(rate(http_server_request_duration_seconds_bucket[5m])))` |
   | In-flight | `http_server_requests_in_flight` |
2. Buat dashboard USE:
   | Panel | Query |
   | --- | --- |
   | CPU container | `rate(container_cpu_usage_seconds_total[5m])` |
   | Memory | `container_memory_usage_bytes` vs limit |
   | DB conns | `pg_stat_activity_count` (butuh postgres-exporter; opsional) |
3. Simpan JSON dashboard ke `evidence/` dan definisi ke provisioning.

**Yang harus terjadi:** panel nunjukin angka live setelah traffic.
**Kalau gagai:** cek data source Prometheus terhubung di Grafana.

**Catat notes:** screenshot dashboard RED sebelum & sesudah traffic.

---

## Hari 3 · Eksperimen (2 jam) — uji alert
1. Matiin API, ukur waktu ke firing:
   ```
   docker compose stop api
   # catat: berapa detik dari stop sampai alert "APIDown" firing?
   docker compose ps | grep api      # harus Exit
   ```
   **Yang harus terjadi:** `up{job="api"} == 0` → `APIDown` firing < 2 menit
   (job scrape 15s + `for: 1m`).
   **Kalau gagal:** cek `for:` di `rules/slo.yml`, cek Alertmanager terhubung.

2. Naikin error rate:
   ```
   docker compose start api
   for i in $(seq 1 50); do curl -sS -o /dev/null -X POST -d 'bad' http://localhost/orders; done
   ```
   **Yang harus terjadi:** error ratio naik, `SlowErrorBudgetBurn`/`HighErrorBudgetBurn`
   bisa firing (lihat `rules/slo.yml`).
   **Kalau gagal:** cek `job:slo_errors_per_request:ratio_rate5m` di Prometheus.

3. Latency tinggi:
   ```
   for i in $(seq 1 10); do curl -sS "http://localhost/slow?ms=2000" & done
   ```
   **Yang harus terjadi:** `job:slo_latency_p95:histogram_quantile5m` > 0,3 →
   `HighLatencyP95` (ticket) firing.

4. **CARA KEMBALI NORMAL (wajib):**
   ```
   docker compose restart api        #hentikan traffic eksperimen & reset
   # atau cukup stop loop /slow; in-flight & latency kembali normal sendiri
   curl -fsS http://localhost:9090/api/v1/query?query=up{job=%22api%22}
   ```
   **Yang harus terjadi:** `up{job="api"}` = 1, error ratio turun ke baseline,
   alert `APIDown` kembali ke `inactive`. Bukan rollback image — cukup pastikan
   container jalan & traffic eksperimen berhenti.

**Catat notes:** isi tabel — alert | lama ke firing | actionable? | false positive?

---

## Hari 4 · Review (2 jam) — runbook
1. Tulis runbook di `labs/week-10-grafana/runbooks/` (format wajib):
   ```markdown
   ## Gejala
   ## Kemungkinan penyebab (urut sering)
   ## Query yang dicek dulu
   ## Mitigasi cepat
   ## Kapan eskalasi
   ```
2. Verifikasi `allowUiUpdates: false` di provisioning (cegah drift).
3. Commit dashboard JSON + runbook + `commands.md`.

**Catat notes:** 1 alert yang menurut lo perlu ditambah di lab ini.

---

## Daftar periksa (checklist)
- [ ] Grafana Up, data source Prometheus terhubung
- [ ] Dashboard RED (rate/error/p95/in-flight) live
- [ ] Dashboard USE (CPU/mem/DB) live
- [ ] `allowUiUpdates: false` aktif
- [ ] `APIDown` firing < 2 menit pas api stop
- [ ] Error/latency alert teruji
- [ ] Setiap alert punya runbook
- [ ] Pipeline log lewat Alloy (bukan Promtail)

## Cara buktikan (acceptance)
1. Screenshot dashboard RED+USE → `evidence/`.
2. Timeline: stop api → alert firing (catat detik) → `evidence/`.
3. Runbook per alert ada di `runbooks/` → `evidence/`.
4. `config.alloy` ada & aktif (bukan promtail) → cek `docker compose ps`.
