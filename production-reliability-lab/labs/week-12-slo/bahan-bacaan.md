# Minggu 12 · SLI, SLO, Error Budget, dan Incident Response — Bahan Bacaan

> "SLO itu alat buat berantem pakai angka, bukan pakai perasaan. Budget habis = rilis dihentikan. Titik."

Dokumen ini berdiri sendiri. Baca dari atas ke bawah, kerjakan `lab.md` setelahnya.

---

## 1. Kenapa minggu ini penting

Minggu 9–11 lo sudah punya metrics, dashboard, alert, log, dan trace. Tapi semua itu cuma alat ukur — minggu ini lo belajar **ngapain sama angkanya**. Tanpa SLI/SLO, "boleh nggak kita rilis fitur baru?" jawabnya tergantung siapa yang berani ngoceh paling keras di meeting. Dengan error budget, jawabnya jadi **angka**: budget masih banyak → gas; budget habis → freeze rilis, kerjaan pindah ke reliabilitas. Itu perubahan budaya, bukan cuma teknis.

Bayangin error budget kayak **kuota bensin di taksi**. SLO 99,5% berarti lo diizinin "habisin" 0,5% perjalanan nyasar — 5.000 request gagal per 1 juta. Selama masih ada sisa bensin, lo bebas muter-muter (eksperimen, rilis agresif). Tapi kalau bensin habis, lo harus berhenti dan isi dulu, bukan lanjut jalan sampai mogok di tengah jalan. Burn rate itu **kecepatan lo ngabisin bensin itu**. Kalau lo ngabisin 14,4× lebih cepat dari rencana, 2% budget 30 hari habis dalam 1 jam — itu yang bikin alert `HighErrorBudgetBurn` page lo.

Masalah kedua yang minggu ini beresin: **insiden ditangani pakai panik, bukan prosedur**. Deteksi lambat, severity telat ditentukan, mitigasi dicampur dengan cari root cause, postmortem cuma curhatan yang nyalahin orang. Minggu ini lo praktik alur: deteksi → triage → **mitigasi dulu** → diagnosis → pemulihan → postmortem **blameless**.

---

## 2. Konsep inti

### 2.1 SLI — angka yang diukur dari sisi user
SLI (Service Level Indicator) adalah proporsi request valid yang berhasil. Contoh lab: "persentase request valid yang balas non-5xx" (availability) dan "persentase request valid yang selesai < 300ms" (latency).

**Penting:** diukur **dari sisi user/proxy**, bukan dari metrik internal app. Kalau app crash sebelum nyatat metrik, request itu nggak kehitung sebagai error — padahal itu kegagalan terburuk. Proxy menutup celah itu.

### 2.2 SLO — target dari SLI
SLO (Service Level Objective) = target SLI. Lab: **availability 99,5%** (30 hari rolling), **latency 95% < 300ms**. SLO bukan janji ke user — itu kompas internal tim.

### 2.3 SLA — janji ke user (ada konsekuensi)
SLA (Service Level Agreement) punya denda/credit kalau melanggar. Biasanya lebih longgar dari SLO (SLO 99,5%, SLA 99,9% dengan credit). Jangan campur: SLO buat tim, SLA buat kontrak.

### 2.4 Error budget & burn rate
Error budget = sisa "jatah gagal" = `1 - SLO`. Burn rate = seberapa cepat lo ngabisin itu:
```
burn rate = (error ratio terukur) / (1 - SLO)
```
Contoh: SLO 99,5% → budget 0,5%. Error ratio 1 jam = 3,6% → burn rate 7,2×.
Dengan 7,2×, 5% budget 30 hari habis dalam 6 jam.

### 2.5 Multiwindow multi-burn-rate (alerting)
Pakai dua jendela (pendek + panjang) biar alert **cepat tapi nggak false positive** di traffic rendah. Dari Google SRE Workbook:

| Jendela panjang | Pendek | Burn rate | Budget habis | Aksi |
| --- | --- | --- | --- | --- |
| 1 jam | 5 mnt | **14,4** | 2% / 1 jam | Page |
| 6 jam | 30 mnt | **6** | 5% / 6 jam | Page |
| 3 hari | 6 jam | **1** | 10% / 3 hari | Ticket |

Di `rules/slo.yml`: `14.4 * 0.005` (page), `6 * 0.005` (page), `1 * 0.005` (ticket). Recording rule dulu (`job:slo_errors_per_request:ratio_rate*`), baru alerting rule.

