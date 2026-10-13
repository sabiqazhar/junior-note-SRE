# Minggu 10 · Grafana dan Alerting — Bahan Bacaan

> "Dashboard yang bagus bukan yang paling cantik, tapi yang bikin lo tahu harus bangunin orang jam 3 pagi atau nggak."

Dokumen ini berdiri sendiri. Baca dari atas ke bawah, kerjakan `lab.md` setelahnya.

---

## 1. Kenapa minggu ini penting

Minggu lalu lo sudah punya angka: request rate, error ratio, p95. Tapi angka yang cuma bisa lo lihat kalau lo *sengaja* buka Prometheus itu bukan observability — itu cuma data. Observability baru jadi kalau angkanya **narak lo** pas ada masalah, dan pas lo bangun, lo langsung tahu mau ngapain. Dua hal itulah yang minggu ini lo bangun: **dashboard** (biar angka gampang dibaca) dan **alerting** (biar angka那个 yang ngarahin lo, bukan sebaliknya).

Bayangin sistem ini kayak panel indikator di mobil. Lo nggak nunggu mesin berasap baru tahu overheating — ada lampu merah di dashboard yang nyala *sebelum* rusak parah. Tapi kalau lampunya nyala tiap lo ngerem (false positive), lo bakal matiin alarm itu dan pas mesin beneran overheat, lo nggak sadar. Itulah **alert fatigue**: alarm yang terlalu sering atau nggak actionable bikin on-call kebal. Minggu ini lo belajar bedain "alarm yang harus dibunyikan" vs "grafik yang cuma perlu dilihat besok".

Masalah kedua yang minggu ini beresin: **dashboard yang nggak bisa diandalin**. Kalau dashboard lo cuma hasil klik-klik di UI dan nggak ada di Git, pas Grafana redeploy panel lo hilang. Makanya lo bakal pakai **provisioning as code** — dashboard dan alert didefinisikan di file, bukan di klik. Sumber kebenaran = Git, bukan UI.

---

## 2. Konsep inti

### 2.1 RED — kacamata buat service/request
RED itu tiga angka yang lo tanya tiap kali ada endpoint: **R**ate (req/detik), **E**rror (berapa yang gagal), **D**uration (sebaran latency). Cocok buat API.

```promql
# Rate
sum by (route) (rate(http_server_requests_total[5m]))
# Error ratio
sum(rate(http_server_requests_total{status=~"5.."}[5m])) / sum(rate(http_server_requests_total[5m]))
# Duration p95
histogram_quantile(0.95, sum by (le,route)(rate(http_server_request_duration_seconds_bucket[5m])))
```

### 2.2 USE — kacamata buat resource/infra
USE buat CPU, memori, disk, koneksi DB: **U**tilization (% waktu sibuk), **S**aturation (seberapa numpuk antrean), **E**rror (error level resource).

```promql
# Utilization CPU container
rate(container_cpu_usage_seconds_total[5m])
# Saturation: request yang nggak selesai
http_server_requests_in_flight
# Error resource (contoh koneksi DB, butuh postgres-exporter)
pg_stat_activity_count
```

Jangan campur RED dan USE di satu panel yang nggak jelas — pisahin: panel "kesehatan service" vs "kesehatan infra".

### 2.3 Alert berbasis gejala, bukan penyebab
Ini prinsip paling penting. "User ngalamin error 5%" lebih berguna dari "CPU 80%" — kalau CPU 80% nggak bikin user rugi, itu bukan insiden. Gejala = apa yang dirasakan user. Penyebab = diagnosis, urusan dashboard & investigasi, bukan alarm.

```yaml
# Alert gejala (sudah ada di rules/slo.yml)
- alert: APIDown
  expr: up{job="api"} == 0
  for: 1m
  labels: { severity: page }
```

### 2.4 Setiap alert harus actionable + punya runbook
Kalau nggak ada yang bisa lo lakuin pas alarm nyala, itu bukan alert — itu grafik. Tiap alert wajib punya runbook (langkah diagnosis & mitigasi). Tanpa runbook = alarm tanpa petunjuk jalan keluar.

### 2.5 Severity: page vs ticket
Beda perlakuan. **page** = bangunin orang (SEV1/2, dampak luas). **ticket** = dilihat besok (SEV3, terdampak sebagian). Satu contact point buat semua severity = yang penting tenggelam. Pisahin channel.

### 2.6 `for:` untuk cegah flapping
Alert tanpa `for:` bakal nyala-mati-nyala (flapping) kalau metriknya bergetar di ambang batas. `for: 2m` artinya baru firing kalau kondisi terpenuhi 2 menit berturut-turut.

```yaml
- alert: HighLatencyP95
  expr: job:slo_latency_p95:histogram_quantile5m > 0.3
  for: 10m          # jangan 0 — hindari flapping
```

