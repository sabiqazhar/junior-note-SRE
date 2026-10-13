# Perintah — Week 12 · SLI, SLO, Error Budget, dan Incident Response

Format: perintah → output → interpretasi.
Fokus: SLI/SLO, multiwindow multi-burn-rate, promtool, simulasi insiden.

## Persiapan

### 1. Cek stack & rules termuat
```
docker compose ps
curl -fsS http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[].name' 2>/dev/null || echo "cek UI: Status > Rules"
```
```
NAME                STATUS
sre-lab-api-1       Up
sre-lab-prometheus-1 Up
```
Interpretasi: api & prometheus Up. Rules dari `rules/slo.yml` harus terdaftar.

## Hari 2 (definisi & uji rules)

### 2. Validasi sintaks rule
```
promtool check rules observability/prometheus/rules/slo.yml
```
```
Checking observability/prometheus/rules/slo.yml
  SUCCESS: 7 rules found
```
Interpretasi: 7 rule (4 recording + 3 alerting) valid. Kalau gagal → indent YAML.

### 3. (Opsional) Unit test rule
```
promtool test rules observability/prometheus/rules/slo_test.yml
```
```
SUCCESS: 4 tests found, 4 passed
```
Interpretasi: test case expr lolos. Ini yang sering kelewat — wajib buat CI.

### 4. Cek recording rule availability (error ratio 5m)
```
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_errors_per_request:ratio_rate5m"
```
```
{} 0.002
```
Interpretasi: 0,2% error baseline. Di bawah budget 0,005 → belum ticket.

### 5. Cek recording rule latency p95
```
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_latency_p95:histogram_quantile5m"
```
```
{route="/orders"} 0.18
```
Interpretasi: p95 180ms < 0,3 → `HighLatencyP95` belum firing.

## Hari 3 (simulasi insiden & burn-rate)

### 6. Traffic normal (baseline)
```
for i in $(seq 1 200); do curl -sS -o /dev/null http://localhost/orders; done
```
Interpretasi: bangun baseline request sebelum di rusak.

### 7. Rusak dependency (Redis down)
```
docker compose stop redis
for i in $(seq 1 50); do curl -sS -o /dev/null http://localhost/orders; done
```
Interpretasi: request yang butuh Redis bakal gagal → error ratio naik.

### 8. Cek error ratio melonjak
```
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_errors_per_request:ratio_rate1h"
```
```
{} 0.082
```
Interpretasi: 8,2% error. Burn rate = 0,082 / 0,005 = 16,4× > 14,4 → page.

### 9. Cek alert firing
```
curl -fsS http://localhost:9090/api/v1/alerts | jq '.data.alerts[]|{alert,state}'
```
```
{"alert":"HighErrorBudgetBurn","state":"firing"}
{"alert":"APIDown","state":"inactive"}
```
Interpretasi: burn-rate tinggi → page sesuai `severity: page` di rules.

### 10. Hitung burn rate manual (rumus)
```
# burn rate = error_ratio / (1 - SLO) = 0.082 / 0.005 = 16.4
echo "scale=2; 0.082 / 0.005" | bc
```
```
16.40
```
Interpretasi: 16,4× — jauh di atas 14,4. 2% budget 30 hari habis < 1 jam.

### 11. Cek slow burn (ticket)
```
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_errors_per_request:ratio_rate6h"
```
```
{} 0.012
```
Interpretasi: 1,2% > (1×0,005) → `SlowErrorBudgetBurn` (ticket) bisa firing.

### 12. CARA KEMBALI NORMAL — pulihkan dependency
```
docker compose start redis
curl -fsS "http://localhost:9090/api/v1/query?query=up{job=%22api%22}"
```
```
{} 1
```
Interpretasi: `up==1`, error ratio turun ke baseline, alert `inactive` lagi.
Verifikasi dengan angka, bukan perasaan.

### 13. Verifikasi recovery (error ratio normal)
```
curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_errors_per_request:ratio_rate5m"
```
```
{} 0.001
```
Interpretasi: kembali ~0,1%. Insiden teratasi.

## Hari 4 (postmortem)

### 14. Template postmortem sudah ada
```
ls postmortems/incident-001.md
```
```
postmortems/incident-001.md
```
Interpretasi: salin jadi `incident-002.md`, isi data simulasi (blameless).

### 15. Pastikan action item ber-owner & tenggat
```
grep -n "Owner\|Tenggat\|P1" postmortems/incident-002.md
```
```
| 1 | Tambah alert saturasi Redis | Deteksi | lo | +1 minggu | P1 |
```
Interpretasi: tanpa owner+tenggat, action item cuma harapan. Wajib diisi.

### 16. Cek severity ditetapkan di 5 menit pertama (dokumentasi)
```
grep -n "Severity\|SEV" docs/incident-response.md
```
```
| SEV1 | ... | Langsung, 24/7 |
```
Interpretasi: severity ada di tabel, tetapkan < 5 menit pas insiden.

### 17. Verifikasi alert gejala (bukan penyebab) di rules
```
grep -n "up{job=\"api\"} == 0\|errors_per_request" observability/prometheus/rules/slo.yml
```
```
expr: up{job="api"} == 0          # gejala: user nggak dilayani
expr: job:slo_errors_per_request:ratio_rate1h > (14.4 * 0.005)  # gejala error
```
Interpretasi: alert berbasis gejala (user terdampak), bukan "CPU tinggi".

## Cara buktikan acceptance minggu 12
- `promtool check/test` SUCCESS → perintah 2, 3.
- Recording & alert burn-rate aktif → perintah 4, 9.
- Simulasi: error naik → alert firing → recovery → perintah 7–13.
- Postmortem blameless ber-owner → perintah 14, 15.
