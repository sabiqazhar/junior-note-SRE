# Minggu 16 · Load Testing, Chaos, dan Final Incident Simulation — Bahan Bacaan

> *"Lo nggak tahu kapasitas sistem lo sampai lo ukur; dan lo nggak tahu dia pulih kalau lo nggak pernah ngerusaknya."* — prinsip minggu ini.

---

## 1. Kenapa penting

Dua pertanyaan yang harus lo bisa jawab di minggu ke-16: **"sistem ini sanggup berapa req/detik sebelum p95 rusak, dan bottleneck-nya apa?"** serta **"kalau rusak jam 3 pagi, apa yang bikin dia pulih sendiri?"**. Tanpa load test, "server kita kuat" cuma mitos. Tanpa chaos experiment, "kalau DB mati aman" cuma harapan.

Kenapa ini penting buat junior SRE: mayoritas insiden produksi bukan karena kode error, tapi karena **sistem kelebihan beban** atau **dependency lambat** yang memicu efek berantai. Lo harus bisa **mengukur** kapasitas dan **membuktikan** mekanisme pemulihan, bukan cuma ngandelin firasat.

Analognya: load test itu kayak **minta kasir dikasih antrean makin panjang** buat tahu kapan dia mulai telat. Little's Law ngasih tahu berapa banyak "orang ngantre" (request in-flight) yang harus lo siapin. Chaos experiment itu **sengaja matiin satu kompor di dapur** pas jam sibuk, buat tau apakah pesanan lain masih jalan — dan yang paling penting, ada tombol berhenti kalau kebakaran makin parah.

---

## 2. Konsep inti

### 2.1 Little's Law: L = λ × W

Jumlah request di dalam sistem (L) = kedatangan per detik (λ) × waktu di sistem (W). Kalau target 500 req/detik dengan latency 200ms (0,2 detik), lo butuh minimal `500 × 0,2 = 100` request in-flight (concurrency). Ini yang nentuin berapa VU/koneksi lo butuhkan.

### 2.2 Utilisasi & saturasi: ρ = λ / μ

ρ = kedatangan / kapasitas layanan. **Saat ρ mendekati 1, latency meledak eksponensial, bukan linear.** Itu definisi **saturasi**. "Server kuat 800 req/detik" itu omong kosong kalau nggak nyebut latency di angka itu — di 801 req/detik, p95 bisa loncat 10×.

### 2.3 Retry storm

Pas dependency lambat, tiap lapisan (client, LB, app) retry otomatis. Tanpa **jitter**, semua retry terjadi **pada detik yang sama** → traffic jadi "gerombolan" yang makin nabrak. Retry di 3 lapis tanpa jitter = amplifikasi sampai 27×. Pembuatannya: `min(2^n × base, cap)` + jitter acak.

```js
// backoff eksponensial + jitter (konsep, bukan k6)
const wait = Math.min(2 ** n * base, cap) * (0.5 + Math.random());
```

### 2.4 Timeout budget

Request nggak boleh nunggu selamanya. Total budget menurun per hop: proxy 2s → app→DB 1s → app→Redis 200ms. Lebih baik **gagal cepat** daripada numpuk koneksi mati. Di k6, `timeout: '2s'` itu budget per request.

```js
const res = http.get(`${BASE}/orders`, { timeout: '2s', tags: { endpoint: 'orders' } });
```

### 2.5 Circuit breaker

Kalau dependency jelas mati, berhenti nembak dia sebentar. Buka setelah 50% error dalam 10s, setengah-buka (coba sedikit) setelah 30s. Ini cegah lo ngebakar resource ke service yang udah drop.

### 2.6 k6 v2.x & load test ber-threshold

k6 **v1.0 rilis 6 Mei 2025** (native TS, SemVer); sekarang **v2.x** — **pin versi di CI** karena beda major bisa ubah perilaku. Threshold bikin load test **bisa gagalkan CI**:

```js
export const options = {
  scenarios: { ramp: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '1m', target: 50 },
      { duration: '3m', target: 50 },   // steady state = baseline
      { duration: '1m', target: 0 },
    ],
  }},
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99)<800'],
    http_req_failed:   ['rate<0.01'],
  },
};
```

### 2.7 Chaos experiment (Chaos Mesh / LitmusChaos / toxiproxy)

Jangan asal ngerusak. Chaos = eksperimen berhipotesis + **abort condition**. Alat: **toxiproxy** (inject latency/timeout di level proxy), **Chaos Mesh** / **LitmusChaos** (pod kill, network partition di K8s). Tanpa hipotesis & abort condition, itu sabotase, bukan eksperimen.

---

### 2.8 Membaca hasil k6 & menentukan SLO

