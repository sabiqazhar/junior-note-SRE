# Minggu 09 · Metrics dengan Prometheus — Bahan Bacaan

> Dokumen ini berdiri sendiri. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham konsepnya. Baru setelah itu kerjakan lab.md.

---

## 1. Kenapa minggu ini penting

Pernah nggak lo dapet chat, "halaman order lelet banget dari tadi", lalu lo buka server, cek CPU, cek memory, lihat log, dan… nggak nemu apa-apa? Atau sebaliknya: dashboard lo hijau semua, tapi user tetap ngeluh. Dua-duanya punya akar yang sama: **lo nggak punya angka yang jujur tentang apa yang dialami user**. Yang lo punya cuma perasaan dan potongan log yang kebetulan lewat.

Minggu ini lo berhenti nebak. Prometheus itu pada dasarnya cuma **database deret waktu (time series)** yang diisi dengan cara ditarik (pull) dari aplikasi lo tiap beberapa detik. Tiap angka yang lo simpan punya nama dan separuh kartu identitas — namanya **label**. Dari angka-angka itu lo bisa jawab tiga pertanyaan yang jadi fondasi semua kerjaan SRE setelah ini: **seberapa banyak traffic masuk**, **berapa persen yang gagal**, dan **seberapa lama user nunggu**. Tiga angka ini nanti jadi SLI minggu 12, jadi panel dashboard minggu 10, dan jadi bahan buat nentuin severity pas insiden.

Masalah kedua yang minggu ini mulai lo beresin: **alert bunyi terus tapi nggak ada yang ngapa-ngapain**. Salah satu penyebabnya teknis, bukan kultural: metriknya salah. Counter dibaca mentah-mentah (angka yang cuma naik bikin grafik "selalu naik" dan nggak pernah bisa dibandingkan), persentil dihitung dari histogram yang ember-nya kelewat lebar, atau satu label `user_id` bikin Prometheus nyimpen jutaan deret waktu sampai disk penuh. Jadi sebelum lo bisa bikin alert yang waras, lo harus paham dulu **tipe metrik, cara bacanya, dan batas kardinalitas**.

---

## 2. Konsep inti

### 2.1 Counter — odometer mobil

Counter cuma bisa **naik**, dan cuma boleh balik ke nol kalau prosesnya restart. Analoginya odometer mobil: angkanya nambah terus, dan nggak pernah berkurang walau lo mundur.

Contoh di `app/main.go`:

```
http_server_requests_total{method="GET",route="/orders",status="200"} 1284
http_server_requests_total{method="GET",route="/orders",status="500"} 17
```

Nilai mentah counter **nggak pernah lo pakai langsung**. Yang lo mau itu *kecepatannya*:

```promql
sum by (route) (rate(http_server_requests_total[5m]))
```

`rate()` ngitung "berapa kenaikan per detik" selama 5 menit terakhir, dan **otomatis ngurusin reset** (restart aplikasi). Ini penting: kadang counter turun bukan karena error, tapi karena pod/container baru mulai dari nol.

Aturan praktis: **range selector minimal 4× scrape interval.** Dengan `scrape_interval: 15s`, jangan pakai `[30s]` — cuma dapet 2 sampel, hasilnya rewel dan gampang nol. Pakai `[5m]` (20 sampel) buat grafik, `[1h]`–`[6h]` buat aturan SLO minggu 12.

### 2.2 Gauge — speedometer

Gauge bisa **naik dan turun**. dipakai buat "keadaan saat ini": koneksi DB aktif, memori terpakai, request yang sedang dilayani.

```promql
http_server_requests_in_flight
```

Gauge langsung dipakai, nggak perlu `rate()`. Kalau mau lihat rata-rata atau puncaknya dalam periode tertentu, baru pakai `avg_over_time()` / `max_over_time()`.

`http_server_requests_in_flight` di lab ini adalah sinyal **saturasi** yang paling murah: kalau angkanya naik terus sementara request rate datar, berarti request nggak selesai-selesai — biasanya dependency (Postgres/Redis) melambat. Ini gejala awal yang sering muncul *sebelum* latency rusak.

### 2.3 Histogram klasik — ember ukuran

