# Perintah — Week 09 · Metrics dengan Prometheus

Format: perintah → output → interpretasi.
Semua query di bawah pakai nama metrik yang ada di `app/main.go`:
`http_server_requests_total`, `http_server_request_duration_seconds_bucket`,
`http_server_requests_in_flight`, dan `up`.

## Cek kondisi stack (persiapan)

### 1. Pastikan stack jalan
```
docker compose ps
```
```
NAME                IMAGE                          STATUS        PORTS
sre-lab-api-1       sre-lab-api                    Up 2 hours    0.0.0.0:8080->8080/tcp
sre-lab-prometheus-1 prom/prometheus:3.13.1        Up 2 hours    0.0.0.0:9090->9090/tcp
sre-lab-postgres-1  postgres:16                    Up 2 hours    5432/tcp
```
Interpretasi: `api` dan `prometheus` berstatus `Up`. Kalau salah satu `Exit`,
scrape akan gagal dan target di `/targets` muncul DOWN.

### 2. Cek target scrape dari sisi Prometheus
```
curl -fsS http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job, health, lastError}'
```
```
{"job":"api","health":"up","lastError":""}
{"job":"prometheus","health":"up","lastError":""}
```
Interpretasi: kedua job `up`. Kalau `health` = `down`, cek `lastError`
(biasanya DNS container, port salah, atau app belum register `/metrics`).

## Hari 2 (kondisi normal)

### 3. Lihat raw metrics dari aplikasi
```
curl -fsS http://localhost:8080/metrics | grep http_server_requests_total
```
```
http_server_requests_total{method="GET",route="/orders",status="200"} 1842
http_server_requests_total{method="GET",route="/orders",status="500"} 11
```
Interpretasi: counter naik tiap request. `status="500"` mencatat error.
Nilai mentah counter nggak boleh dipakai langsung — pakai `rate()`.

### 4. Request rate per route (req/detik)
```promql
sum by (route) (rate(http_server_requests_total[5m]))
```
```
{route="/orders"}   12.4
{route="/slow"}     0.2
```
Interpretasi: `/orders` dilayani ~12,4 req/detik rata-rata 5 menit terakhir.
`rate()` otomatis ngurusin reset pas aplikasi restart.

### 5. Error ratio per route (0..1)
```promql
sum by (route) (
  rate(http_server_requests_total{status=~"5.."}[5m])
) / sum by (route) (rate(http_server_requests_total[5m]))
```
```
{route="/orders"}   0.006
```
Interpretasi: 0,6% request ke `/orders` gagal 5xx. Di bawah SLO 0,5%?
Belum — tapi ini cuma satu route. Agregasi global lihat perintah 12.

### 6. Request in-flight (saturasi murah)
```promql
http_server_requests_in_flight
```
```
{}   3
```
Interpretasi: 3 request sedang diproses. Naik terus padahal rate datar =
request nggak selesai (dependency melambat). Gejala awal sebelum latency rusak.

## Hari 3 (eksperimen kegagalan)

### 7. Kirim 100 request normal
```
for i in $(seq 1 100); do curl -sS -o /dev/null http://localhost/orders; done
echo "kirim 100 ok"
```
```
kirim 100 ok
```
Interpretasi: baseline counter naik ~100 di `status="200"`.

### 8. Kirim 20 request error (POST body rusak)
```
for i in $(seq 1 20); do curl -sS -o /dev/null -X POST -d 'bad' http://localhost/orders; done
curl -sS http://localhost:8080/metrics | grep 'http_server_requests_total{.*status="500"'
```
```
http_server_requests_total{method="POST",route="/orders",status="500"} 20
```
Interpretasi: tepat 20 request gagal tercatat. Bukti instrumentasi jalan.