k6 mencetak ringkasan di akhir: `http_reqs` (total), `http_req_failed` (rate error), dan `http_req_duration` dengan persentil `p(50)/p(95)/p(99)`. **Jangan cuma lihat rata-rata** — rata-rata menyembunyikan ekor. p95 itu "95% request lebih cepat dari angka ini"; p99 buat nangkep kasus terburuk. Di lab ini threshold ditetapkan `p(95)<300` dan `p(99)<800` serta `error<1%`; kalau melanggar, k6 keluar dengan status non-zero sehingga **CI gagal**.

Cara nentuin angka SLO-nya: mulai dari harapan user (misal "halaman order harus balas < 300ms untuk mayoritas"), lalu ukur baseline, lalu pasang threshold sedikit di atas baseline sehat. Jangan pasang terlalu longgar (nggak terdeteksi nyata) atau terlalu ketat (flaky tiap CI). `handleSummary` di `loadtest/loadtest.js` juga nulis `results.json` supaya lo bisa plot tabel beban dan lihat **di VU berapa p95 mulai melengkung** — itu titik saturasi yang dicari di Hari 2.

## 3. Mekanisme di balik layar

Gimana load test + chaos jalan, step by step:

1. **Baseline:** k6 di 30 VU → catat p95, error rate, CPU, koneksi DB.
2. **Naik beban:** ramp ke 50, 100, 200 VU → di tiap tahap catat tabel. Cari titik p95 naik tajam padahal throughput datar = **saturasi**.
3. **Rusak (chaos):** inject latency 200ms ke Redis (toxiproxy) atau `SELECT pg_sleep(0.5)`.
4. **Amati:** alert apa nyala, berapa lama, dashboard mana kepakai.
5. **Diagnosis:** metrics → logs → trace. Catat menit yang lo butuh.
6. **Mitigasi & pulih:** rollback/restart/failover; bandingkan angka sebelum vs sesudah.

---

## 4. Glosarium mini

| Istilah | Arti singkat |
| --- | --- |
| Little's Law | L = λ × W; hubungkan concurrency, throughput, latency |
| Utilisasi (ρ) | λ / μ; mendekati 1 = saturasi |
| Saturasi | Titik latency meledak eksponensial |
| Retry storm | Retry beramai-ramai tanpa jitter menabrak sistem |
| Backoff + jitter | Penundaan eksponensial + acak cegah gerombolan |
| Timeout budget | Batas waktu per hop, menurun ke bawah |
| Circuit breaker | Hentikan trafik ke dependency yang mati |
| Threshold (k6) | Syarat yang gagalkan load test / CI |
| Steady state | Keadaan normal sebelum di-rusak (baseline) |
| Abort condition | Syarat hentikan chaos otomatis |
| toxiproxy | Proxy buat inject latency/timeout |
| Chaos Mesh | Alat chaos di Kubernetes |

---

## 5. Bahan bacaan wajib

1. **SRE Workbook** — https://sre.google/workbook/ (capacity, postmortem, error budget)
2. **k6 Documentation** — https://grafana.com/docs/k6/latest/ (threshold, scenarios, v2.x)
3. **k6 Docs (k6.io)** — https://k6.io/docs/ (panduan praktis load test)
4. **Chaos Mesh** — https://chaos-mesh.org/ (chaos experiment di K8s)
5. **toxiproxy** — https://github.com/Shopify/toxiproxy (inject latency/timeout)
6. **SRE Book: Postmortem Culture** — https://sre.google/sre-book/postmortem-culture/ (format insiden)

---

## 6. Self-check

1. Pakai Little's Law, kalau λ=500 req/detik dan W=200ms, berapa L (concurrency)?
2. Kenapa "kuat 800 req/detik" tanpa sebut latency itu omong kosong?
3. Apa itu saturasi, dan kapan terjadi secara matematis?
4. Kenapa retry tanpa jitter bisa jadi 27× amplifikasi di 3 lapis?
5. Apa itu timeout budget, dan kenapa menurun per hop?
6. Kapan circuit breaker "setengah-buka", dan kenapa?
7. Kenapa chaos tanpa hipotesis + abort condition = sabotase?

---

## 7. Yang sering salah paham

1. **"Cukup ukur rata-rata latency."** Yang penting p95/p99 & error rate — rata-rata nyembunyiin ekor.
2. **"Naikin VU terus sampai error."** Lo cuma ngukur antrean, bukan kapasitas. Cek utilisasi CPU/DB tiap tahap.
3. **"Load test dari laptop."** Yang kebatas bandwidth lo sendiri, bukan sistem. Jalankan dekat target.
4. **"Chaos tanpa abort condition."** Malah bikin insiden beneran. Selalu ada tombol berhenti.
5. **"Retry di tiap lapis tanpa jitter."** Amplifikasi 27× saat dependency pulih barengan.
6. **"k6 versi bebas."** v1.0 (Mei 2025) → v2.x; pin versi di CI, beda major bisa ubah perilaku.