### 2.6 Error budget policy
Ubah "boleh rilis?" jadi angka:
- > 50% tersisa → rilis bebas.
- 25–50% → rilis normal, risiko butuh review ekstra.
- < 25% → rilis fitur dihentikan, cuma perbaikan reliabilitas.
- Habis → **freeze**, wajib 1 postmortem + minimal 1 item reliabilitas.

### 2.7 Incident response: severity & alur
Tetapkan **severity di 5 menit pertama**. Alur: deteksi → triage → **mitigasi dulu** → diagnosis → pemulihan → postmortem (48 jam). Urutan 3 sebelum 4 sengaja: pulihkan dulu, analisis belakangan. IC **ngarahin, bukan debugging**.

### 2.8 Blameless postmortem
Fokus ke **sistem yang memungkinkan**, bukan orang. Template di `postmortems/incident-001.md`: impact, timeline, root cause (1 kalimat jujur), 5-whys, what went well/wrong, action item ber-owner & tenggat. Tanpa owner+tenggat = harapan, bukan tindakan.

---

## 3. Mekanisme di balik layar

1. **Scrape**: Prometheus tarik `http_server_requests_total` dari `api`.
2. **Recording rule**: tiap 30s hitung `job:slo_errors_per_request:ratio_rate{5m,30m,1h,6h}` dan `job:slo_latency_p95`.
3. **Alerting rule**: bandingkan ratio dengan `burn_rate × budget`; kalau dua jendela patri benar selama `for:`, firing.
4. **Notification**: page (on-call) / ticket (issue) sesuai severity.
5. **Insiden**: deteksi → severity → mitigasi (rollback/scale) → diagnosis → pulih (verifikasi angka).
6. **Postmortem**: blameless, dalam 48 jam, action item ber-owner.

---

## 4. Glosarium mini

| Istilah | Arti singkat |
| --- | --- |
| SLI | Indikator level service (diukur dari user) |
| SLO | Target SLI (kompas internal) |
| SLA | Janji ke user (ada denda) |
| Error budget | Sisa jatah gagal = 1 − SLO |
| Burn rate | Kecepatan ngabisin budget |
| Multiwindow burn-rate | Alert 2 jendela: cepat & anti-false-positive |
| Recording rule | Hitung metrik turunan dulu, reuse |
| Severity | SEV1/2/3, ditetapkan 5 menit pertama |
| Incident Commander | Yang ngarahin, bukan debugging |
| Mitigasi dulu | Pulihkan sebelum cari root cause |
| Blameless | Fokus sistem, bukan orang |
| 5-whys | Telusuri akar proses, bukan akar orang |

---

## 5. Bahan bacaan wajib (primer)

1. Google SRE Workbook — Implementing SLOs: https://sre.google/workbook/implementing-slos/
2. Google SRE Workbook — Alerting on SLOs (multiwindow burn-rate): https://sre.google/workbook/alerting-on-slos/
3. Google SRE Book — Error budgets: https://sre.google/sre-book/embracing-risk/
4. Prometheus — Recording rules: https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
5. Prometheus — Checking rules with promtool: https://prometheus.io/docs/prometheus/latest/command-line/promtool/#checking-rules
6. Grafana — Incident response (on-call): https://grafana.com/docs/oncall/latest/

---

## 6. Self-check (harus bisa dijawab tanpa Google)

1. Bedanya SLI, SLO, SLA? Mana yang punya konsekuensi kontrak?
2. Kenapa SLO diukur dari proxy, bukan metrik internal app?
3. Hitung burn rate: SLO 99,5%, error ratio 1 jam = 3,6%.
4. Kenapa alert burn-rate pakai dua jendela (multiwindow)?
5. Kenapa mitigasi dilakukan SEBELUM cari root cause?
6. Apa itu blameless postmortem dan kenapa action item butuh owner?

---

## 7. Yang sering salah paham

- **"SLO diukur dari metrik internal aplikasi."** Salah. App yang crash sebelum nyatat = error terburuk nggak kehitung. Ukur dari sisi user/proxy.
- **"Alert statis 'error > 1%' cukup."** Di traffic rendah, 1 request gagal/jam = burn rate 1000×. Butuh burn-rate + multiwindow.
- **"Service traffic rendah nggak perlu aturan khusus."** Justru butuh — persentase gampang meledak oleh 1 request.
- **"Postmortem = nyari siapa yang salah."** Salah. Blameless → fokus sistem. Nyebut nama orang = postmortem gagal.
- **"Severity bisa ditentukan belakangan."** Telat tentuin severity = respons kacau. Tetapkan di 5 menit pertama.
- **"Recording rule optional."** Tanpa recording rule, alert burn-rate manggil `rate()` berulang & lambat. Reuse dulu.