### 2.7 Provisioning as code & `allowUiUpdates: false`
Dashboard dan alert didefinisikan di file YAML/JSON di Git, bukan hasil klik UI. Kunci: `allowUiUpdates: false` biar nggak ada drift.

```yaml
# observability/grafana/provisioning/dashboards/default.yaml
apiVersion: 1
providers:
  - name: sre-core
    folder: SRE
    type: file
    allowUiUpdates: false      # sumber kebenaran = Git
    options:
      path: /var/lib/grafana/dashboards
```

### 2.8 Promtail EOL → pakai Alloy
**Promtail resmi End-of-Life 2 Maret 2026.** Pipeline log lab ini jalan lewat **Grafana Alloy** (`observability/otel/config.alloy`), bukan Promtail. Jangan mulai lab baru pakai Promtail.

---

## 3. Mekanisme di balik layar

1. **Scrape**: Prometheus tarik `/metrics` dari `api` tiap 15s, simpan deret waktu.
2. **Rule evaluation**: tiap `evaluation_interval`, Prometheus jalanin recording rule (`job:slo_errors_per_request:ratio_rate5m`) lalu alerting rule (`HighErrorBudgetBurn`, `APIDown`).
3. **Alert pending → firing**: kalau `expr` benar dan tahan selama `for:`, status berubah `pending` → `firing`. Tanpa `for:`, langsung firing (rawan flapping).
4. **Notification**: Prometheus kirim ke Alertmanager (atau Grafana alert ke contact point). Severity `page` → on-call; `ticket` → issue tracker.
5. **Visualisasi**: Grafana query Prometheus lewat PromQL, render panel. Dashboard di-load dari file provisioning, bukan dari DB UI.
6. **Runbook**: tiap alert punya `runbook_url` — on-call buka, ikutin langkah, pulihkan dulu baru cari root cause.

---

## 4. Glosarium mini

| Istilah | Arti singkat |
| --- | --- |
| RED | Rate, Error, Duration — 3 metrik service |
| USE | Utilization, Saturation, Error — 3 metrik resource |
| Symptom-based alerting | Alert berdasar gejala user, bukan penyebab |
| Alert fatigue | Kebal terhadap alarm karena terlalu sering/noise |
| Flapping | Alert nyala-mati berulang di ambang batas |
| `for:` | Durasi kondisi harus bertahan sebelum firing |
| Provisioning as code | Dashboard/alert didefinisikan di file, bukan UI |
| Drift | Bedanya state nyata vs yang di Git |
| Severity page/ticket | Bangunin orang vs catat besok |
| Alloy | Pengganti Promtail untuk pipeline log/trace |
| Runbook | Langkah diagnosis & mitigasi tiap alert |
| Mute timing | Jendela maintenance agar alert diam sementara |

---

## 5. Bahan bacaan wajib (primer)

1. Grafana — Alerting overview: https://grafana.com/docs/grafana/latest/alerting/
2. Grafana — Provision dashboards/alerts as code: https://grafana.com/docs/grafana/latest/administration/provisioning/
3. Grafana — Alloy (pengganti Promtail): https://grafana.com/docs/alloy/latest/
4. Prometheus — Alerting (rules, `for:`, severity): https://prometheus.io/docs/practices/alerting/
5. Google SRE Book — Monitoring distributed systems (symptom-based): https://sre.google/sre-book/monitoring-distributed-systems/
6. Prometheus — Querying functions (rate, histogram_quantile): https://prometheus.io/docs/prometheus/latest/querying/functions/

---

## 6. Self-check (harus bisa dijawab tanpa Google)

1. Apa bedanya RED dan USE, dan kapan masing-masing dipakai?
2. Kenapa alert "CPU > 80%" sering bukan alert yang berguna?
3. Apa fungsi `for:` dan kenapa tanpa itu alert bisa flapping?
4. Apa bedanya severity `page` dan `ticket`?
5. Kenapa `allowUiUpdates: false` penting buat provisioning as code?
6. Promtail masih dipakai di lab ini nggak? Kenapa tidak?

---

## 7. Yang sering salah paham

- **"Dashboard = observability."** Salah. Dashboard cuma visualisasi. Tanpa alert yang narak lo, lo tetap nggak tahu pas rusak.
- **"Alert bisa nyala tiap detik."** Boleh, tapi tanpa `for:` dan tanpa actionable runbook → alert fatigue. On-call jadi matiin alarm.
- **"Satu channel buat semua alert."** Salah. Page dan ticket harus pisah, kalau nggak yang penting tenggelam.
- **"Promtail masih oke."** Promtail EOL 2 Maret 2026 — pakai Alloy untuk lab baru.
- **"Alert penyebab lebih bagus dari alert gejala."** Kebalik. User rugi karena gejala, bukan karena CPU tinggi. Alert gejala dulu, penyebab diinvestigasi.
- **"Klik di UI lebih cepat dari YAML."** Kelihatannya iya, tapi pas redeploy dashboard ilang. Provisioning as code menyelamatkan lo.
