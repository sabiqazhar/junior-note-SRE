# Ringkasan — Minggu 16 · Load Testing, Chaos, dan Final Incident Simulation

> Satu halaman untuk dibaca ulang sebelum ujian/wawancara. Bukan pengganti bahan-bacaan.md.

## Inti 5 kalimat

1. **Little's Law (L = λ × W)** ngasih tahu concurrency yang lo butuh buat capai throughput tertentu.
2. Saat **utilisasi ρ → 1**, latency meledak eksponensial = **saturasi**; itu titik kapasitas lo.
3. **Retry storm** terjadi kalau retry di banyak lapis tanpa **jitter** → traffic jadi gerombolan (sampai 27×).
4. **Timeout budget** (menurun per hop) + **circuit breaker** cegah sistem numpuk nunggu dependency mati.
5. Chaos = eksperimen ber**hipotesis + abort condition**; tanpa itu, lo cuma ngerusak, bukan belajar.

## Peta konsep

```
λ (req/s) ─┐
           ├─ Little: L = λ × W  →  butuh concurrency berapa?
W (latency)┘
           ├─ ρ = λ / μ  →  ρ→1 = saturasi (p95 meledak)
           │
dependency lambat ─> retry tanpa jitter ─> RETRY STORM
           │
           ├─ timeout budget 2s > 1s > 200ms  (gagal cepat)
           ├─ circuit breaker: buka(50% err/10s) → setengah-buka(30s)
           │
k6 ramp VU ─> catat p95/err/CPU/DB ─> titik saturasi
toxiproxy/Chaos Mesh ─> rusak ─> abort condition ─> pulih ─> bandingkan
```

## Perintah kunci + output

```bash
# 1. Cek versi k6 (v2.x — PIN di CI)
$ k6 version
# output: k6 v2.0.x (...)
# interpretasi: catat versi; jangan biarkan CI pakai latest.

# 2. Baseline load test
$ k6 run loadtest/loadtest.js
# output: ✓ p(95)<300 ✓ p(99)<800 ✓ http_req_failed<0.01
# interpretasi: di 50 VU threshold lolos = kapasitas cukup.

# 3. Simpan hasil JSON
$ k6 run --summary-mode=full --out json=loadtest/results.json loadtest/loadtest.js
# output: results.json berisi metrics lengkap
# interpretasi: buat dibandingkan per tahap beban.

# 4. Naik beban (stages ke 200 VU)
$ k6 run loadtest/loadtest.js
# output: p(95)=820ms ✗ (threshold failed)
# interpretasi: di 200 VU SATURASI — p95 meledak.

# 5. Little's Law concurrency
$ python3 -c "print(500*0.2)"
# output: 100.0
# interpretasi: butuh ~100 request in-flight buat 500 req/s @200ms.

# 6. Threshold gagal = CI gagal
$ k6 run loadtest/loadtest.js; echo $?
# output: ✗ thresholds failed / exit 99
# interpretasi: CI harus berhenti saat SLO jebol.

# 7. Toxiproxy: proxy Redis
$ toxiproxy-cli create redis -l 0.0.0.0:26379 -u redis:6379
# output: Created proxy redis
# interpretasi: proxy di depan Redis buat inject gangguan.

# 8. Inject latency 200ms
$ toxiproxy-cli add redis --latency 200
# output: Added toxic latency to redis
# interpretasi: tiap request Redis lambat 200ms.

# 9. Chaos Mesh: kill 1 pod
$ kubectl apply -f podchaos-kill-api.yaml
# output: podchaos.chaos-mesh.org/kill-api created
# interpretasi: sengaja rusak 1 replica, lihat self-healing.

# 10. Amati pemulihan
$ kubectl get pods -w
# output: pod lama hilang, baru Running 3/3 ~10s
# interpretasi: ReplicaSet ganti otomatis.

# 11. Metric Prometheus (p99)
$ curl -s 'http://localhost:9090/api/v1/query?query=histogram_quantile(0.99,...)'
# output: {value:[...,0.81]}
# interpretasi: p99 = 0.81s, dari data nyata bukan tebakan.

# 12. Diagnosis log
$ kubectl logs deploy/sre-api --tail=50 | grep -i redis
# output: timeout connecting redis
# interpretasi: akar masalah = latency Redis.

# 13. ABORT & pulih
$ kubectl delete podchaos kill-api && toxiproxy-cli delete redis
# output: deleted
# interpretasi: eksperimen dihentikan, proxy dibersihkan.

# 14. Ulangi setelah perbaikan
$ k6 run loadtest/loadtest.js
# output: p(95)=290ms ✓
# interpretasi: perbaikan turunkan p95 (before/after terukur).

# 15. CLEANUP wajib
$ kubectl delete -f kubernetes/ && kind delete cluster --name sre
# output: deleted / Deleted cluster sre
# interpretasi: lab bersih, nggak ada chaos nyangkut.
```

