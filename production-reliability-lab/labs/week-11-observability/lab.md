# Lab — Week 11 · Logs dan Distributed Tracing

> Target: nyambungin gejala di metrics ke penyebab di request tertentu.
> Waktu: 4 hari × 2 jam. Stack: `api`, `prometheus`, `loki`, `tempo`, `alloy` (Alloy, bukan Promtail).

## Persiapan (wajib)
```
docker compose ps
```
Pastikan `sre-lab-api-1`, `sre-lab-loki-1`, `sre-lab-tempo-1`,
`sre-lab-alloy-1` status `Up`. Cek `observability/otel/config.alloy` aktif
(pipeline log + trace). Buka Grafana → Explore → pilih data source **Loki** &
**Tempo**. Pastikan app log JSON lewat `slog` (lihat `app/main.go`).

---

## Hari 1 · Teori (2 jam)
- Baca `bahan-bacaan.md` seksi 1–7.
- Catat di `notes.md`: tabel "metrics vs logs vs traces" + 3 prinsip korelasi.
- Pahami `context.Context` dan kenapa `QueryRowContext` cegah trace putus.

**Yang harus terjadi:** lo bisa jelasin kenapa trace putus kalau ctx lupa diteruskan.

**Kalau gagal:** balik ke `bahan-bacaan.md` 2.4.

**Catat notes:** tulis contoh `slog` yang benar dengan `trace_id`.

---

## Hari 2 · Praktik (2 jam) — structured log & trace
1. Cek log app keluar sebagai JSON:
   ```
   docker compose logs api --since 1m | head -3
   ```
   **Yang harus terjadi:** baris JSON berisi `level`, `msg`, `route`, `status`,
   `duration_ms`. (App sudah pakai `slog.NewJSONHandler`.)
   **Kalau gagal:** cek `app/main.go` logger setup.

2. Cek Alloy terima & teruskan ke Loki:
   ```logql
   {job="api"} | json | limit 5
   ```
   **Yang harus terjadi:** 5 log terbaru muncul terstruktur.
   **Kalau gagal:** Alloy mati? cek `docker compose logs alloy`.

3. Cek trace masuk Tempo (kirim 1 request dulu):
   ```
   curl -sS http://localhost/orders >/dev/null
   ```
   Di Grafana Explore → Tempo, cari trace terbaru. **Yang harus terjadi:** trace
   ada dengan span `sre-api` (HTTP) + child span DB.
   **Kalau gagal:** cek OTLP receiver Alloy & app ekspor ke `:4317`.

**Catat notes:** screenshot 1 trace dengan span app + DB.

---

## Hari 3 · Eksperimen (2 jam) — telusuri 1 request lambat
1. Bikin endpoint lambat karena DB (di lab, pakai `pg_sleep` via query, atau
   panjangin `handleOrders`). Kirim request:
   ```
   curl -sS -D - http://localhost/orders >/dev/null
   ```
   Ambil `trace_id` dari log:
   ```logql
   {job="api"} | json | route="/orders" | limit 1
   ```
   **Yang harus terjadi:** log berisi `trace_id` (mis. `4b2c...`).

2. Cari log by `trace_id` di Loki:
   ```logql
   {job="api"} | json | trace_id="4b2c..."
   ```
   **Yang harus terjadi:** semua log untuk request itu muncul.
   **Kalau gagal:** `trace_id` lupa di-inject ke `slog` → tambahin.

3. Buka trace di Tempo by `trace_id`:
   **Yang harus terjadi:** lo lihat total 1,2s, dan 1,1s habis di span database.
   **"Oh, ternyata"**: 90% waktu numpuk di DB, bukan di app.

4. **CARA KEMBALI NORMAL (wajib):**
   ```
   docker compose restart api        # kalau lo ubah handler/query eksperimen
   # kalau cuma kirim request normal, nggak perlu restart — trace/log balik normal
   curl -sS http://localhost:9090/api/v1/query?query=up{job=%22api%22}
   ```
   **Yang harus terjadi:** `up{job="api"}` = 1, log/trace kembali ke pola normal
   (tanpa request lambat). Bukan rollback image — cukup pastikan handler eksperimen
   dikembalikan & container jalan.

**Catat notes:** tulis berapa % waktu di span DB vs app untuk request lambat tadi.

---

## Hari 4 · Review (2 jam) — troubleshooting guide
1. Tulis `labs/week-11-observability/troubleshooting.md`: langkah metrics →
   log → trace dengan screenshot (cara nTelusuri 1 request).
2. Verifikasi tail sampling Alloy simpan error+lambat 100%:
   ```
   grep -n "status_code\|latency\|probabilistic" observability/otel/config.alloy
   ```
3. Commit `troubleshooting.md` + `commands.md` + `evidence/`.

**Catat notes:** 1 hal yang bikin lo "oh, ternyata" minggu ini.

---

## Daftar periksa (checklist)
- [ ] `api`, `loki`, `tempo`, `alloy` Up
- [ ] Log app JSON (`slog`)
- [ ] Log masuk Loki lewat Alloy (bukan Promtail)
- [ ] Trace masuk Tempo, span app + DB nyambung
- [ ] Log bisa difilter by `trace_id`
- [ ] Request lambat berhasil ditelusuri ke span DB
- [ ] Tail sampling: error+lambat 100%, baseline 1%
- [ ] Troubleshooting guide ada

## Cara buktikan (acceptance)
1. Screenshot LogQL by trace_id → `evidence/`.
2. Screenshot Tempo trace (span DB dominan) → `evidence/`.
3. `config.alloy` aktif (Alloy, bukan Promtail) → `docker compose ps`.
4. `troubleshooting.md` langkah lengkap → `evidence/`.
