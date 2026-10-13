# Ringkasan — Week 09 · Metrics dengan Prometheus

## Inti (5 kalimat)
1. Prometheus itu time-series database yang **menarik (pull)** metrik dari app lo tiap beberapa detik lewat endpoint `/metrics`.
2. Ada 4 tipe metrik: **Counter** (cuma naik), **Gauge** (naik-turun), **Histogram** (distribusi, pakai ember), **Summary** (kuantil di aplikasi — jarang dipakai).
3. Nilai mentah counter nggak pernah dipakai langsung; yang lo pakai adalah `rate()` / `increase()` untuk dapat kecepatan, dan `histogram_quantile()` untuk persentil dari histogram.
4. **Kardinalitas** (jumlah kombinasi label) adalah pembunuh nomor satu — label `user_id` atau `path` mentah bikin Prometheus meledak; pakai `route` (pola) bukan nilai.
5. Di Prometheus 3.8+, **native histogram nggak aktif default** — harus `scrape_native_histograms: true` di scrape config.

## Peta konsep
```
Prometheus (pull tiap 15s)
   │  scrape job "api" -> api:8080/metrics
   ▼
app/main.go mengekspos:
   ├─ http_server_requests_total{method,route,status}   (Counter)
   ├─ http_server_request_duration_seconds_bucket       (Histogram)
   └─ http_server_requests_in_flight                    (Gauge)
   │
   ▼
PromQL:
   rate() / increase()      -> kecepatan counter
   histogram_quantile()     -> persentil dari _bucket
   avg_over_time()/max()    -> gauge
   │
   ▼
Recording rule (minggu 12) -> job:slo_errors_per_request:ratio_rate5m
Alert burn rate            -> HighErrorBudgetBurn
```

## Query / perintah kunci + output
| Query | Output contoh | Arti |
| --- | --- | --- |
| `sum by (route)(rate(http_server_requests_total[5m]))` | `{route="/orders"} 12.4` | 12,4 req/detik ke `/orders` |
| `sum(rate(http_server_requests_total{status=~"5.."}[5m]))/sum(rate(http_server_requests_total[5m]))` | `0.013` | 1,3% error global (5m) |
| `histogram_quantile(0.95, sum by (le,route)(rate(http_server_request_duration_seconds_bucket[5m])))` | `{route="/orders"} 0.18` | p95 = 180ms |
| `http_server_requests_in_flight` | `3` | 3 request sedang diproses (saturasi) |
| `up{job="api"}` | `1` | target hidup (0 = mati) |
| `topk(10, count by (__name__)({__name__=~".+"}))` | `http_server_request_duration_seconds_bucket 120` | cek metrik paling gemuk |
| `job:slo_errors_per_request:ratio_rate5m` | `0.013` | hasil recording rule SLO |

## Failure matrix
| Gejala | Kemungkinan penyebab | Cek | Mitigasi |
| --- | --- | --- | --- |
| Target DOWN di `/targets` | App belum jalan / port salah | `docker compose ps`, `curl :8080/metrics` | Pastikan container Up, port 8080 diekspos |
| `lastError: connection refused` | App crash / path salah | `docker compose logs api` | Perbaiki app, cek `scrape_configs` |
| Counter grafik "lompat" ke bawah | Pod restart (counter reset) | `rate()` vs nilai mentah | Selalu pakai `rate()`, bukan nilai mentah |
| p95 mentok di batas bucket (+Inf) | Bucket terlalu sempit / request lambat banget | `sum by (le)(increase(_bucket[5m]))` | Lebarin buckets / perbaiki latency |
| Prometheus disk penuh | Kardinalitas meledak | `topk(20, count by (__name__,...))` | Hapus label high-cardinality |
| `rate()` hasil nol/aneh | Range < 4× scrape interval | pakai `[5m]` bukan `[30s]` | Naikin range selector |
| Histogram quantile kosong | Nama metric `_bucket` salah | cek `curl /metrics \| grep duration` | Samakan nama dengan `app/main.go` |

## Jebakan
- `histogram_quantile` itu **aproksimasi** (asumsi distribusi seragam dalam bucket). Bucket lebar = hasil ngaco.
- **Jangan rata-rata quantile antar instance.** Jumlahkan `_bucket` dulu, baru quantile.
- `rate()` butuh range **minimal 4× scrape interval** (15s → pakai `[1m]`+).
- Native histogram nggak nyala sendiri — `scrape_native_histograms: true` wajib.
- Label dilarang: `user_id`, `path` mentah, `request_id` → kardinalitas meledak.
- Counter mentah (tanpa `rate`) dipakai buat grafik → selalu "naik", nggak informatif.

## Acceptance
1. Prometheus scrape API → target UP di `/targets`. ✅ cek `curl :9090/api/v1/targets`
2. Metrics naik sesuai traffic & error → bukti sebelum/sesudah (`commands.md` #3, #7, #8).
3. rate, error ratio, p95 jalan & masuk akal → `commands.md` #4, #5, #10.
4. Tanpa label berkardinalitas tinggi → `commands.md` #13, #14.

## Satu kalimat
Sebelum lo bikin alert yang waras, lo harus paham tipe metrik, cara baca pakai `rate()`/`histogram_quantile()`, dan batas kardinalitas — karena 80% alert sampah berawal dari metrik yang salah.