## Failure matrix

| Gejala | Penyebab | Cek | Penyelesaian |
| --- | --- | --- | --- |
| p95 naik, throughput datar | saturasi (ρ→1) | `kubectl top`, conn DB | naik resource / optimasi query |
| error rate lonjak pas retry | retry storm | trace retry count | backoff + jitter |
| Antrean numpuk | timeout terlalu besar | app timeout config | turunkan timeout budget |
| Dependency mati → kaskade | nggak ada circuit breaker | error rate spike | pasang breaker |
| Chaos bikin insiden beneran | nggak ada abort condition | — | hentikan, pasang abort |
| Threshold k6 flaky di CI | versi k6 beda | `k6 version` | pin versi di CI |

## Jebakan

- Cuma ukur rata-rata latency (penting p95/p99 + error rate).
- Naikin VU terus tanpa cek utilisasi → ukur antrean, bukan kapasitas.
- Load test dari laptop → kebatas bandwidth sendiri.
- Chaos tanpa abort condition → insiden beneran.
- Retry di 3 lapis tanpa jitter → amplifikasi 27×.
- k6 nggak di-pin versi → perilaku beda tiap CI.

## Cheat sheet cepat

| Mau ngapain | Perintah |
| --- | --- |
| Load test | `k6 run loadtest/loadtest.js` |
| Simpan hasil | `k6 run --out json=results.json ...` |
| Inject latency | `toxiproxy-cli add redis --latency 200` |
| Chaos pod kill | `kubectl apply -f podchaos-*.yaml` |
| Baca p99 | query Prometheus `histogram_quantile` |
| Hentikan chaos | `kubectl delete podchaos ...` |
| Bersih | `kind delete cluster --name sre` |

## Tanya jawab kilat

- **Q: Little's Law buat apa?** A: tahu concurrency (L) dari λ & W.
- **Q: ρ→1?** A: saturasi, latency meledak eksponensial.
- **Q: Jitter itu?** A: acak di backoff biar retry nggak gerombolan.
- **Q: Timeout budget?** A: batas waktu menurun per hop, gagal cepat.
- **Q: Chaos tanpa abort?** A: sabotase, bukan eksperimen.

## Acceptance

1. Load test bisa diulang & punya threshold yang ngegagalkan CI.
2. Bottleneck didukung data (saturasi, ρ, koneksi DB), bukan dugaan.
3. Insiden terdeteksi, didiagnosis, dimitigasi, dipulihkan dengan timeline tercatat.
4. Hasil sebelum vs sesudah perbaikan bisa dibandingkan kuantitatif.

## Studi kasus mini

- **Kasus A:** naik VU terus tanpa cek CPU → yang diukur cuma antrean, bukan kapasitas. Pelajaran: catat utilisasi tiap tahap, cari titik saturasi.
- **Kasus B:** retry di client+LB+app tanpa jitter → pas dependency pulih, semua retry bareng = 27× amplifikasi. Pelajaran: backoff + jitter di tiap lapis.
- **Kasus C:** chaos tanpa abort → Redis mati bikin insiden beneran. Pelajaran: selalu ada hipotesis + abort condition.

## Istilah kunci (flashcard)

- **Little's Law** = L = λ × W.
- **Utilisasi (ρ)** = λ / μ; →1 = saturasi.
- **Saturasi** = latency meledak eksponensial.
- **Retry storm** = retry beramai-ramam tanpa jitter.
- **Timeout budget** = batas waktu menurun per hop.
- **Circuit breaker** = henti trafik ke dependency mati.
- **Threshold (k6)** = syarat gagalkan CI.
- **Steady state** = keadaan normal sebelum di-rusak.
- **Abort condition** = hentikan chaos otomatis.

## Pertanyaan wawancara

1. Hitung L untuk λ=500, W=200ms?
2. Kenapa "kuat 800 rps" tanpa latency omong kosong?
3. Kapan saturasi terjadi matematis?
4. Kenapa jitter penting di retry?
5. Kenapa chaos butuh abort condition?

## Pola yang sering kelewat

- Cuma ukur rata-rata latency.
- Naik VU tanpa cek utilisasi.
- Load test dari laptop.
- Chaos tanpa abort / k6 nggak di-pin.

## Cek cepat sebelum CI

| Cek | Perintah |
| --- | --- |
| Load test | `k6 run loadtest/loadtest.js` |
| Threshold gagal | turunin p95 → exit non-zero |
| Chaos aman | ada hipotesis + abort |
| Versi | `k6 version` (pin di CI) |

## Satu kalimat

> Ukur kapasitas pakai Little's Law & cari titik saturasi, lindungi sistem dengan timeout budget + circuit breaker + jitter, dan buktikan pemulihan lewat chaos berhipotesis — bukan sekadar harapan.
