# Perintah — Week 10 · Grafana dan Alerting

Format: perintah → output → interpretasi.
Fokus: RED/USE dashboard, alert gejala, provisioning as code, Alloy.

## Persiapan

### 1. Cek stack jalan
```
docker compose ps
```
```
NAME                IMAGE                    STATUS
sre-lab-api-1       sre-lab-api              Up
sre-lab-prometheus-1 prom/prometheus:3.13.1  Up
sre-lab-grafana-1   grafana/grafana:12.0     Up
sre-lab-alloy-1     grafana/alloy:latest     Up
```
Interpretasi: `grafana` & `alloy` harus Up. Alloy = pipeline log (bukan Promtail).

### 2. Cek data source & target
```
curl -fsS http://localhost:9090/api/v1/targets | jq '.data.activeTargets[]|{job,health}'
```
```
{"job":"api","health":"up"}
{"job":"prometheus","health":"up"}
```
Interpretasi: job `api` up → Grafana bisa tarik metrik.

## Hari 2 (dashboard RED/USE)

### 3. RED — request rate per route
```promql
sum by (route)(rate(http_server_requests_total[5m]))
```
```
{route="/orders"} 12.4
```
Interpretasi: 12,4 req/detik. Panel time series di Grafana.

### 4. RED — error ratio global
```promql
sum(rate(http_server_requests_total{status=~"5.."}[5m]))
/ sum(rate(http_server_requests_total[5m])) * 100
```
```
{} 0.6
```
Interpretasi: 0,6% error. Panel + threshold merah di 1% (misal).

### 5. RED — p95 latency
```promql
histogram_quantile(0.95, sum by (le,route)(rate(http_server_request_duration_seconds_bucket[5m])))
```
```
{route="/orders"} 0.18
```
Interpretasi: p95 180ms. Di bawah SLO 300ms.

### 6. RED — in-flight (saturasi)
```promql
http_server_requests_in_flight
```
```
{} 3
```
Interpretasi: gauge request aktif. Panel stat/gauge.

### 7. USE — CPU container
```promql
rate(container_cpu_usage_seconds_total[5m])
```
```
{container="api"} 0.05
```
Interpretasi: 5% core terpakai. Butuh cAdvisor (job di-prometheus.yml komentar).

### 8. USE — memory vs limit
```promql
container_memory_usage_bytes / on(container) container_spec_memory_limit_bytes
```
```
{container="api"} 0.42
```
Interpretasi: 42% limit memori terpakai.

## Hari 3 (eksperimen alert)

### 9. Matiin API → uji APIDown (page)
```
docker compose stop api
curl -fsS "http://localhost:9090/api/v1/query?query=up{job=%22api%22}"
```
```
{} 0
```
Interpretasi: `up==0`. Setelah `for: 1m` + scrape, `APIDown` firing (page).
Waktu ke firing ≈ 15s×beberapa + 60s.

### 10. Cek alert firing di Prometheus
```
curl -fsS http://localhost:9090/api/v1/alerts | jq '.data.alerts[]|{alert,state}'
```
```
{"alert":"APIDown","state":"firing"}
```
Interpretasi: alert aktif. Di lab nyata ini diteruskan ke Alertmanager/page.

### 11. Hidupkan lagi & ukur recovery
```
docker compose start api
curl -fsS "http://localhost:9090/api/v1/query?query=up{job=%22api%22}"
```
```
{} 1
```
Interpretasi: `up==1`, `APIDown` kembali `inactive`.

### 12. Naikkan error rate → uji burn-rate alert
```
for i in $(seq 1 50); do curl -sS -o /dev/null -X POST -d 'bad' http://localhost/orders; done
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_errors_per_request:ratio_rate5m"
```
```
{} 0.021
```
Interpretasi: 2,1% > 0,005 → `SlowErrorBudgetBurn` (ticket) bisa firing.

### 13. Cek alert latency (recording rule)
```
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_latency_p95:histogram_quantile5m"
```
```
{route="/orders"} 0.18
```
Interpretasi: di bawah 0,3 → `HighLatencyP95` belum firing. Bikin lambat dulu.

### 14. Bikin lambat → uji HighLatencyP95
```
for i in $(seq 1 10); do curl -sS "http://localhost/slow?ms=2000" & done
sleep 30
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_latency_p95:histogram_quantile5m"
```
```
{route="/slow"} 1.0
```
Interpretasi: p95 `/slow` = 1s > 0,3 → setelah `for: 10m` bakal firing (ticket).

## Hari 4 (provisioning & Alloy)

### 15. Verifikasi provisioning as code (no drift)
```
grep -n "allowUiUpdates" observability/grafana/provisioning/dashboards/default.yaml
```
```
    allowUiUpdates: false
```
Interpretasi: dashboard di-Git, UI nggak bisa ubah permanen. Cegah drift.

### 16. Verifikasi pipeline log lewat Alloy (bukan Promtail)
```
docker compose ps | grep -E "alloy|promtail"
```
```
sre-lab-alloy-1   grafana/alloy:latest   Up
```
Interpretasi: cuma Alloy yang jalan. Promtail nggak ada (EOL 2 Mar 2026).

### 17. Cek log masuk Loki lewat Alloy
```logql
{job="api"} | json | level="info" | limit 5
```
```
2026-... level=info msg=request_done route=/orders status=200
```
Interpretasi: Alloy parse JSON, `level` jadi label. Log terstruktur bisa difilter.

## Cara buktikan acceptance minggu 10
- Dashboard RED+USE live → perintah 3–8.
- APIDown firing < 2 menit → perintah 9–11.
- Error/latency alert teruji → perintah 12, 14.
- Pipeline Alloy, bukan Promtail → perintah 16, 17.