### 9. Error ratio global (5 menit)
```promql
sum(rate(http_server_requests_total{status=~"5.."}[5m]))
/ sum(rate(http_server_requests_total[5m]))
```
```
{}   0.013
```
Interpretasi: ~1,3% error secara keseluruhan setelah eksperimen. Di atas
budget 0,5% → kalau ini 30 hari, alert burn rate minggu 12 bakal nyala.

### 10. p95 latency — agregasi bucket DULU, baru quantile
```promql
histogram_quantile(0.95,
  sum by (le, route) (rate(http_server_request_duration_seconds_bucket[5m]))
)
```
```
{route="/orders"}   0.18
{route="/slow"}     1.0
```
Interpretasi: p95 `/orders` = 180ms, `/slow` = 1s (karena endpoint sengaja
lambat). Jangan rata-rata hasil quantile antar instance — selalu jumlahkan
`_bucket` dulu.

### 11. Cek bucket mana yang naik (distribusi)
```promql
sum by (le) (increase(http_server_request_duration_seconds_bucket[5m]))
```
```
{le="0.1"}   520
{le="0.25"}  610
{le="0.5"}   640
{le="+Inf"}  642
```
Interpretasi: hampir semua request <0,5s, 642 total. Beda kecil antar ember
terakhir = distribusi menumpuk di bawah 0,5s.

## Hari 4 (review & kardinalitas)

### 12. Error ratio global via recording rule (sudah didefinisikan minggu 12)
```promql
job:slo_errors_per_request:ratio_rate5m
```
```
{}   0.013
```
Interpretasi: nama yang sama dengan rule di `rules/slo.yml`. Dipakai alert
burn rate. Di sini 0,013 jauh di atas `(14.4 * 0.005)` = 0,072? Tidak — belum
page, tapi di atas 0,005 → ticket SlowErrorBudgetBurn bakal nyala.

### 13. Cek kardinalitas: 10 metrik paling gemuk
```promql
topk(10, count by (__name__)({__name__=~".+"}))
```
```
http_server_requests_total                             42
http_server_request_duration_seconds_bucket            120
```
Interpretasi: jumlah seri rendah & wajar. Kalau ada `user_id`/`request_id`
di sini, itu red flag kardinalitas.

### 14. Cek kardinalitas per label combo (route, bukan path)
```promql
count by (route) (http_server_requests_total)
```
```
{route="/orders"}   2
{route="/slow"}     2
```
Interpretasi: cuma 2 seri per route (status 200/500). Kardinalitas rendah &
aman karena pakai `r.Pattern`, bukan `r.URL.Path` mentah.

### 15. Native histogram aktif? (Prometheus 3.8+)
```promql
count({__name__=~".+_bucket", le="0.5"})
```
```
120
```
Interpretasi: bucket klasik masih ada. Kalau mau native histogram, butuh
`scrape_native_histograms: true` di `prometheus.yml` (sudah disetel di lab ini)
plus metrik diekspos sebagai native histogram. Cek:
```
curl -fsS http://localhost:9090/api/v1/query?query=count%28native_histogram%29
```

### 16. Cek range selector minimal 4× scrape interval
```promql
rate(http_server_requests_total[30s])   # CUMA 2 sampel (15s x 2) -> rewel
rate(http_server_requests_total[5m])    # 20 sampel -> stabil
```
Interpretasi: pakai `[5m]` untuk grafik, `[1h]`–`[6h]` untuk SLO.

### 17. Log Prometheus kalau target down
```
docker compose logs prometheus --since 5m | grep -i "scrape"
```
```
ts=... level=warn msg="Error on ingesting samples" ... connection refused
```
Interpretasi: kalau lihat `connection refused`, app belum jalan atau port
salah di `prometheus.yml`.

## Cara buktikan acceptance minggu 9
- Target UP (`/targets`) → perintah 2.
- Metrics naik sesuai traffic/error → perintah 3, 7, 8.
- rate / error ratio / p95 masuk akal → perintah 4, 5, 10.
- Tanpa label berkardinalitas tinggi → perintah 13, 14.
