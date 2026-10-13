# Ringkasan — Week 12 · SLI, SLO, Error Budget, dan Incident Response

## Inti (5 kalimat)
1. **SLI** = angka dari sisi user (Availability: non-5xx; Latency: <300ms), **SLO** = targetnya (99,5% / 95%), **SLA** = janji ber-denda.
2. **Error budget** = `1 − SLO` (0,5% = 5.000 gagal/1 juta); **burn rate** = `error_ratio / (1−SLO)` = kecepatan ngabisin budget.
3. Alert pakai **multiwindow multi-burn-rate** (14,4 page / 6 page / 1 ticket) biar cepat tapi anti-false-positive di traffic rendah.
4. **Error budget policy** ubah "boleh rilis?" jadi angka: budget habis → freeze rilis, kerjaan pindah ke reliabilitas.
5. Insiden: **tetapkan severity di 5 menit**, **mitigasi dulu** baru cari root cause, lalu **postmortem blameless** ber-action-item ber-owner.

## Peta konsep
```
http_server_requests_total (proxy)
   │ recording rule (30s)
   ▼
job:slo_errors_per_request:ratio_rate{5m,30m,1h,6h}
   │ alerting rule: burn_rate × 0.005
   ▼
HighErrorBudgetBurn (page 14.4/6) / SlowErrorBudgetBurn (ticket 1)
   │
   ▼
Insiden: detect → triage(SEV) → MITIGASI → diagnosis → recover → postmortem
Error budget policy: >50% bebas | 25-50% normal | <25% stop fitur | 0 freeze
```

## Query / perintah kunci + output
| Perintah | Output | Arti |
| --- | --- | --- |
| `sum(rate(http_server_requests_total{status=~"5.."}[5m]))/sum(rate(...[5m]))` | `0.013` | error ratio 1,3% |
| `job:slo_errors_per_request:ratio_rate1h` | `0.013` | recording rule (SLO) |
| `job:slo_latency_p95:histogram_quantile5m` | `0.18` | p95 180ms |
| `promtool check rules rules/slo.yml` | `SUCCESS` | sintaks rule ok |
| `promtool test rules rules/slo_test.yml` | `SUCCESS` | unit test rule ok |
| `up{job="api"} == 0` | `0` | API down → page |

## Failure matrix
| Gejala | Penyebab | Cek | Mitigasi |
| --- | --- | --- | --- |
| `promtool check` gagal | YAML salah/indent | jalankan check | perbaiki expr/indent |
| Alert nggak nyala | burn-rate kebesaran / for: lama | cek recording rule | turunkan threshold sementara |
| False positive di traffic rendah | alert statis, bukan burn-rate | pakai multiwindow | ganti ke burn-rate |
| Severity telat | nggak ada IC | tetapkan di 5 mnt | assign IC tiap insiden |
| Mitigasi lambat | nyari root cause dulu | MITIGASI dulu | rollback/scale dulu |
| Postmortem mandek | action item tanpa owner | isi owner+tenggat | wajib ber-owner |

## Jebakan
- SLO diukur dari metrik internal, bukan proxy → crash nggak kehitung.
- Alert statis "error > 1%" → false positive di traffic rendah.
- Traffic rendah butuh aturan khusus (1 gagal = burn 1000×).
- Postmortem tanpa owner & tenggat = curhatan.
- Severity telat ditentukan → respons kacau.
- Recording rule dilewati → alert lambat & `rate()` berulang.

## Acceptance
1. SLI dihitung dari metrics (sumber + window jelas). ✅ `docs/slo.md`
2. SLO punya periode evaluasi & eksklusi request. ✅ `docs/slo.md`
3. Alert burn-rate terpasang & diuji `promtool test rules`. ✅ `rules/slo.yml`
4. Postmortem blameless: impact, timeline, root cause, 5-whys, action item. ✅ `postmortems/`

## Satu kalimat
SLO dan error budget ngubah debat "boleh rilis nggak?" dari opini jadi angka, dan insiden cuma bisa ditangani kalau lo mitigasi dulu, baru cari penyebab, lalu tulis postmortem tanpa nyalahin orang.
