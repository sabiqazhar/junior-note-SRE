# Ringkasan — Week 10 · Grafana dan Alerting

## Inti (5 kalimat)
1. Metrics dari minggu 9 baru berguna kalau dibungkus **dashboard** (biar gampang dibaca) dan **alerting** (biar narak lo pas rusak).
2. Pakai dua kacamata: **RED** (Rate/Error/Duration) buat service, **USE** (Utilization/Saturation/Error) buat resource — jangan dicampur.
3. **Alert berbasis gejala**, bukan penyebab: "user kena error 5%" lebih penting dari "CPU 80%".
4. Tiap alert wajib **actionable + punya runbook + severity jelas** (`page` vs `ticket`) dan pakai `for:` biar nggak flapping.
5. **Provisioning as code** (`allowUiUpdates: false`) + pipeline log lewat **Alloy** (bukan Promtail, EOL 2 Mar 2026) = dashboard & alert bisa di-reprovision dari Git.

## Peta konsep
```
Prometheus (metrik) ──recording rule──> job:slo_*_rate5m
                              │
                              ├─ Alerting rule (rules/slo.yml)
                              │     APIDown (page), HighLatencyP95 (ticket)
                              ▼
                   Alertmanager / Grafana contact point
                      page ──> on-call     ticket ──> issue
                              ▲
Grafana ◄── PromQL query ────┘
   ├─ Dashboard RED (rate/error/p95/in-flight)
   └─ Dashboard USE (CPU/mem/DB conn)
Provisioning as code: allowUiUpdates: false
Log pipeline: Alloy (bukan Promtail)
```

## Query / perintah kunci + output
| Query / YAML | Output / Arti |
| --- | --- |
| RED rate: `sum by (route)(rate(http_server_requests_total[5m]))` | req/detik per route |
| RED error: `sum(rate(...{status=~"5.."}[5m]))/sum(rate(...[5m]))` | error ratio global |
| RED p95: `histogram_quantile(0.95, sum by (le,route)(rate(http_server_request_duration_seconds_bucket[5m])))` | p95 latency |
| USE: `http_server_requests_in_flight` | saturasi request |
| `up{job="api"} == 0` (alert APIDown) | 1=hidup, 0=mati → page |
| `job:slo_latency_p95:histogram_quantile5m > 0.3` | p95 >300ms → ticket |
| `allowUiUpdates: false` (provisioning) | cegah drift dari UI |

## Failure matrix
| Gejala | Penyebab | Cek | Mitigasi |
| --- | --- | --- | --- |
| Alert nggak nyala pas API mati | `for:` kepanjangan / target DOWN | `/targets`, cek `for: 1m` | perpendek `for:`, pastikan scrape jalan |
| Alert flapping | `for:` = 0 / ambang di batas bergetar | lihat history alert | tambah `for: 2m`+ |
| Dashboard ilang pas redeploy | `allowUiUpdates: true` | provisioning YAML | set `false`, definisikan di Git |
| Alert noise pas maintenance | nggak ada mute timing | Alertmanager/Grafana mute | buat maintenance window |
| Page & ticket nyamber | 1 contact point | notification policy | pisah channel per severity |
| Pipeline log mati | masih pakai Promtail | `config.alloy` | pakai Alloy |

## Jebakan
- `allowUiUpdates: true` → dashboard berubah di UI, hilang pas redeploy (drift).
- Alert tanpa `for:` → flapping, on-call kebal.
- Satu contact point buat semua severity → yang penting tenggelam.
- Nggak ada mute timing buat maintenance → noise.
- Alert "CPU > 80%" tanpa konteks dampak = nggak ditindaklanjuti.
- Masih pakai Promtail (EOL 2 Mar 2026) → pakai Alloy.

## Acceptance
1. Dashboard RED+USE real-time, bisa di-reprovision dari Git. ✅ provisioning YAML + `allowUiUpdates: false`
2. Alert service-down kepicu < 2 menit. ✅ matiin api, ukur waktu ke firing
3. Setiap alert punya severity, runbook, langkah diagnosis. ✅ `runbooks/`
4. Pipeline log lewat Alloy, bukan Promtail. ✅ `config.alloy`

## Satu kalimat
Dashboard bagus nunjukin angka, tapi alert yang *gejala* + *actionable* + *provisioned as code* yang bikin lo bisa tidur tenang.
