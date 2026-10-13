# Perintah — Week 16 · Load Test, Chaos & Incident Simulation

Format: perintah → output → interpretasi.

## Hari 2 (kondisi normal)

```bash
# 1. Cek versi k6 (sekarang v2.x — PIN di CI)
$ k6 version
# output: k6 v2.0.x (... )
# interpretasi: catat versi, jangan biarkan CI pakai latest.

# 2. Baseline load test
$ k6 run loadtest/loadtest.js
# output: ✓ p(95)<300  ✓ p(99)<800  ✓ http_req_failed<0.01
# interpretasi: di 50 VU, threshold lolos = kapasitas cukup.

# 3. Simpan hasil JSON (full summary)
$ k6 run --summary-mode=full --out json=loadtest/results.json loadtest/loadtest.js
# output: hasil metrics ke results.json
# interpretasi: buat dibandingkan per tahap beban.

# 4. Naik beban (edit stages ke 200 VU dulu)
$ k6 run loadtest/loadtest.js
# output: p(95)=820ms ✗ (threshold failed)
# interpretasi: di 200 VU sistem SATURASI — p95 meledak.

# 5. Hitung concurrency via Little's Law
$ python3 -c "print(500*0.2)"   # λ=500, W=0.2
# output: 100.0
# interpretasi: butuh ~100 request in-flight buat 500 req/detik @200ms.

# 6. Threshold gagal = CI gagal
$ k6 run --env EXPECT=p95_100 loadtest/loadtest.js; echo $?
# output: ✗ thresholds failed / exit 99
# interpretasi: CI harus berhenti (non-zero exit) saat SLO jebol.
```

## Hari 3 (eksperimen kegagalan / chaos)

```bash
# 7. Siapkan proxy latency (toxiproxy)
$ toxiproxy-cli create redis -l 0.0.0.0:26379 -u redis:6379
# output: Created proxy redis
# interpretasi: proxy di depan Redis buat inject gangguan.

# 8. Inject latency 200ms
$ toxiproxy-cli add redis --listen 0.0.0.0:26379 --upstream redis:6379 --latency 200
# output: Added toxic latency to redis
# interpretasi: tiap request Redis lambat 200ms.

# 9. Jalankan k6 selama chaos
$ k6 run loadtest/loadtest.js
# output: p(95)=540ms ✗ / error 3%
# interpretasi: latency & error naik — hipotesis "tetap <500ms" GAGAL.

# 10. Chaos Mesh: kill 1 pod api
$ kubectl apply -f - <<'EOF'
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata: { name: kill-api }
spec: { action: pod-kill, mode: one, selector: { labelSelectors: { app: sre-api } } }
EOF
# output: podchaos.chaos-mesh.org/kill-api created
# interpretasi: sengaja rusak 1 replica, lihat self-healing.

# 11. Amati pemulihan
$ kubectl get pods -w
# output: pod lama hilang, pod baru Running 3/3 dalam ~10s
# interpretasi: ReplicaSet ganti otomatis (self-healing terbukti).

# 12. Cek metric Prometheus (contoh p99)
$ kubectl port-forward svc/prometheus 9090 &
$ curl -s 'http://localhost:9090/api/v1/query?query=histogram_quantile(0.99,rate(http_request_duration_seconds_bucket[5m]))'
# output: {value:[...,0.81]}  (p99 = 0.81s)
# interpretasi: p99 terbaca dari Prometheus, bukan tebakan.

# 13. Diagnosis log
$ kubectl logs deploy/sre-api --tail=50 | grep -i redis
# output: ... timeout connecting redis
# interpretasi: akar masalah = latency Redis (sesuai hipotesis).

# 14. ABORT: hentikan chaos & pulihkan
$ kubectl delete podchaos kill-api
$ toxiproxy-cli delete redis
# output: deleted
# interpretasi: eksperimen dihentikan, proxy dibersihkan.

# 15. Ulangi setelah perbaikan (contoh: circuit breaker / cache turbo)
$ k6 run loadtest/loadtest.js
# output: p(95)=290ms ✓
# interpretasi: perbaikan turunkan p95 di bawah 300ms (before/after terukur).

# 16. CLEANUP wajib
$ kubectl delete -f kubernetes/
$ kind delete cluster --name sre
# output: deleted / Deleted cluster sre
# interpretasi: lab bersih, nggak ada chaos nyangkut di produksi.
```
