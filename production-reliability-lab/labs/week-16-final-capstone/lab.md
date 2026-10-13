# Lab — Minggu 16 · Load Testing, Chaos, dan Final Incident Simulation

> Ikuti BERURUTAN, jangan lompat. Setiap langkah formatnya: perintah → Yang harus terjadi → Kalau gagal → Catat notes.

## Persiapan

- [ ] Cluster kind dari minggu 15 masih ada, atau buat ulang (`kind create cluster --name sre --image kindest/node:v1.34.0`).
- [ ] `k6` terinstall (catat versi: `k6 version` — sekarang v2.x, **PIN di CI**).
- [ ] `toxiproxy` / `toxiproxy-cli` tersedia buat inject latency.
- [ ] `loadtest/loadtest.js` sudah ada (threshold p95<300, p99<800, err<1%).

---

## Hari 1 · Teori (2 jam)

1. **Baca** `bahan-bacaan.md` section 1–2 → **Yang harus terjadi:** paham Little's Law & saturasi. → **Kalau gagal:** ulang 2.1–2.2. → **Catat notes:** hitung L untuk λ=500, W=200ms (=100) di `notes.md`.
2. **Tiga pelindung** → **Yang harus terjadi:** jelasin timeout budget, backoff+jitter, circuit breaker. → **Kalau gagal:** ulang 2.3–2.5. → **Catat notes:** nilai awal masing-masing.
3. **Disiplin chaos** → **Yang harus terjadi:** paham hipotesis + abort condition. → **Kalau gagal:** ulang 2.7. → **Catat notes:** template hipotesis.
4. **Baca** `loadtest/loadtest.js` → **Yang harus terjadi:** paham `ramping-vus`, `thresholds`, `timeout: '2s'`. → **Kalau gagal:** ulang 2.6. → **Catat notes:** kenapa `timeout` itu budget.

---

## Hari 2 · Praktik (2 jam)

1. **Baseline:** **`k6 run loadtest/loadtest.js`** → **Yang harus terjadi:** threshold lolos, p95<300. → **Kalau gagal:** cek apakah app jalan (`kubectl get pods`). → **Catat notes:** isi tabel VU=50 (p50/p95/p99/err/CPU/DB conn).
2. **Naik beban:** edit `stages` ke 100, lalu 200 VU → **`k6 run --summary-mode=full --out json=loadtest/results.json loadtest/loadtest.js`** → **Yang harus terjadi:** di satu titik p95 naik tajam padahal throughput datar = **saturasi**. → **Kalau gagal:** cek `kubectl top pod`. → **Catat notes:** titik saturasi (VU berapa).
3. **Hitung ρ:** dari throughput max & kapasitas → **Yang harus terjadi:** ρ mendekati 1 di titik saturasi. → **Kalau gagal:** kurang data, ulang beban. → **Catat notes:** ρ di tiap tahap.
4. **Threshold gagal CI:** sengaja turunin threshold jadi p95<100 → **`k6 run`** → **Yang harus terjadi:** exit non-zero ("thresholds failed"). → **Kalau gagal:** cek sintaks thresholds. → **Catat notes:** bukti CI bisa digagalkan.
5. **Catat tabel beban** (VU / req-s / p50 / p95 / p99 / err% / CPU / DB conn) untuk 10, 50, 100, 200. → **Yang harus terjadi:** tabel lengkap. → **Kalau gagal:** ulang step 2. → **Catat notes:** ini bukti bottleneck.

---

## Hari 3 · Eksperimen (2 jam) — Game day

Siapkan **hipotesis tertulis** dulu (wajib):

```markdown
## Hipotesis
Kalau Redis mati, API tetap balas < 500ms dan error rate < 1%,
karena cache miss jatuh ke database yang masih sanggup.
## Metrik steady-state: p95 < 300ms, error < 1%
## Blast radius: cuma container Redis di lab.
## Abort condition: kalau error > 20% selama 2 menit, eksperimen dihentikan & Redis nyala lagi.
```

1. **Traffic normal:** k6 di 30 VU, catat baseline. → **Yang harus terjadi:** p95<300. → **Kalau gagal:** app bermasalah. → **Catat notes:** baseline.
2. **Rusak (toxiproxy latency):** **`toxiproxy-cli create redis -l 0.0.0.0:26379 -u redis:6379`** lalu **`toxiproxy-cli add redis --latency 200`** → **Yang harus terjadi:** latency naik. → **Kalau gagal:** cek koneksi proxy. → **Catat notes:** seberapa p95 naik.
3. **Amati:** alert apa nyala, berapa lama. → **Yang harus terjadi:** lo catat menit deteksi. → **Kalau gagal:** cek Prometheus/Grafana. → **Catat notes:** dashboard kepakai.
4. **Diagnosis:** metrics → logs → trace. → **Yang harus terjadi:** tau Redis penyebab. → **Kalau gagal:** `kubectl logs`. → **Catat notes:** waktu diagnosis.
5. **Mitigasi & pulih:** matikan toxiproxy / restart Redis → **Yang harus terjadi:** p95 balik normal. → **Kalau gagal:** cek abort condition. → **Catat notes:** dampak.
6. **Ulangi setelah perbaikan** (misal pasang circuit breaker / turbo cache) → bandingkan angka. → **Yang harus terjadi:** p95 lebih baik. → **Catat notes:** before vs after.