Histogram nyatet **distribusi** nilai, bukan satu nilai. Bayangin deretan ember berukuran: `<5ms`, `<10ms`, `<25ms`, `<50ms`, `<100ms`, `<250ms`, `<500ms`, `<1s`, `<2.5s`, `<5s`, `<+Inf`. Tiap request yang selesai, lo masukin ke **semua ember yang batasnya lebih besar dari durasinya**. Jadi bucket itu **kumulatif**.

Di `/metrics`, satu histogram beranak jadi tiga:

```
http_server_request_duration_seconds_bucket{method="GET",route="/orders",le="0.1"} 1200
http_server_request_duration_seconds_bucket{method="GET",route="/orders",le="+Inf"} 1284
http_server_request_duration_seconds_sum{...} 42.7
http_server_request_duration_seconds_count{...} 1284
```

Dua hal yang sering bikin junior salah:

1. **`le` itu label, bukan pilihan.** Karena `le` jumlahnya 11 per kombinasi label lain, histogram itu *mahal*. Satu histogram dengan 3 label nilai 10 = 30 deret waktu; dengan 11 bucket = 330 deret.
2. **Persentil wajib dihitung dari agregasi bucket, bukan dari hasil quantile per instance.**

```promql
# BENAR: jumlahkan bucket DULU, baru hitung quantile
histogram_quantile(0.95,
  sum by (le, route) (rate(http_server_request_duration_seconds_bucket[5m]))
)

# SALAH: quantile per instance lalu dirata-ratakan — hasilnya nggak bermakna
avg(histogram_quantile(0.95, rate(http_server_request_duration_seconds_bucket[5m])))
```

`histogram_quantile()` itu **aproksimasi**. Dia menganggap nilai di dalam satu bucket tersebar **seragam (linear interpolation)**. Jadi kalau bucket lo jarang (misal cuma `0.1` dan `5`), p95 lo bisa jauh dari kenyataan. Makanya bucket di lab ini dibikin rapat di area yang penting (5ms–100ms). Dan kalau quantile lo jatuh di bucket `+Inf`, hasilnya nggak ada artinya — berarti bucket terbesarmu kelewat kecil.

Rata-rata latency (sering lebih jujur buat kapasitas) dihitung dari `_sum`/`_count`:

```promql
sum by (route) (rate(http_server_request_duration_seconds_sum[5m]))
  / sum by (route) (rate(http_server_request_duration_seconds_count[5m]))
```

### 2.4 Summary — kuantil dihitung di aplikasi

Summary ngitung kuantil **di proses aplikasi**, lalu diekspos sebagai `<nama>{quantile="0.95"}`. Kedengarannya enak, tapi:

- Kuantilnya **nggak bisa diagregasi**. Rata-rata dari p95 lima instance bukan p95 gabungan — secara matematis nggak valid.
- Lo harus milih kuantil **sebelum** deploy. Butuh p99 padahal cuma definisiin p95? Deploy ulang.
- Biayanya dibayar di aplikasi (CPU + memori untuk sliding window).

Kesimpulan praktis: **di Prometheus, pakai histogram.** Summary kepakai cuma kalau lo tahu betul bahwa lo nggak akan pernah perlu agregasi lintas instance.

### 2.5 Scrape, interval, dan staleness

Prometheus itu **pull**: dia yang datang ke `/metrics` lo tiap `scrape_interval` (di lab: 15 detik). Konsekuensinya:

- Aplikasi lo nggak perlu tahu alamat Prometheus. Cukup buka endpoint `/metrics`.
- Kalau aplikasi mati, Prometheus cuma lihat target **DOWN** — dan metrik `up{job="api"}` jadi `0`. Ini alert paling murah dan paling penting (minggu 10 & 12).
- Kalau sebuah deret waktu **berhenti muncul** (misal endpoint nggak pernah dipanggil lagi), Prometheus nandain **stale** setelah 5 menit dan `rate()` berhenti menghitungnya. Jadi "nggak ada data" itu beda dengan "nol".

Kalau lo mau tahu apakah scrape-nya sehat, jangan lihat `up` doang:

```promql
scrape_duration_seconds{job="api"}      # berapa lama satu kali scrape
scrape_samples_scraped{job="api"}       # berapa deret waktu yang ditarik per scrape
```

Kalau `scrape_duration_seconds` mendekati `scrape_interval`, lo sebentar lagi punya masalah.

