# Minggu 8 · Production Readiness & Reliability Baseline — Bahan Bacaan

> Dokumen ini **berdiri sendiri**. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham konsepnya. Baru setelah itu kerjakan `lab.md`.

## 1. Kenapa minggu ini penting

Lima minggu lo bangun: aplikasi (3), proxy (4), CI/CD (5), deploy (6), backup (7). Minggu 8 adalah tempat lo **menyatukan semuanya** dan menjawab satu pertanyaan jujur: *"kalau sistem ini rusak jam 3 pagi, apakah orang yang lagi jaga bisa ngerti dan benerin tanpa nanya-nanya?"*

Ini bukan dokumen formalitas. Production Readiness Review (PRR) adalah pertanyaan yang bakal lo dapat di interview SRE level mana pun, dan yang membedakan "engineer yang bisa bikin fitur" dari "engineer yang bisa dijaga sistemnya". PRR gagal bukan karena nggak ditulis — tapi karena ditulis sekali lalu dibiarin basi, atau cuma checklist kosong tanpa bukti.

## 2. Konsep inti

### 2.1 Production Readiness Review (PRR)
PRR = audit apakah sistem siap "dioperasikan", bukan cuma "jalan di laptop lo". Lima kategori wajib: **Security, Availability, Observability, Recovery, Deployment**. Setiap item punya status `TODO` / `PARTIAL` / `DONE` — dan `DONE` cuma boleh kalau **sudah diuji**, bukan cuma sudah ditulis.

### 2.2 Dependency mapping
Gambar siapa bergantung ke siapa. Di lab: API → Postgres, API → Redis, Client → Nginx → API. Kalau Postgres mati, endpoint tulis ikut mati. Lo harus tahu rantai ini sebelum bisa milih mana yang SPOF.

### 2.3 Failure Mode and Effects Analysis (FMEA)
Untuk tiap komponen, jawab: gimana dia gagal, apa dampaknya, gimana lo **mendeteksinya** (ini kuncinya — tanpa cara deteksi, item itu nggak bisa ditindaklanjuti), dan mitigasinya. FMEA tanpa kolom "gimana deteksinya" cuma daftar ketakutan, bukan rencana.

### 2.4 Single Point of Failure (SPOF)
Satu komponen yang kalau mati = layanan mati total. Di lab: Postgres, Redis, Nginx masing-masing SPOF. PRR mewajibkan lo **menulis eksplisit** SPOF yang diterima (dan alasannya), bukan pura-pura nggak ada.

### 2.5 Operational checklist & "satu perintah"
Syarat tertinggi minggu ini: **dari repo kosong, satu perintah sampai sistem jalan**. Kalau lo butuh 5 langkah manual dari ingatan, reproduksinya gagal — dan itu bug. PRR juga butuh pemilik on-call tertulis.

### 2.6 Recovery yang teruji, bukan cuma tertulis
Minggu 7 lo bikin backup. Minggu 8 lo buktikan: reboot mesin → sistem balik sendiri tanpa campur tangan. Catat waktunya: reboot → `/readyz` balas 200 = berapa detik?

## 3. Mekanisme di balik layar

Waktu lo jalanin `docker compose up -d --build --wait` dari repo bersih:
1. Compose baca `compose.yaml`, build image `api` dari `Dockerfile`.
2. Buat network bridge, mount volume (data Postgres persist).
3. Jalankan `postgres` & `redis`, tunggu `service_healthy` (lewat `depends_on: condition`).
4. Jalankan `api` setelah dependency sehat, jalankan healthcheck.
5. Jalankan `nginx`.
6. `curl /health` dan `/readyz` → 200.
7. Saat reboot: container otomatis `restart: unless-stopped` (atau `docker compose restart`), volume bikin data nggak ilang.

## 4. Glosarium mini

| Istilah | Arti sederhana | Contoh |
| --- | --- | --- |
| PRR | Audit kesiapan operasional | 5 kategori di `docs/production-readiness.md` |
| FMEA | Analisis cara gagal & dampak | "Postgres lambat → tulis gagal" |
| SPOF | Titik yang kalau mati = layanan mati | Database tunggal |
| RTO | Waktu pulih setelah mati | reboot → readyz 200 = 45 detik |
| RPO | Batas data yang boleh hilang | backup tiap 1 jam → RPO 1 jam |
| On-call | Orang yang jaga kalau rusak | tertulis, bukan "siapa aja" |

## 5. Bahan bacaan wajib

| # | Sumber | Apa yang dibaca | Estimasi |
| --- | --- | --- | --- |
| 1 | [Google SRE — Evolving SRE Engagement Model (PRR)](https://sre.google/sre-book/evolving-sre-engagement-model/) | Konsep PRR | 20 mnt |
| 2 | [Launch Coordination Checklist](https://sre.google/sre-book/launch-checklist/) | Checklist peluncuran | 15 mnt |
| 3 | [Google SRE — Failure Modes and Root Cause Analysis](https://sre.google/sre-book/failure-modes/) | FMEA & analisis kegagalan | 20 mnt |
| 4 | `docs/production-readiness.md` (di repo ini) | Template 5 kategori + dependency map | 10 mnt |
| 5 | [Docker Compose `depends_on`](https://docs.docker.com/compose/how-tos/startup-order/) | `condition: service_healthy` | 10 mnt |

## 6. Self-check (harus bisa dijawab tanpa buka Google)

1. Apa bedanya PRR sama sekadar "deploy sukses"?
2. Kenapa FMEA butuh kolom "gimana deteksinya"?
3. Apa itu SPOF dan sebutkan 3 di lab lo?
4. Kenapa "satu perintah dari repo kosong" itu syarat penting?
5. Kenapa risiko harus ditulis eksplisit, bukan disembunyikan?

## 7. Yang sering salah paham

| Mitos | Kenyataan |
| --- | --- |
| "PRR = isi checklist sekali" | Dokumen yang nggak direvisi tiap perubahan = basi & berbahaya. |
| "SPOF nggak ada di sistem bagus" | SPOF hampir pasti ada; yang bener itu **nyatain** & terima risikonya. |
| "Backup cukup" | Backup = harapan. Restore **teruji** = kepastian. |
| "Dependency pihak ketiga nggak perlu masuk peta" | Auth/payment gateway paling sering bikin outage justru karena lupa dicatat. |
| "On-call siapa aja" | Tanpa pemilik eksplisit, jam 3 pagi nggak ada yang gerak. |