🧹 **HARI 3 WAJIB — CLEANUP (chaos cuma lab):**

```bash
toxiproxy-cli delete redis           # hapus proxy latency
kubectl delete podchaos kill-api     # hentikan chaos (kalau pakai Chaos Mesh)
kubectl delete -f kubernetes/        # bersihkan manifest
kind delete cluster --name sre       # hapus cluster
# Yang harus terjadi: nggak ada proxy/latency/chaos nyangkut.
# Chaos di atas CUMA di environment lab — jangan dijalankan di produksi.
```

---

## Hari 4 · Final review (2 jam)

1. **Incident report** (format postmortem minggu 12): timeline, deteksi, diagnosis, mitigasi. → **Yang harus terjadi:** timeline tercatat. → **Kalau gagal:** lihat `postmortems/`. → **Catat notes.**
2. **Benchmark before/after** (tabel VU, p50/p95/p99, err, CPU, DB conn, 2 versi). → **Yang harus terjadi:** dibandingkan kuantitatif. → **Kalau gagal:** ulang hari 2. → **Catat notes.**
3. **Batas kapasitas terukur:** "sistem sanggup X req/detik p95<300ms di spes A, bottleneck = B, buktinya ...". → **Yang harus terjadi:** pernyataan berangka. → **Catat notes.**
4. **Rapikan** README utama, diagram final, runbook. → **Yang harus terjadi:** reviewer paham 30 detik. → **Catat notes.**

---

## Cheat sheet per hari

| Hari | Fokus | Perintah inti |
| --- | --- | --- |
| 1 | Teori | Little's Law, timeout budget, chaos discipline |
| 2 | Praktik | `k6 run`, naik VU, hitung ρ, threshold gagal |
| 3 | Rusak | `toxiproxy-cli`, game day, `destroy` cluster |
| 4 | Review | incident report, before/after, batas kapasitas |

## Rubrik observasi per hari

| Hari | Yang lo amati | Tanda lulus |
| --- | --- | --- |
| 1 | Paham Little's Law & pelindung | Bisa hitung L & jelasin breaker |
| 2 | Kapasitas terukur | titik saturasi ketemu dari data |
| 3 | Insiden teratasi | hipotesis + abort + timeline |
| 4 | Laporan final | incident report + before/after |

## Troubleshooting cepat

| Error | Penyebab | Solusi |
| --- | --- | --- |
| p95 naik, throughput datar | saturasi (ρ→1) | naik resource / optimasi |
| error lonjak pas retry | retry storm | backoff + jitter |
| antrean numpuk | timeout kegedean | turunkan timeout budget |
| dependency mati → kaskade | nggak ada breaker | pasang circuit breaker |
| chaos jadi insiden | nggak ada abort | hentikan, pasang abort |
| threshold k6 flaky | versi beda | pin versi di CI |

## Alat bantu wajib

| Alat | Kegunaan | Cara cek |
| --- | --- | --- |
| `k6` | load test (v2.x) | `k6 version` |
| `toxiproxy-cli` | inject latency | `toxiproxy-cli --help` |
| `kubectl` | chaos di K8s | `kubectl version` |
| Prometheus | baca p99 | `kubectl port-forward svc/prometheus 9090` |

## Langkah pemulihan darurat

- Load test ngegas CPU: turunkan VU di `stages`, ulang bertahap.
- Chaos makin parah: jalankan abort — `kubectl delete podchaos`, `toxiproxy-cli delete`.
- Cluster corrupt: `kind delete cluster --name sre` lalu buat ulang.

## Verifikasi akhir

- [ ] `k6 run` bisa diulang dengan threshold gagal-CI.
- [ ] Tabel beban nunjukin titik saturasi.
- [ ] Game day punya hipotesis + abort + timeline; toxiproxy/chaos bersih.

## Daftar periksa

- [ ] k6 bisa diulang & punya threshold gagal-CI.
- [ ] Titik saturasi ditemukan dari data (p95 naik, throughput datar).
- [ ] Chaos berhipotesis + abort condition; insiden terdeteksi→diagnosis→pulih.
- [ ] Before/after perbaikan bisa dibandingkan kuantitatif.
- [ ] toxiproxy / chaos / cluster dibersihkan (nggak nyangkut).

## Cara buktikan

Lo lulus kalau: (1) `k6 run` bisa diulang dengan threshold yang gagal saat dibawah standar, (2) lo bisa nunjukin titik saturasi dari tabel beban, (3) game day punya hipotesis + abort condition dan lo catat timeline pemulihan, (4) tabel before/after menunjukkan perbaikan terukur. Bukti di `evidence/` sudah disanitasi.