### 2.6 Label dan kardinalitas — pembunuh nomor satu

Satu kombinasi unik `nama_metrik + label` = **satu deret waktu**. Dokumentasi resmi Prometheus bilang begini: *"every unique combination of key-value label pairs represents a new time series, which can dramatically increase the amount of data stored. Do not use labels to store dimensions with high cardinality (many different label values), such as user IDs, email addresses, or other unbounded sets of values."*

**Label yang dilarang:**

| Dilarang | Kenapa | Ganti pakai |
| --- | --- | --- |
| `user_id`, `email`, `session_id` | tak terbatas, tiap user baru = deret baru | jangan disimpan di metrik; simpan di log |
| `path` mentah (`/orders/8821`) | tiap ID = deret baru | `route` = pola (`/orders/{id}`) |
| `request_id`, `trace_id` | sekali pakai | field di log (minggu 11) |
| `error_message` | isinya bebas dan panjang | `error_type` / `status` |
| timestamp, IP klien | nyaris unik tiap request | tingkatkan agregasinya |

Di `app/main.go` ini udah dikerjain dengan benar: `route := r.Pattern` (pola dari ServeMux), dan kalau kosong jadi `"unmatched"` — **bukan** `r.URL.Path`. Itu satu baris kode yang nyelamatin Prometheus lo.

Cara cek dari dalam Prometheus:

```promql
topk(10, count by (__name__) ({__name__=~".+"}))        # 10 metrik paling "gemuk"
count(count by (route, status) (http_server_requests_total))  # jumlah deret metrik ini
```

### 2.7 PromQL yang lo pakai tiap hari

```promql
# 1. Request rate per route (requests/detik)
sum by (route) (rate(http_server_requests_total[5m]))

# 2. Error ratio 0..1 (ini calon SLI minggu 12)
sum by (route) (rate(http_server_requests_total{status=~"5.."}[5m]))
  / sum by (route) (rate(http_server_requests_total[5m]))

# 3. p95 latency
histogram_quantile(0.95,
  sum by (le, route) (rate(http_server_request_duration_seconds_bucket[5m])))

# 4. Saturasi
http_server_requests_in_flight

# 5. Target hidup?
up{job="api"}
```

Satu jebakan klasik: kalau traffic kecil, hasil pembagian nomor 2 bisa `NaN` atau melonjak gila-gilaan — karena pembaginya nol (nggak ada request) atau cuma beberapa request. Makanya minggu 12 lo bakal nambahin aturan khusus buat **low-traffic service**.

### 2.8 Native histogram di Prometheus 3.x

Prometheus 3.0 rilis November 2024, dan jalur LTS yang dipakai lab ini adalah **3.13.x**. Native histogram — histogram yang disimpan sebagai satu sampel komposit, bukan puluhan deret float — **stabil sejak 3.8.0**, tapi **nggak aktif default**.

Lo harus nyalakan sendiri per scrape job:

```yaml
scrape_configs:
  - job_name: api
    scrape_native_histograms: true     # WAJIB kalau mau native histogram
    static_configs:
      - targets: ["api:8080"]
```

Itu udah ada di `observability/prometheus/prometheus.yml` lab ini. Kenapa repot-repot? Karena native histogram:

- jauh lebih hemat (satu deret, bukan N bucket),
- resolusinya dinamis, jadi lo nggak perlu menebak bucket di awal,
- **selalu bisa diagregasi**, walau resolusinya beda.

Kalau lo baca metriknya lewat `curl` dan lihat ada `# TYPE ... histogram` tapi nggak pernah muncul `_bucket`, cek dulu flag ini. Tanpa `scrape_native_histograms: true`, yang lo dapet ya histogram klasik biasa.

---

## 3. Mekanisme di balik layar

Alur lengkap dari aplikasi sampai alert bunyi:

1. **Instrumentasi (di aplikasi).** `main.go` bikin tiga metrik: counter `http_server_requests_total`, histogram `http_server_request_duration_seconds`, gauge `http_server_requests_in_flight`. Tiap kali ada request, middleware `instrument()` naikin in-flight, catat waktu mulai, panggil handler, lalu `Observe()` durasinya dan `.Inc()` counter-nya. Semua ini **di memori proses** — nggak ada yang dikirim ke mana pun.
2. **Expose.** `mux.Handle("GET /metrics", promhttp.Handler())`. Saat endpoint ini dipanggil, `client_golang` **merender** isi registry jadi teks (format OpenMetrics/Protobuf). Jadi `/metrics` itu cuma *foto kondisi saat ini*, bukan file yang disimpan.
3. **Service discovery + scrape.** Tiap 15 detik, Prometheus baca `prometheus.yml`, cari target (`api:8080`), lalu HTTP GET ke `/metrics`. Hasilnya ditambahin label `job="api"` dan `instance`. Kalau gagal → `up = 0`.
4. **Ingest ke TSDB.** Sampel masuk ke **head block** (di memori, `/prometheus`), yang tiap 2 jam di-compact dan ditulis ke blok permanen di disk (retensi lab: 15 hari lewat `--storage.tsdb.retention.time=15d`). Deret waktu diindeks oleh nama metrik + label.
5. **PromQL.** Waktu lo ngetik query di UI, Prometheus: (a) cari deret yang cocok sama selector, (b) ambil sampel di rentang waktu yang lo minta, (c) jalanin fungsi (`rate`, `sum by`, `histogram_quantile`) per titik waktu. Hasilnya bisa *instant vector* (satu nilai per deret) atau *range vector* (seri waktu).
6. **Rule evaluation.** `evaluation_interval: 15s` bikin Prometheus ngejalanin rule file (`rules/*.yml`) secara periodik. Recording rule nulis hasilnya balik ke TSDB (biar query berat nggak dihitung ulang tiap kali dashboard dibuka). Alerting rule ngecek kondisi; kalau kondisi true terus selama `for:`, statusnya `pending` → `firing`.
7. **Alertmanager / notifikasi.** Prometheus nggak kirim notifikasi sendiri — dia lempar alert ke Alertmanager (atau, kalau lo pakai Grafana-managed alerting minggu 10, Grafana yang ngevaluasi).

Yang penting lo sadari: **antara "user ngerasain masalah" dan "alert bunyi" ada jeda**. Minimal `scrape_interval`, lalu `evaluation_interval`, lalu `for:`. Minggu 10 lo bakal ngukur jeda ini dengan stopwatch.

---

## 4. Glosarium mini

| Istilah | Arti dengan bahasa sederhana | Contoh konkret |
| --- | --- | --- |
| **Time series** | Satu angka yang dicatat berulang-ulang bersama waktu | `http_server_requests_total{route="/orders",status="200"}` nilainya 1284 lalu 1290 |
| **Label** | "Kartu identitas" yang nempel di metrik buat memisahkan data | `method="GET"`, `route="/orders"`, `status="500"` |
| **Kardinalitas** | Jumlah kombinasi label unik — makin banyak, makin berat | `status` (5 nilai) vs `user_id` (jutaan nilai) |
| **Scrape** | Prometheus "datang dan ambil" metrik dari aplikasi | Tiap 15 detik GET `http://api:8080/metrics` |
| **Sample** | Satu titik data: pasangan (timestamp, nilai) | `@1730000000 → 1284` |
| **Instant vector** | Hasil query berupa satu nilai per deret, untuk satu waktu | `up{job="api"}` → `1` |
| **Range vector** | Kumpulan sampel dalam rentang waktu; cuma bisa dipakai fungsi tertentu | `rate(http_server_requests_total[5m])` |
| **Bucket (`le`)** | Ember "≤ batas ini ada berapa" | `le="0.1"` = berapa request yang ≤ 100 ms |
| **Staleness** | Deret yang berhenti di-update dianggap mati setelah 5 menit | Endpoint nggak dipanggil lagi, `rate()` jadi kosong |
| **Recording rule** | Query berat yang hasilnya disimpan sebagai metrik baru | `job:slo_errors_per_request:ratio_rate5m` |

---

## 5. Bahan bacaan wajib

| # | Sumber (link) | Bagian mana yang dibaca & apa yang dicari | Estimasi |
| --- | --- | --- | --- |
| 1 | [Prometheus — Metric types](https://prometheus.io/docs/concepts/metric_types/) | Seluruh halaman. Cari perbedaan counter/gauge/histogram/summary, dan suffix `_bucket`, `_sum`, `_count`, `{quantile="..."}`. Ini sumber keempat analogi di atas. | 25 mnt |
| 2 | [Prometheus — Metric and label naming](https://prometheus.io/docs/practices/naming/) | Bagian **Labels**, terutama kotak **CAUTION** soal kardinalitas. Catat contoh label yang dilarang. | 15 mnt |
| 3 | [Prometheus — Histograms and summaries](https://prometheus.io/docs/practices/histograms/) | Bagian "Apdex score", "Quantiles", dan "Errors of quantile estimation". Cari kenapa quantile itu aproksimasi dan kenapa bucket perlu rapat di area penting. | 30 mnt |
| 4 | [Prometheus — Query functions](https://prometheus.io/docs/prometheus/latest/querying/functions/) | Cari entri `rate()`, `increase()`, `histogram_quantile()`, `topk()`. Perhatiin syarat `le` dan asumsi interpolasi linear. Jangan dihafal semua, cukup yang dipakai lab. | 30 mnt |
| 5 | [Prometheus — Native histograms](https://prometheus.io/docs/specs/native_histograms/) | Cari cara mengaktifkan: `scrape_native_histograms: true`. Cek apakah default-nya aktif atau nggak. | 15 mnt |
| 6 | [Prometheus — Configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/) | Cari entri `scrape_native_histograms` di bagian `<scrape_config>`, plus `scrape_interval` dan `evaluation_interval` di `global`. | 20 mnt |

---

## 6. Self-check (harus bisa dijawab tanpa buka Google)

1. Apa beda counter dan gauge, dan apa fungsi PromQL yang wajib dipakai buat masing-masing?
2. Kenapa `rate(http_server_requests_total[5m])` lebih berguna daripada nilai mentah counter-nya? Apa yang terjadi kalau aplikasi restart?
3. Kenapa `histogram_quantile()` hasilnya cuma aproksimasi? Apa yang bikin aproksimasinya makin meleset?
4. Kamu punya 2 instance API, masing-masing sudah menghitung p95 sendiri. Kenapa rata-rata dari dua angka itu bukan p95 gabungan? Apa yang harus dilakukan?
5. Sebutkan 3 label yang dilarang di Prometheus dan apa penggantinya.
6. Dengan `scrape_interval: 15s`, apa range selector terkecil yang masih waras dipakai, dan kenapa?
7. Apa yang terjadi kalau lo nggak nyalain `scrape_native_histograms: true` di Prometheus 3.x? Apakah metriknya hilang?

---

## 7. Yang sering salah paham

- **"histogram_quantile itu nilai persentil asli."** → Salah. Dia interpolasi linear di dalam bucket, dengan asumsi nilai tersebar seragam. Makin lebar bucket, makin ngaco. Dia juga nggak bisa kasih tahu lo nilai di luar bucket terbesar.
- **"Bisa rata-ratakan quantile dari tiap instance buat dapet quantile global."** → Salah, dan ini jebakan paling sering. Yang benar: `sum by (le, ...) (rate(..._bucket[5m]))` **dulu**, baru `histogram_quantile()`.
- **"Counter nilainya naik terus, berarti traffic naik terus."** → Nilai mentah counter selalu naik selama proses hidup. Yang naik-turun itu `rate()`-nya.
- **"Nambahin satu label kan cuma nambahin satu kolom."** → Enggak. Tiap *nilai* baru = deret waktu baru. Satu label `user_id` di 100 ribu user = 100 ribu deret waktu per metrik, dan kalau metriknya histogram ×11 bucket.
- **"Native histogram aktif otomatis di Prometheus 3.x."** → Nggak. Stabil sejak 3.8.0 tapi harus `scrape_native_histograms: true` di scrape config.
- **"rate() dengan range pendek lebih responsif, jadi lebih bagus."** → Range `[30s]` dengan interval 15s cuma dapet 2 sampel; sedikit saja scrape telat, hasilnya nol atau anjlok. Pakai minimal 4× interval.
- **"Kalau target DOWN, metriknya jadi 0."** → Yang jadi 0 cuma `up`. Metrik lain **hilang** (stale), dan itu beda perlakuan di query: pembagian dengan pembagi kosong menghasilkan kosong, bukan nol.
