# Menjadi SRE dalam 16 Minggu

> Dokumen ini adalah **versi yang diperdalam** dari roadmap 16 minggu yang lo punya.  
> Roadmap asli udah bagus: ritme jelas, acceptance criteria jelas, struktur repo jelas.  
> Yang kurang biasanya cuma satu hal: **bagian "gue harus ngapain tepatnya, dan gimana tahu kalau salah"**.  
> Dokumen ini ngisi bagian itu.
>
> **Ditulis untuk:** backend engineer yang sudah nyaman dengan Python, Go, Docker, database, Redis, dan debugging aplikasi.  
> **Ditulis pada:** 12 Oktober 2026. Semua versi ditandai: **[2025]** = milestone yang lo minta, **[2026]** = versi aktual hari ini (jangan sampai lo install yang udah pensiun).

---

## Cara pakai dokumen ini

1. **Jangan dibaca sekali duduk.** Baca minggu yang lagi lo kerjain, minggu depan baca lagi.
2. **Setiap minggu punya 4 "pintu keluar".** Kalau 4 acceptance criteria di bawah nggak lolos, minggu itu **belum selesai**. Lanjut minggu berikutnya itu boleh, tapi jangan centang yang ini.
3. **Bukti > perasaan.** "Kayaknya udah paham" bukan bukti. Log, screenshot, angka, dan commit itu bukti.
4. **Rusakkan dengan sengaja.** Hampir semua minggu punya sesi "Hari 3: Eksperimen kegagalan". Ini bagian yang paling berharga — karena insiden nyata nggak pernah datang dalam kondisi lo lagi siap.
5. **Satu repo, bukan 16 repo.** Semua masuk ke `production-reliability-lab/`. Reviewer (dan calon interviewer) bisa ngikutin progres lo dari Linux sampai incident response tanpa buka 16 tab.

### 5 kebiasaan yang sebenarnya dibentuk 16 minggu ini

Roadmap ini kelihatannya soal tools. Sebenernya soal **cara berpikir**:

| Kebiasaan                              | Kelihatannya di minggu ke-                      |
| -------------------------------------- | ----------------------------------------------- |
| **Ukur** dulu sebelum opini            | 2 (latency fase), 9 (metrics), 16 (load test)   |
| **Deteksi** sebelum user ngeluh        | 10 (alerting), 12 (burn rate)                   |
| **Diagnosis** pakai bukti berjenjang   | 1 (Linux), 11 (trace), 16 (chaos)               |
| **Mitigasi** dulu, analisis belakangan | 4 (failure matrix), 6 (rollback), 12 (incident) |
| **Cegah** yang sama terulang           | 8 (PRR), 12 (postmortem), 16 (capacity)         |

Kalau habis 16 minggu lo cuma hapal perintah `kubectl` tapi masih nebak-nebak pas ada masalah, berarti 128 jam itu kepakai buat nonton, bukan buat latihan.

---

## Peta besar

| Bulan | Minggu | Tema                                | Yang bikin lo beda dari "backend engineer yang bisa deploy"            |
| ----- | ------ | ----------------------------------- | ---------------------------------------------------------------------- |
| 1     | 1–4    | Linux, networking, container, proxy | Bisa jawab "kenapa lambat/mati" pakai bukti, bukan "coba restart dulu" |
| 2     | 5–8    | CI/CD, secret, backup, readiness    | Bisa rilis, rollback, dan pulihin data tanpa panik                     |
| 3     | 9–12   | Metrics, alert, log, trace, SLO     | Bisa bilang "layanan kita sehat 99,7%, dan ini buktinya"               |
| 4     | 13–16  | IaC, cloud, K8s, load test, chaos   | Bisa ukur kapasitas dan buktiin sistem bisa pulih                      |

---

## Setup awal: stack dan versi

Roadmap lo udah milih stack yang tepat. Catatan tambahan buat 2025–2026:

| Komponen      | Pilihan                     | Catatan versi                                                                                                                                                                                                         |
| ------------- | --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Aplikasi      | Go HTTP API                 | Go 1.25 rilis Agustus **[2025]**; versi stabil terbaru hari ini **Go 1.27.x**. Pakai minimal 1.22+ biar dapet routing pattern di `ServeMux`, dan manfaatin `log/slog` (bawaan sejak Go 1.21)                          |
| Dependency    | PostgreSQL + Redis          | **PostgreSQL 18 GA 25 Sep 2025** — `initdb` sekarang default **data checksum ON**, ada AIO (`io_method`), `md5` deprecated. PG 17/18 dua-duanya aman buat lab                                                         |
| Runtime       | Docker Compose              | Compose v1 (`docker-compose` Python) **sudah dihapus**. Pakai `docker compose` (v2+). Compose sekarang sudah v5.x **[2026]**                                                                                          |
| Reverse proxy | Nginx                       | Stable 1.28/1.29 di 2025; hari ini stable 1.30.x. Mulai **nginx 1.29.7**, `keepalive` di blok `upstream` dan `proxy_http_version 1.1` **default ON**                                                                  |
| CI/CD         | GitHub Actions              | `actions/setup-go` v7, `docker/build-push-action` v7, `actions/attest` v4 (SLSA attestation)                                                                                                                          |
| Metrics       | Prometheus                  | **Prometheus 3.0 rilis 14 Nov 2024** — UI baru, Remote Write 2.0, UTF-8 di nama metric, OTLP receiver. Rilis 2025: 3.2 → 3.8. Sekarang 3.15.x, jalur LTS **3.13.x**                                                   |
| Dashboard     | Grafana                     | Grafana 12.0 rilis 5 Mei **[2025]**; sekarang 13.x                                                                                                                                                                    |
| Logs          | Loki + **Grafana Alloy**    | **PENTING: Promtail EOL 2 Maret 2026.** Jangan mulai lab pakai Promtail — langsung Alloy                                                                                                                              |
| Tracing       | OpenTelemetry               | OTLP sudah **stable untuk traces, metrics, dan logs**. Semconv HTTP dan DB client sudah stable                                                                                                                        |
| Load test     | k6                          | **k6 v1.0.0 rilis 6 Mei 2025** (native TypeScript, SemVer, 2 tahun support). Sekarang sudah **k6 v2.x** — pin versi di CI                                                                                             |
| IaC           | Terraform **atau** OpenTofu | Terraform berlisensi BUSL sejak 2023; **IBM selesaikan akuisisi HashiCorp 27 Feb 2025**. OpenTofu (MPL-2.0, Linux Foundation) makin banyak dipakai. **Buat lab, OpenTofu lebih bebas**; buat kerja, pahami dua-duanya |
| Cloud lab     | VPS / free tier             | **AWS Free Tier berubah 15 Juli 2025**: bukan lagi "12 bulan gratis", tapi kredit $100 (+$100) berlaku **6 bulan**. Pasang budget alert dari hari pertama                                                             |

**Saran gue:** mulai dengan **1 VPS Ubuntu 24.04 LTS** (~$5/bulan) + Docker. Lebih murah, lebih sedikit jebakan, dan cukup buat minggu 1–12. Pakai cloud managed service (RDS/Cloud SQL) **cuma** pas minggu 13–14, karena itu minggu yang memang ngajarin cloud.

---

# Bulan 1: Linux, Networking, dan Diagnosis

Tujuan bulan ini: berhenti jadi orang yang bilang "coba restart dulu".

---

## Minggu 1 · Linux observability dan troubleshooting

**Target:** bisa menjelaskan kenapa service lambat atau mati, pakai bukti.

### Kenapa minggu ini nomor satu

Karena hampir semua insiden produksi ujungnya cuma 3 pertanyaan:

1. Siapa yang makan CPU?
2. Siapa yang makan memori / kena limit?
3. Prosesnya masih hidup nggak, dan kalau hidup, kenapa nggak bisa diajak bicara?

Kalau lo nggak bisa jawab 3 ini dalam 5 menit, observability stack secanggih apa pun cuma jadi dekorasi.

### Hari 1 · Teori (2 jam)

Yang wajib lo pahami, dan kenapa:

| Konsep                        | Kenapa penting buat SRE                                                                                 |
| ----------------------------- | ------------------------------------------------------------------------------------------------------- |
| **Process vs thread**         | "CPU 100%" bisa jadi 1 thread nakal di proses yang kelihatannya normal. `top` biasa nyembunyiin ini     |
| **Signals (SIGTERM/SIGKILL)** | Bedanya minta berhenti dengan sopan vs ditembak. Ini nyambung langsung ke graceful shutdown di minggu 6 |
| **systemd**                   | Siapa yang nyalain ulang service, dan kenapa `journalctl` jadi sumber kebenaran                         |
| **Memory: free vs available** | "Memory penuh" paling sering salah diagnosis                                                            |
| **File descriptor**           | "Too many open files" itu bukan bug eksotis — ini koneksi yang nggak ditutup                            |
| **cgroup v2**                 | Ini yang bikin container kena OOM meski RAM host masih lega                                             |

Baca: [Linux man-pages](https://man7.org/linux/man-pages/) dan [cgroup v2 docs](https://docs.kernel.org/admin-guide/cgroup-v2.html).

Konsep yang **sering kelewat tapi sangat berguna**: **PSI (Pressure Stall Information)**. Ini jawaban buat pertanyaan "CPU-nya nggak 100%, tapi kok lambat?" PSI ngasih tahu berapa persen waktu proses *menunggu* CPU/memori/IO.

```bash
cat /proc/pressure/cpu
cat /proc/pressure/memory
cat /proc/pressure/io
# some avg10=0.00 avg60=0.00 avg300=2.31 total=1234567
```

### Hari 2 · Praktik (2 jam)

```bash
uname -a                             # kernel + arsitektur (Ubuntu 24.04.3 LTS bawa HWE 6.14 [2025])
uptime                               # load average 1/5/15 menit
free -h                              # lihat "available", BUKAN "free"
df -h && df -i                       # disk penuh dan inode habis itu dua masalah beda
ps aux --sort=-%cpu | head -10
ps aux --sort=-%mem | head -10
systemctl --failed
journalctl -p err -b --no-pager
```

Lalu nyalain API Go, dan **buktikan** hal-hal ini:

```bash
go run . &                           # atau ./sre-api
pgrep -af 'sre-api|go run'           # PID + command lengkap
PID=$(pgrep -f sre-api | head -1)
ss -lntp | grep 8080                 # siapa yang listen di 8080
lsof -p $PID | head -30              # file + koneksi yang dibuka proses ini
lsof -i :8080                        # kalau port bentrok, ini hakimnya
cat /proc/$PID/status | grep -E 'Threads|VmRSS'
ls /proc/$PID/fd | wc -l             # jumlah file descriptor
```

**Output yang harus lo simpan sebagai bukti:** `ss -lntp` nunjukin proses Go lo listen di 8080, dan `lsof -p $PID` nunjukin koneksi ke Postgres dan Redis.

### Hari 3 · Eksperimen (2 jam)

Tiga eksperimen wajib:

**Eksperimen A: CPU tinggi.**

```bash
# bikin proses CPU-intensive
yes > /dev/null & 
BAD=$!
top -H -p $BAD                       # lihat thread-level, bukan cuma proses
pidstat -urd -p $BAD 1 5             # CPU + memori + IO per proses, interval 1s
kill -TERM $BAD                      # minta berhenti sopan
# kalau bandel:
kill -KILL $BAD                      # paksa — dan pahami kapan ini jadi masalah
```

**Eksperimen B: Service gagal di systemd.** Bikin unit file sederhana, sengaja salahin (path binary salah, atau env yang nggak ada), lalu:

```bash
systemctl status sre-lab.service
journalctl -u sre-lab.service -n 100 --no-pager
systemctl --failed
```

**Eksperimen C: bedain 3 kondisi yang sering ketuker.**

1. **Proses mati** → `pgrep` kosong, `ss` nggak ada yang listen, `curl` → `connection refused`.
2. **Proses hidup, port nggak listen** → `pgrep` ada, `ss` kosong. Ini yang bikin orang bingung 30 menit.
3. **Port listen, tapi aplikasi nggak responsif** → `ss` ada, `curl` nggak pernah balas (hang). Biasanya thread pool / koneksi DB habis.

```bash
# pembuktiannya
curl -v --max-time 3 http://localhost:8080/health
echo "exit code: $?"                 # 7 = refused, 28 = timeout, 0 = ok, 52 = empty reply
```

Simpan tabel perbandingan 3 kondisi ini di `labs/week-01-linux/commands.md`. Ini salah satu artefak paling kepakai sepanjang karier lo.

### Hari 4 · Review (2 jam)

Bikin 3 runbook di `labs/week-01-linux/runbooks/`:

| File              | Gejala                      | Isi minimum                                                                     |
| ----------------- | --------------------------- | ------------------------------------------------------------------------------- |
| `high-cpu.md`     | load tinggi, request lambat | Cara nemu thread nakal <5 menit, keputusan kapan scale vs kapan perbaiki kode   |
| `oom-symptoms.md` | container restart mendadak  | Bukti kernel `oom-kill` vs `systemd-oomd`, beda OOM host vs OOM cgroup          |
| `service-down.md` | endpoint nggak bisa diakses | Decision tree: proses hidup? port listen? aplikasi responsif? dependency sehat? |

### Jebakan umum minggu 1

- **Load average tinggi ≠ CPU tinggi.** Bisa jadi proses numpuk nunggu I/O. Cek kolom `wa` di `vmstat` dan `b` (blocked).
- **"Memory penuh" palsu.** Linux pakai sisa RAM buat page cache. Lihat `MemAvailable`, dan `free -h` baris `buff/cache` bukan musuh lo.
- **OOM di container nggak selalu muncul di `dmesg` host.** Yang kena limit itu cgroup, jadi cek `memory.events` dan `systemd-oomd`.
- **`pidstat` iterasi pertama itu rata-rata sejak boot**, bukan kondisi sekarang. Tunggu baris ke-2.
- **`kill -9` sebagai solusi pertama.** Itu nggak ngasih aplikasi kesempatan nutup koneksi DB dengan rapi — dan bakal jadi bug misterius di minggu 6.

### Acceptance criteria minggu 1

1. Bisa nunjuk proses/thread penyebab CPU tinggi dalam **< 5 menit**, dengan bukti.
2. Bisa nemuin error dari system journal (`journalctl`) dan bedain error aplikasi vs error sistem.
3. Bisa bedain **service down** vs **service hidup tapi nggak nerima koneksi**, dan nulis decision tree-nya.
4. Bisa sebutin apa bedanya `kill -TERM` dan `kill -KILL`, serta kapan `-KILL` jadi keputusan yang buruk.

**Yang masuk GitHub:** `labs/week-01-linux/` berisi `README.md`, `notes.md`, `commands.md` (perintah + output tersanitasi), `runbooks/` (3 file), `evidence/`.

---

## Minggu 2 · Networking untuk backend engineer

**Target:** paham perjalanan request dari client ke aplikasi dan dependency-nya.

### Konsep yang harus nempel

Bayangin request itu kayak kurir:

1. **DNS** = nyari alamat dari nama. Gagal di sini = "alamatnya nggak ada".
2. **TCP handshake** = ketuk pintu, tunggu dibukain. Gagal = `connection refused` (pintu dikunci) atau **timeout** (nggak ada yang nyahut).
3. **TLS handshake** = saling kenal dan bikin kunci rahasia. Gagal = certificate/versi TLS nggak cocok.
4. **HTTP** = ngobrol. Gagal di sini = 4xx/5xx, maksudnya aplikasi **jawab**, tapi jawabannya jelek.
5. **Transfer body** = barangnya diantar.

Kerennya: `curl` bisa nunjukin **semua fase ini dalam angka**. Dan angka itu bikin debat "lelet di server atau di jaringan?" selesai dalam 10 detik.

### Hari 1 · Teori (2 jam)

Pelajari DNS, routing, TCP handshake, listening socket, connection timeout, HTTP, TLS, reverse proxy.  
Referensi: [Cloudflare Learning Center](https://www.cloudflare.com/learning/network-layer/).

Fakta yang baiknya lo tahu (biar nggak kaget pas ngobrol sama orang infra):

- Pangsa traffic global Cloudflare **2025**: **HTTP/2 ~50%**, HTTP/1.x ~29%, **HTTP/3 ~21%** — [Cloudflare Radar 2025 Year in Review](https://blog.cloudflare.com/radar-2025-year-in-review/). HTTP/3 (RFC 9114 di atas QUIC) bukan lagi "eksperimen".
- **OpenSSL 3.5 LTS** rilis 8 April **[2025]**, didukung sampai 2030 — [pengumuman](https://openssl-library.org/post/2025-04-08-openssl-35-final-release/).

### Hari 2 · Praktik (2 jam)

```bash
ip addr                              # alamat IP + interface
ip route                             # ke mana traffic dilempar
ss -lntp                             # port yang listen + proses pemiliknya
dig example.com +trace               # DNS dari root → TLD → authoritative
dig +short example.com               # cuma jawabannya
getent hosts redis                   # resolusi ala aplikasi (bukan ala dig!)
```

Ukur fase latency:

```bash
curl -sS -o /dev/null -w 'dns=%{time_namelookup} tcp=%{time_connect} tls=%{time_appconnect} ttfb=%{time_starttransfer} total=%{time_total} http=%{http_code} ver=%{http_version}\n' https://example.com
```

Coba juga ke service lokal lo:

```bash
curl -sS -o /dev/null -w 'dns=%{time_namelookup} tcp=%{time_connect} ttfb=%{time_starttransfer} total=%{time_total}\n' http://localhost:8080/health
```

**Latihan wajib:** jalanin perintah itu 5 kali, lalu jawab: kenapa `dns` ke `localhost` nyaris nol tapi ke domain publik bisa puluhan milidetik? Kenapa `ttfb` turun naik? Kenapa `total` jauh lebih besar dari `ttfb` di response gede?

### Hari 3 · Eksperimen (2 jam)

Reproduksi 4 kegagalan klasik, dan **catat persis error-nya**:

```bash
# 1. Port salah → connection refused
curl --connect-timeout 3 http://localhost:9999
# 2. Service mati tapi DNS/host ada
docker compose stop api && curl --connect-timeout 3 http://localhost:8080/health
# 3. DNS gagal: host yang nggak ada di jaringan Docker
docker compose exec api getent hosts nope-service
docker compose exec api curl -v --connect-timeout 3 http://nope-service:8080
# 4. Timeout: iptables DROP (bukan REJECT) di lab, atau firewall cloud
```

Bikin tabel begini di `labs/week-02-networking/commands.md`:

| Kondisi                 | Error yang muncul        | exit code curl | Cara deteksi           |
| ----------------------- | ------------------------ | -------------- | ---------------------- |
| DNS gagal               | `Could not resolve host` | 6              | `dig` / `getent hosts` |
| Port nggak listen       | `Connection refused`     | 7              | `ss -lntp`             |
| Firewall drop / routing | hang lalu `Timeout`      | 28             | `mtr`, `tcpdump`       |
| Aplikasi error          | HTTP 500                 | 0              | log aplikasi           |

**Jebakan besar:** exit code `curl` itu informasi, bukan gangguan. Biasakan cek `$?` — kebanyakan engineer cuma lihat outputnya.

### Hari 4 · Review (2 jam)

1. **Diagram request flow**: client → DNS → TCP → TLS → Nginx → Go API → Postgres/Redis. Gambar di `labs/week-02-networking/` (bisa pakai ASCII art, Mermaid, atau Excalidraw — yang penting ada).
2. **Decision tree troubleshooting** untuk: DNS failure, connection refused, timeout.

Bentuknya gini:

```
Request gagal
├─ curl exit 6  → DNS. Cek: dig, getent hosts, /etc/resolv.conf, resolver internal Docker
├─ curl exit 7  → Nggak ada yang listen. Cek: ss -lntp, prosesnya hidup?, port-nya bener?
├─ curl exit 28 → Timeout. Cek: firewall/SecurityGroup, routing, service overload
└─ exit 0 tapi 5xx → Aplikasi. Cek: log aplikasi, dependency health, trace
```

### Jebakan umum minggu 2

- **`time_starttransfer` (TTFB) itu bukan "waktu jaringan"** — udah termasuk waktu server mikir. `total - starttransfer` baru waktu kirim body.
- **Hasil `dig` belum tentu yang dipakai aplikasi.** Ada cache, `systemd-resolved`, stub resolver `127.0.0.53`. Makanya `getent hosts` lebih jujur.
- **Ping diblok ≠ service mati.** ICMP dan TCP itu jalan tol berbeda.
- **`openssl s_client` tanpa `-servername`** bisa dapet sertifikat yang salah, karena SNI.
- **DNS di Docker Compose itu internal.** `api` bisa resolve `redis`, tapi laptop lo nggak bisa. Ini sumber 90% kebingungan pemula.

### Acceptance criteria minggu 2

1. Bisa jelasin beda **DNS failure, TCP failure, HTTP error, timeout** — dan mereproduksi keempatnya.
2. Bisa nemuin port yang listen **dan** proses pemiliknya.
3. Bisa ukur dan jelasin 5 fase latency satu request.
4. Bisa buktiin versi TLS dan protokol (h2/h3) yang dipakai:

   ```bash
   openssl s_client -connect example.com:443 -servername example.com -tls1_3 -alpn h2 -brief
   ```

**Yang masuk GitHub:** `labs/week-02-networking/` — diagram arsitektur, hasil eksperimen (tabel error), runbook decision tree.

---

## Minggu 3 · Docker production fundamentals

**Target:** operasi container dengan lifecycle, resource, dan network yang terkontrol.

### Hari 1 · Teori (2 jam)

Image layer, container lifecycle, bridge network, volume, health check, cgroups, beda `CMD` vs `ENTRYPOINT`.  
Referensi: [Docker Get Started](https://docs.docker.com/get-started/) + [Compose file reference](https://docs.docker.com/reference/compose-file/services/).

**Perubahan yang wajib lo tahu:**

- **`docker-compose` (v1, Python) sudah dihapus.** Sekarang `docker compose`. Semua tutorial lama yang pakai strip `-` itu usang.
- **`depends_on` bentuk pendek itu jebakan.** Dia cuma nunggu container *start*, bukan nunggu *siap*. Bentuk panjang dengan `condition: service_healthy` yang benar.
- **`deploy.replicas` dan `update_config` itu fitur Swarm.** Kalau lo pakai Compose biasa, jangan kaget waktu nggak jalan. Yang tetap jalan di Compose non-Swarm: `deploy.resources.limits`.

### Hari 2 · Praktik (2 jam)

Bikin Compose stack: Go API + PostgreSQL + Redis. Ini versi yang udah bener:

```yaml
# docker/compose.yaml
services:
  api:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    user: "10001:10001"                 # jangan jalan sebagai root
    env_file: ../.env
    depends_on:
      postgres:
        condition: service_healthy      # BUKAN cuma "started"
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "/app/sre-api", "-healthcheck"]   # atau wget, asal ada di image
      interval: 10s
      timeout: 3s
      retries: 3
      start_period: 15s                 # kasih waktu buat warm-up
    deploy:
      resources:
        limits: { cpus: "0.50", memory: 256M, pids: 200 }

  postgres:
    image: postgres:18
    environment:
      POSTGRES_USER: ${PGUSER}
      POSTGRES_PASSWORD: ${PGPASSWORD}   # dari .env, JANGAN ditulis di sini
      POSTGRES_DB: ${PGDATABASE}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${PGUSER}"]
      interval: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      retries: 5

volumes:
  pgdata:
```

> Catatan: `postgres` dan `redis` **tidak punya `ports:`** — mereka cuma bisa diakses dari jaringan internal Compose. Ini kebiasaan yang bikin beda antara "lab" dan "calon insiden keamanan".

```bash
docker compose config                 # validasi sebelum jalan
docker compose up -d --build --wait   # --wait = tunggu sampai healthcheck OK
docker compose ps --format 'table {{.Name}}\t{{.Status}}\t{{.Health}}'
docker compose logs -f api
docker inspect <container>
docker stats --no-stream
docker network ls
```

### Hari 3 · Eksperimen (2 jam)

```bash
# 1. Batasi memori sampai mepet, lalu lihat apa yang terjadi
docker run --rm --memory=64m --cpus=.5 --pids-limit=100 <image>
# 2. Matikan dependency
docker compose stop redis
docker compose logs -f api            # apakah API panic, hang, atau gagal dengan terkendali?
docker compose ps                     # health status berubah?
# 3. Cari tahu kenapa container restart terus
docker inspect --format '{{json .State}}' <container> | jq
docker events --since 10m
docker compose start redis
```

**Hal yang paling penting dari minggu ini:** kalau Redis mati dan API lo malah *hang* (bukan balas 503 cepat), berarti lo baru nemuin bug nyata. Itu yang namanya **fail without controlled degradation**. Catat, terus perbaiki di minggu 6.

### Hari 4 · Review (2 jam)

1. Healthcheck yang **beneran nguji kesiapan** — bukan `test: ["CMD", "true"]`.
2. Dokumentasi perbandingan di `labs/week-03-docker/commands.md`:

| Mekanisme                 | Yang dijawab                             | Yang TIDAK dijawab                            |
| ------------------------- | ---------------------------------------- | --------------------------------------------- |
| `depends_on`              | "apakah container lain sudah distart?"   | "apakah dia sudah siap menerima traffic?"     |
| `healthcheck`             | "apakah proses ini sehat?"               | "apakah dependency-nya sehat?"                |
| Application retry/backoff | "gimana kalau dependency sebentar mati?" | — (ini tanggung jawab aplikasi, bukan Docker) |

1. **.env.example + .gitignore.** Password database nggak boleh ada di repo.

### Jebakan umum minggu 3

- Healthcheck pakai `curl` di image distroless → **selalu unhealthy**, karena nggak ada `curl`.
- Tanpa `USER` di Dockerfile → container jalan sebagai **root**.
- Tanpa limit memori → yang kena OOM itu **host**, bukan container.
- Image tag `:latest` + layer cache → **rollback jadi mustahil**. Biasakan tag SHA atau versi.
- Volume database nggak dipisah → `docker compose down -v` = data hilang. Pahami beda `down` dan `down -v`.

### Acceptance criteria minggu 3

1. Seluruh service hidup dari kondisi bersih dengan **satu perintah**.
2. API **gagal secara terkendali** saat dependency nggak tersedia (bukan hang/panic).
3. Setiap container punya limit resource masuk akal + healthcheck yang berfungsi.
4. Nggak ada password hardcoded; `.env` ada di `.gitignore`, `.env.example` ada di repo.

**Yang masuk GitHub:** `labs/week-03-docker/`, `docker/compose.yaml`, `docker/Dockerfile`, `.env.example`, troubleshooting guide.

---

## Minggu 4 · DNS, TLS, dan reverse proxy

**Target:** akses HTTP/HTTPS yang benar, dan bisa diagnosis masalah di depan aplikasi.

### Hari 1 · Teori (2 jam)

Reverse proxy, TLS certificate, upstream health, request buffering, keep-alive, proxy timeout, graceful shutdown.  
Referensi: [NGINX Beginner's Guide](https://nginx.org/en/docs/beginners_guide.html) + [ngx\_http\_proxy\_module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html).

**Yang berubah dan perlu lo tahu:**

| Hal                   | Fakta                                                                                                                                                |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| Default timeout nginx | `proxy_connect_timeout`, `proxy_read_timeout`, `proxy_send_timeout` semua **60 detik** — terlalu lama buat API modern                                |
| Keep-alive upstream   | Mulai **nginx 1.29.7** `keepalive` di blok `upstream` dan `proxy_http_version 1.1` **default ON**                                                    |
| HTTP/3                | Modul `ngx_http_v3_module` ada sejak 1.25.0 tapi masih **eksperimental**; butuh UDP/443 terbuka                                                      |
| Sertifikat pendek     | Let's Encrypt mulai terbitkan sertifikat **6 hari** (pertama 20 Feb **[2025]**, GA awal 2026) → renewal harus otomatis harian, cron bulanan = outage |
| Alternatif            | **Caddy** (HTTPS otomatis, paling enak buat lab), **Traefik** (cocok buat K8s/Docker), **Envoy** (kelas service mesh)                                |

### Hari 2 · Praktik (2 jam)

```nginx
# nginx/nginx.conf
upstream api {
    server api:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://api;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 5s;
        proxy_read_timeout    30s;
        proxy_send_timeout    30s;
        proxy_next_upstream   error timeout http_502;
    }

    access_log /var/log/nginx/access.log;
    error_log  /var/log/nginx/error.log warn;
}
```

Tambahin ke compose, **cuma nginx yang punya `ports:`**:

```yaml
  nginx:
    image: nginx:1.29-alpine
    ports:
      - "80:80"
    volumes:
      - ../nginx/nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      api: { condition: service_healthy }
```

```bash
docker compose up -d
curl -v http://localhost/health
docker compose exec nginx nginx -t     # SELALU ini sebelum reload
docker compose logs nginx
```

### Hari 3 · Eksperimen (2 jam) — Failure matrix

Ini bagian paling berharga minggu ini. Reproduksi 4 error, dan **buktikan dari log**:

| Kode                      | Arti sebenarnya                                                          | Cara bikin di lab                                       | Di mana cek                                               |
| ------------------------- | ------------------------------------------------------------------------ | ------------------------------------------------------- | --------------------------------------------------------- |
| **502 Bad Gateway**       | Nginx bisa nyambung tapi upstreamnya nolak/reset/header kegedean         | Matiin container `api`                                  | `error.log`: `connect() failed (111: Connection refused)` |
| **504 Gateway Timeout**   | `proxy_read_timeout` habis — upstreamnya mikir kelewat lama              | Bikin endpoint `/slow?ms=60000`                         | `error.log`: `upstream timed out`                         |
| **499**                   | **Bukan standar HTTP.** Kode internal nginx: client nutup koneksi duluan | Client timeout lebih kecil dari `proxy_read_timeout`    | `access.log` kolom status                                 |
| **TLS handshake failure** | Versi TLS/cipher/sertifikat nggak cocok                                  | `curl --tls-max 1.1` ke server yang cuma nerima TLS 1.3 | `error.log`: `SSL_do_handshake() failed`                  |

```bash
docker compose stop api && curl -sS -o /dev/null -w '%{http_code}\n' http://localhost/health
docker compose exec nginx tail -5 /var/log/nginx/error.log
docker compose start api
```

### Hari 4 · Review (2 jam)

1. **Dokumentasi request flow** (update diagram minggu 2, sekarang ada proxy-nya).
2. **Failure matrix** di atas, diisi dengan log asli hasil eksperimen lo.
3. **TLS lokal.** Buat lab:

   ```bash
   mkcert -install            # masukin CA lokal ke trust store
   mkcert app.local 127.0.0.1 ::1
   ```

   Atau kalau mau lebih serius: [step-ca](https://github.com/smallstep/certificates).
   > mkcert CA cuma dipercaya di mesin lo sendiri. **Jangan** dipakai buat domain publik.

### Jebakan umum minggu 4

- **502 dan 504 sering ketuker.** Intinya: 502 = upstream *nolak*, 504 = upstream *kelamaan*.
- Lupa `proxy_set_header Host $host` → backend salah pilih vhost.
- Dua proxy beruntun → header `X-Forwarded-For` dobel atau ketimpa.
- `proxy_read_timeout` lebih panjang dari timeout client → hasilnya **499**, bukan 504.
- Sertifikat 6 hari tanpa auto-renewal harian = outage berkala. Ini alasan kenapa ACME automation itu wajib dipahami.

### Acceptance criteria minggu 4

1. Client akses API **cuma lewat reverse proxy**.
2. Port Postgres dan Redis **tidak terekspos ke host** (buktikan dengan `ss -lntp` dan `nmap` dari luar).
3. Kegagalan upstream menghasilkan error yang **bisa ditelusuri dari log**.
4. Bisa jelasin dan reproduksi 502 / 504 / 499.

**Yang masuk GitHub:** `labs/week-04-reverse-proxy/` — konfigurasi nginx, diagram, failure matrix.

---

# Bulan 2: Deployment dan Operasi Production

Tujuan bulan ini: berhenti jadi orang yang takut deploy hari Jumat.

---

## Minggu 5 · CI/CD dan release management

**Target:** hilangin deployment manual yang rawan lupa langkah.

### Kenapa ini minggu pertama bulan 2

Karena deployment manual itu sumber insiden nomor satu yang **bisa dicegah dengan murah**. Bukan karena engineer-nya ceroboh, tapi karena otak manusia nggak didesain buat ngikutin 14 langkah berurutan dengan konsisten.

### Hari 1 · Teori (2 jam)

CI vs CD, immutable artifacts, versioning, environment separation, secrets, rollback.  
Referensi: [GitHub Actions docs](https://docs.github.com/en/actions).

Yang sering nggak diajarin tutorial:

- **Immutable artifact**: satu build = satu artefak yang nggak pernah diubah. Kalau lo build ulang di tiap environment, lo nggak pernah nguji barang yang sama dengan yang jalan di produksi.
- **Concurrency group**: dua push beruntun jangan bikin dua deploy nabrak.
- **`permissions:` eksplisit**: default repo lama itu `write`, dan itu bahaya. Tulis eksplisit.
- **OIDC, bukan secret statis**: GitHub Actions bisa "login" ke AWS/GCP pakai token OIDC yang umurnya pendek. Nggak ada lagi AWS key yang berumur 5 tahun di repo.
- **Attestation**: bukti kriptografis bahwa image ini dibangun dari commit ini, oleh workflow ini. Ini level "supply chain" — mulai jadi standar 2025+.

### Hari 2 · Praktik (2 jam)

```yaml
# .github/workflows/ci.yml
name: ci
on:
  pull_request:
  push:
    branches: [main]

permissions:            # WAJIB: tulis eksplisit, jangan andalkan default
  contents: read

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
          cache: true
      - run: go vet ./...
      - run: go test -race -covermode=atomic ./...
      - uses: golangci/golangci-lint-action@v9   # config v2 (golangci-lint migrate)

  integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:18
        env: { POSTGRES_PASSWORD: test, POSTGRES_DB: sre_lab }
        options: >-
          --health-cmd "pg_isready -U postgres" --health-interval 5s
          --health-retries 10
        ports: ["5432:5432"]
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping" --health-interval 5s --health-retries 10
        ports: ["6379:6379"]
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v7
        with: { go-version-file: go.mod, cache: true }
      - run: go test -tags=integration ./tests/integration/...

  build:
    needs: [test, integration]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      id-token: write           # buat OIDC / attestation
      attestations: write
    steps:
      - uses: actions/checkout@v5
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with: { registry: ghcr.io, username: ${{ github.actor }}, password: ${{ secrets.GITHUB_TOKEN }} }
      - id: meta
        run: echo "sha=$(git rev-parse --short=12 HEAD)" >> "$GITHUB_OUTPUT"
      - uses: docker/build-push-action@v7
        with:
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: ghcr.io/${{ github.repository }}/api:sha-${{ steps.meta.outputs.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

Lokal:

```bash
go test ./...
go vet ./...
docker build -t sre-lab-api:$(git rev-parse --short=12 HEAD) .
```

### Hari 3 · Eksperimen (2 jam)

1. **Sengaja gagalin test** → buktiin pipeline nolak release. Screenshot/log run merah itu bukti lo.
2. **Tag berbasis commit SHA** → `sha-abc123def456`. Bukan `latest`.
3. **Uji rollback:**

   ```bash
   # deploy versi lama dengan cara nunjuk digest image, bukan tag
   docker pull ghcr.io/owner/api@sha256:<digest-lama>
   ```

4. **Cek kebocoran secret:** jalanin `gitleaks detect` atau minimal grep manual. Pastikan nggak ada token di log.

### Hari 4 · Review (2 jam)

Tulis `docs/release-process.md` yang njawab 4 pertanyaan ini:

| Pertanyaan                       | Jawaban yang harus ada                                             |
| -------------------------------- | ------------------------------------------------------------------ |
| Gimana cara build?               | Perintah + di mana jalannya                                        |
| Gimana cara publish image?       | Registry + siapa yang punya akses                                  |
| Gimana deploy versi tertentu?    | Nunjuk **digest**, bukan tag mutable                               |
| Gimana balikin versi sebelumnya? | Perintah rollback + berapa lama + apa yang diverifikasi setelahnya |

### Jebakan umum minggu 5

- **`latest` itu musuh rollback.** Dia mutable; lo nggak pernah tahu persis apa yang jalan.
- **`pull_request_target` + checkout PR head** = celah RCE. Jangan.
- **Lupa `permissions:`** di repo lama → workflow dapet akses `write` gratis.
- **Attestation tanpa verifikasi** = teater keamanan. Sertakan juga cara verifikasinya (`gh attestation verify`).
- **Cache Go bentrok sama matrix build** → build sukses di satu versi, gagal di versi lain secara acak.

### Acceptance criteria minggu 5

1. Test jalan otomatis tiap PR, selesai **< 10 menit**.
2. Image punya versi yang bisa dilacak (commit SHA) dan immutable.
3. Test gagal → release ditolak (ada buktinya).
4. Nggak ada secret di log/artefak; `permissions:` ditulis eksplisit.

**Yang masuk GitHub:** `.github/workflows/`, integration tests, `docs/release-process.md`.

---

## Minggu 6 · Secrets, konfigurasi, dan deployment

**Target:** deployment konsisten, aman, dan gampang di-rollback.

### Hari 1 · Teori (2 jam)

Configuration management, environment variables, secret injection, readiness vs liveness, graceful shutdown, rolling deployment.

**Konsep yang paling sering bikin insiden:**

> **Liveness probe yang ngecek dependency = mesin pembuat outage.**

Kalau database down 30 detik, dan liveness lo ngecek database, Kubernetes akan **restart semua pod**. Sekarang lo punya: database down + semua aplikasi restart bersamaan + connection storm begitu DB balik. Ini namanya **cascading failure**.

Aturan praktis:

| Probe              | Pertanyaan                    | Kalau gagal                 | Boleh ngecek dependency?                     |
| ------------------ | ----------------------------- | --------------------------- | -------------------------------------------- |
| **startupProbe**   | "udah selesai booting belum?" | jangan restart, tunggu      | boleh (ini wajar)                            |
| **readinessProbe** | "boleh nerima traffic?"       | keluarin dari load balancer | **ya, ini tempatnya**                        |
| **livenessProbe**  | "prosesnya masih waras?"      | restart                     | **jangan** — cuma cek proses masih responsif |

### Hari 2 · Praktik (2 jam)

Graceful shutdown di Go — ini kodenya, dan tiap baris ada alasan:

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,   // cegah slowloris
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            slog.Error("server failed", "err", err)
            os.Exit(1)
        }
    }()

    slog.Info("server ready", "addr", srv.Addr)
    <-ctx.Done()                              // dapet SIGTERM
    slog.Info("shutting down gracefully")

    // timeout ini HARUS lebih kecil dari terminationGracePeriodSeconds (K8s)
    // atau dari waktu yang dikasih orchestrator lo
    c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()

    if err := srv.Shutdown(c); err != nil {
        slog.Error("forced shutdown", "err", err)
        _ = srv.Close()
    }
    db.Close()      // tutup pool koneksi dengan rapi
    redis.Close()
    slog.Info("shutdown complete")
}
```

Dua endpoint yang beda:

```go
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
    // liveness: cuma bilang "proses gue hidup"
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("ok"))
})

mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
    // readiness: cek dependency yang wajib
    c, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    if err := db.PingContext(c); err != nil {
        http.Error(w, "db not ready", http.StatusServiceUnavailable)
        return
    }
    if err := redis.Ping(c).Err(); err != nil {
        http.Error(w, "redis not ready", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
})
```

### Hari 3 · Eksperimen (2 jam)

```bash
# 1. Kirim request terus-terusan
hey -n 20000 -c 50 http://localhost/health &

# 2. Di sesi lain, kirim SIGTERM
docker compose stop api       # atau: kill -TERM <PID>

# 3. Lihat: berapa 5xx yang muncul?
# target: 0
```

Lalu:

- **Simulasi startup gagal**: kasih password DB salah → `readyz` harus 503, dan container nggak pernah dibilang healthy.
- **Dependency belum siap**: matiin Redis → `readyz` 503, `healthz` tetap 200. Ini beda yang penting.
- **Catat berapa lama** dari SIGTERM sampai proses benar-benar exit.

### Hari 4 · Review (2 jam)

1. **Deployment checklist** (di `docs/deployment.md`): pre-check, langkah deploy, post-check (health, metrics, log), kriteria "kapan harus rollback".
2. **Rollback procedure**: perintah spesifik + estimasi waktu + apa yang diverifikasi.
3. **Catat keterbatasan lab dengan jujur.** Docker Compose **tidak punya** perintah `rollout`; `update_config` itu fitur Swarm. Jadi nol downtime di Compose itu terbatas — tulis ini, jangan pura-pura. Ini justru nilai plus di mata reviewer: lo tahu batas alat lo.

### Jebakan umum minggu 6

- Shutdown timeout **lebih besar** dari grace period orchestrator → proses dipaksa mati di tengah-tengah, request putus.
- Secret di environment variable gampang bocor ke `docker inspect`, log error, dan crash report.
- SIGTERM nggak nyampe ke proses lo kalau PID 1 di container itu shell, bukan binary lo. Pakai exec form di Dockerfile, atau `init: true`.
- Readiness tanpa **connection drain** → 502 saat rolling update meski aplikasinya sehat.

### Acceptance criteria minggu 6

1. Aplikasi berhenti **graceful** saat SIGTERM, dan request in-flight selesai.
2. `readyz` **tidak** bilang siap saat kondisi wajib belum terpenuhi.
3. Deployment gagal nggak menimpa versi sehat tanpa rencana recovery.
4. Nggak ada secret plaintext di Git (SOPS/age atau secret manager).

**Yang masuk GitHub:** `labs/week-06-deployment/` — konfigurasi, kode shutdown, deployment checklist, laporan eksperimen.

---

## Minggu 7 · Backup dan disaster recovery

**Target:** data bisa dipulihkan, bukan cuma "pernah jalanin cron".

### Hari 1 · Teori (2 jam)

Logical vs physical backup, retention, RPO, RTO, point-in-time recovery (PITR), backup off-site.  
Referensi: [PostgreSQL Backup and Restore](https://www.postgresql.org/docs/current/backup.html) dan [Continuous Archiving & PITR](https://www.postgresql.org/docs/current/continuous-archiving.html).

**Definisi yang harus lo hapal:**

| Istilah                            | Arti                                                         | Contoh target lab                           |
| ---------------------------------- | ------------------------------------------------------------ | ------------------------------------------- |
| **RPO** (Recovery Point Objective) | Seberapa banyak data yang **boleh hilang**                   | "maksimal 5 menit transaksi"                |
| **RTO** (Recovery Time Objective)  | Berapa lama **boleh** nggak bisa dipakai                     | "maksimal 30 menit"                         |
| **PITR**                           | Balik ke titik waktu tertentu, bukan cuma ke backup terakhir | "balikin ke 10:47 sebelum ada DELETE salah" |

Perbedaan paling penting: `pg_dump` itu **logical backup**. Bisa restore ke titik backup, tapi **nggak bisa PITR**. PITR butuh **base backup + WAL archiving**.

Tool yang dipakai di dunia nyata:

| Tool                                    | Versi (2025)          | Cocok buat                                                     |
| --------------------------------------- | --------------------- | -------------------------------------------------------------- |
| **pgBackRest**                          | v2.57.0 (18 Okt 2025) | Standar de-facto: full/incr/diff, kompresi, enkripsi, S3, PITR |
| **Barman**                              | 3.14.1 (Jun 2025)     | Tradisional, kuat di enterprise                                |
| **WAL-G**                               | 3.0.5 (Feb 2025)      | Ringan, cloud-native                                           |
| **CloudNativePG + plugin-barman-cloud** | CNPG ≥ 1.26           | Kalau Postgres jalan di Kubernetes                             |

### Hari 2 · Praktik (2 jam)

Mulai dari yang sederhana (roadmap lo):

```bash
docker compose exec -T postgres pg_dump -U app -d sre_lab -Fc > backup.dump
pg_restore --list backup.dump          # lihat isinya tanpa restore
```

Lalu tingkatkan ke PITR dengan pgBackRest + MinIO (S3-compatible lokal):

```bash
# stanza = "set konfigurasi backup" di pgBackRest
pgbackrest --stanza=app --type=full backup
pgbackrest --stanza=app --type=incr backup
pgbackrest --stanza=app info           # lihat daftar backup + ukuran + waktu
pgbackrest --stanza=app verify         # cek integritas repo
```

Yang wajib ada di konfigurasi:

```ini
[global]
repo1-type=s3
repo1-s3-bucket=sre-lab-backup
repo1-s3-endpoint=minio:9000
repo1-cipher-type=aes-256-cbc         # backup terenkripsi
repo1-cipher-pass=<dari secret, bukan plaintext>
repo1-retention-full=7
repo1-retention-diff=14
compress-type=zst
```

### Hari 3 · Eksperimen (2 jam) — Restore drill

Ini inti minggu ini. **Backup yang belum pernah di-restore itu bukan backup, cuma harapan.**

```bash
# 1. Siapkan database baru yang KOSONG (instance terpisah!)
docker compose -f docker/compose.restore.yaml up -d postgres-restore

# 2. Restore
pgbackrest --stanza=app --type=time \
  --target="2026-10-12 10:00:00+00" --target-action=promote restore

# 3. Verifikasi — jangan cuma lihat "restore selesai"
psql -h localhost -U app -d sre_lab -c "SELECT count(*) FROM orders;"
psql -c "SELECT md5(string_agg(id::text, ',' ORDER BY id)) FROM orders;"   # checksum sederhana
psql -c "\d+ orders"        # constraint, index, foreign key ikut kebawa?
```

Ukur dan catat:

| Metrik            | Cara ukur                                    | Target lab                             |
| ----------------- | -------------------------------------------- | -------------------------------------- |
| Waktu backup full | `time pgbackrest --type=full backup`         | catat apa adanya                       |
| Ukuran backup     | `pgbackrest info`                            | catat + bandingkan dengan ukuran DB    |
| Waktu restore     | `time pgbackrest ... restore`                | bandingkan dengan RTO                  |
| Waktu verifikasi  | waktu yang lo butuh buat yakin datanya benar | sering kelewat, padahal ini bagian RTO |

> **Simulasikan kehilangan data cuma di database yang disposable.** Jangan pernah latihan `DROP TABLE` di database pribadi atau punya orang lain. Ini aturan yang nggak bisa ditawar.

### Hari 4 · Review (2 jam)

Tulis `docs/disaster-recovery.md` berisi:

1. **Target RPO dan RTO** buat lab, beserta alasan kenapa angka itu.
2. **Recovery runbook**: langkah berurutan, siapa yang jalanin, apa yang diverifikasi.
3. **Retention & lokasi**: backup disimpan di mana, berapa lama, terenkripsi apa nggak.
4. **Yang belum ditangani** — ditulis eksplisit. Contoh: "backup belum diuji restore otomatis berkala", "belum ada off-site ke region lain". Ini yang bikin dokumen lo kelihatan ditulis oleh orang yang pernah kena insiden.

### Jebakan umum minggu 7

- `pg_dump` tanpa pernah di-restore sering dianggap "backup aman" padahal nggak bisa PITR.
- Lupa `archive_mode = on` → WAL nggak kearsip → PITR gagal total.
- Retensi WAL lebih pendek dari retensi full backup → backup jadi nggak berguna.
- Backup dan WAL di bucket yang sama tanpa object lock → kena ransomware, hilang dua-duanya.
- **Verifikasi cuma "restore exit code 0".** Itu bukan bukti datanya benar.

### Acceptance criteria minggu 7

1. Restore benar-benar berhasil ke instance/database terpisah.
2. Data hasil restore lolos verifikasi (jumlah row + constraint + checksum).
3. Durasi recovery diukur dan dibandingkan dengan target RTO.
4. Backup terenkripsi dan lokasinya terpisah dari volume database.

**Yang masuk GitHub:** `labs/week-07-backup-recovery/` — script backup/restore, runbook, hasil uji recovery.

---

## Minggu 8 · Production readiness dan reliability baseline

**Target:** gabungin deployment, jaringan, dan recovery jadi sistem yang bisa dioperasikan.

### Hari 1 · Teori (2 jam)

Production Readiness Review (PRR), dependency mapping, failure mode, operational checklist, single point of failure.

Referensi: [Google SRE — Evolving SRE Engagement Model (PRR)](https://sre.google/sre-book/evolving-sre-engagement-model/) dan [Launch Coordination Checklist](https://sre.google/sre-book/launch-checklist/).

**PRR itu bukan formulir.** Dia jawab: "kalau sistem ini rusak jam 3 pagi, apakah orang yang lagi jaga bisa ngerti dan benerin tanpa nanya-nanya?"

### Hari 2 · Praktik (2 jam)

```bash
docker compose config            # validasi
docker compose up -d --build --wait
docker compose ps
curl -fsS http://localhost/health
curl -fsS http://localhost/readyz
```

Syaratnya: **satu perintah** dari repo kosong sampai sistem jalan. Kalau lo harus ngejalanin 5 perintah tambahan dari ingatan, berarti reproduksinya belum selesai — dan itu bug.

### Hari 3 · Eksperimen (2 jam)

```bash
sudo reboot                      # atau docker compose down && docker compose up -d
```

Lalu verifikasi, dan ini yang sering bikin malu:

- [ ] Aplikasi balik online **tanpa campur tangan manual**?
- [ ] Data Postgres masih ada (volume persist)?
- [ ] Env/secret masih terbaca?
- [ ] Healthcheck jalan lagi?
- [ ] Log masih nyambung?

Ulangi recovery dari kondisi service mati, lalu tulis: berapa lama dari reboot sampai `/readyz` balas 200.

### Hari 4 · Review (2 jam)

Bikin `docs/production-readiness.md` dengan 5 kategori:

**1. Security**

- [ ] Nggak ada secret di repo/image (`gitleaks`, `trivy`)
- [ ] Container non-root, image di-scan
- [ ] Port internal nggak terekspos
- [ ] TLS aktif, sertifikat punya auto-renewal

**2. Availability**

- [ ] Minimal 2 instance (atau catat eksplisit kalau cuma 1)
- [ ] Healthcheck + readiness terpisah
- [ ] Graceful shutdown + rollback teruji

**3. Observability**

- [ ] Metrics (RED) — *ini kepakai mulai minggu 9*
- [ ] Log terstruktur dan ada correlation ID
- [ ] Alert buat "service down"
- [ ] Dashboard yang njawab "lagi sehat nggak?"

**4. Recovery**

- [ ] Backup otomatis + restore pernah diuji
- [ ] RPO/RTO tertulis
- [ ] Runbook buat 3 skenario kegagalan terbesar

**5. Deployment**

- [ ] Pipeline otomatis, artifact immutable
- [ ] Rollback teruji
- [ ] Nggak ada langkah manual yang wajib

Lengkapi dengan **dependency map + FMEA sederhana**:

| Komponen   | Gagalnya gimana | Dampak                     | Gimana deteksinya         | Mitigasi                           | SPOF?                  |
| ---------- | --------------- | -------------------------- | ------------------------- | ---------------------------------- | ---------------------- |
| PostgreSQL | mati / lambat   | semua endpoint tulis gagal | readiness + latency query | connection pool + timeout + backup | **ya**                 |
| Redis      | mati            | cache miss → DB kebanjiran | metrics hit ratio         | fallback: anggap cache kosong      | ya (tapi ada fallback) |
| Nginx      | mati            | semua akses gagal          | healthcheck external      | —                                  | **ya**                 |

### Jebakan umum minggu 8

- PRR jadi dokumen sekali pakai yang langsung basi.
- FMEA tanpa angka "gimana deteksinya" → nggak bisa ditindaklanjuti.
- Dependency pihak ketiga (auth provider, payment gateway) nggak masuk peta — padahal itu yang paling sering bikin outage.
- Nggak ada yang ditulis sebagai pemilik on-call.

### Acceptance criteria minggu 8

1. Sistem bisa dibangun dan dijalanin ulang **hanya dari repository**.
2. Semua langkah pemulihan **terdokumentasi DAN teruji**.
3. Risiko yang belum ditangani **ditulis eksplisit**.

**Yang masuk GitHub:** `docs/production-readiness.md`, diagram arsitektur v1, checklist audit.

---

# Bulan 3: Observability dan SRE Fundamentals

Tujuan bulan ini: berhenti nunggu user ngeluh, dan mulai ngomong pakai angka.

---

## Minggu 9 · Metrics dengan Prometheus

**Target:** tahu kondisi service dari metrics, bukan dari chat user.

### Hari 1 · Teori (2 jam)

Counter, gauge, histogram, scrape interval, labels, cardinality, PromQL.  
Referensi: [Prometheus Getting Started](https://prometheus.io/docs/tutorials/getting_started/).

**Versi yang relevan:**

- **Prometheus 3.0** rilis 14 Nov 2024 — rilis mayor pertama setelah 7 tahun: UI baru, Remote Write 2.0, UTF-8 di nama metric/label, OTLP receiver. ([blog](https://prometheus.io/blog/2024/11/14/prometheus-3-0/))
- Rilis 2025: **3.2.0** (Feb), **3.3.0** (Apr), **3.4.0** (Mei), **3.5.0** (Jul), **3.8.0** (Des). Hari ini 3.15.x; kalau mau stabil panjang, ikutin jalur LTS **3.13.x**.
- **Native histograms stabil sejak 3.8.0**, tapi **nggak aktif default** — harus `scrape_native_histograms: true`.

**4 tipe metric, dan analoginya:**

| Tipe          | Analogi                                                   | Contoh                       | Operasi yang benar                  |
| ------------- | --------------------------------------------------------- | ---------------------------- | ----------------------------------- |
| **Counter**   | Odometer mobil — cuma naik                                | total request, total error   | `rate()`, `increase()`              |
| **Gauge**     | Speedometer — naik turun                                  | koneksi DB aktif, memori     | langsung dipakai, `avg_over_time()` |
| **Histogram** | Ember ukuran: berapa request yang <10ms, <50ms, <100ms... | latency request              | `histogram_quantile()`              |
| **Summary**   | Kuantil dihitung di aplikasi                              | jarang dipakai di Prometheus | hati-hati, nggak bisa diagregasi    |

### Hari 2 · Praktik (2 jam)

Instrumentasi Go:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "http_server_requests_total",
            Help: "Total HTTP requests"},
        []string{"method", "route", "status"},   // route, BUKAN path mentah!
    )
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_server_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
        []string{"method", "route"},
    )
    inFlight = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "http_server_requests_in_flight",
        Help: "Requests currently being served",
    })
)

func init() {
    prometheus.MustRegister(httpRequests, httpDuration, inFlight)
}

// middleware
func instrument(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        inFlight.Inc()
        defer inFlight.Dec()
        start := time.Now()
        rw := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(rw, r)

        route := r.Pattern                      // "/orders/{id}" — kardinalitas rendah
        if route == "" {
            route = "unmatched"                 // JANGAN pakai r.URL.Path mentah
        }
        httpDuration.WithLabelValues(r.Method, route, ).
            Observe(time.Since(start).Seconds())
        httpRequests.WithLabelValues(r.Method, route, strconv.Itoa(rw.status)).Inc()
    })
}

mux.Handle("GET /metrics", promhttp.Handler())
```

Konfigurasi scrape:

```yaml
# observability/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: api
    scrape_native_histograms: true     # WAJIB kalau mau native histogram
    static_configs:
      - targets: ["api:8080"]

  - job_name: postgres
    static_configs:
      - targets: ["postgres-exporter:9187"]
```

### Hari 3 · Eksperimen (2 jam)

```bash
curl -fsS http://localhost:8080/metrics | grep http_server
docker compose logs prometheus
```

Lakukan:

1. Kirim 100 request normal.
2. Kirim 20 request yang sengaja error (misal `POST /orders` dengan body rusak).
3. Cek di Prometheus UI: `http_server_requests_total` naik sesuai jumlah? `status="500"` naik 20?
4. Bikin endpoint lambat, cek histogram bucket mana yang naik.

```bash
# kirim error dan hitung
for i in $(seq 1 20); do curl -sS -o /dev/null -X POST -d 'bad' http://localhost/orders; done
curl -sS http://localhost:8080/metrics | grep 'http_server_requests_total{.*status="500"'
```

### Hari 4 · Review (2 jam) — Query wajib

Simpan di `labs/week-09-prometheus/*.promql`:

```promql
# Request rate (requests/detik) per route
sum by (route) (rate(http_server_requests_total[5m]))

# Error ratio (0..1)
sum by (route) (
  rate(http_server_requests_total{status=~"5.."}[5m])
) / sum by (route) (rate(http_server_requests_total[5m]))

# p95 latency — agregasi bucket DULU, baru quantile
histogram_quantile(0.95,
  sum by (le, route) (rate(http_server_request_duration_seconds_bucket[5m]))
)

# Request in-flight (saturasi)
http_server_requests_in_flight

# Cek kardinalitas: 10 metric paling "gemuk"
topk(10, count by (__name__)({__name__=~".+"}))
```

### Jebakan umum minggu 9

- **`histogram_quantile` itu aproksimasi**, diasumsikan distribusi seragam dalam bucket. Bucket terlalu lebar = hasil ngaco. Quantile yang jatuh di bucket `+Inf` nggak ada artinya.
- **Jangan rata-ratakan hasil quantile antar instance.** Yang benar: jumlahkan `_bucket` dulu, baru hitung quantile.
- **Kardinalitas adalah pembunuh nomor satu.** Label `user_id`, `path` mentah, `request_id` → Prometheus meledak. Pakai `http_route` (pola, bukan nilai).
- **`rate()` butuh range selector minimal 4× scrape interval.** Interval 15s tapi pakai `[30s]` = hasilnya rewel.
- Native histogram nggak aktif sendiri — harus disetel di scrape config.

### Acceptance criteria minggu 9

1. Prometheus berhasil scrape API (target UP, ada di `/targets`).
2. Metrics berubah sesuai traffic dan error (bukti sebelum/sesudah).
3. Bisa jalanin query request rate, error ratio, dan p95 — dan hasilnya masuk akal.
4. Nggak ada label berkardinalitas tinggi.

**Yang masuk GitHub:** `labs/week-09-prometheus/` — instrumentasi Go, `prometheus.yml`, file `.promql`.

---

## Minggu 10 · Grafana dan alerting

**Target:** ubah metrics jadi informasi operasional.

### Hari 1 · Teori (2 jam)

Dashboard design, RED metrics, USE metrics, alert threshold, alert fatigue, symptom-based alerting.

**RED vs USE** — dua kacamata yang beda:

|              | **RED** (untuk service/request) | **USE** (untuk resource/infra)          |
| ------------ | ------------------------------- | --------------------------------------- |
| **R**ate     | request per detik               | **U**tilization: % waktu resource sibuk |
| **E**rror    | jumlah/jenis error              | **S**aturation: seberapa numpuk antrean |
| **D**uration | distribusi latency              | **E**rror: error di level resource      |

Pakai RED buat API, USE buat CPU/memori/disk/koneksi DB. Jangan dicampur jadi satu panel yang nggak jelas.

**Prinsip alerting yang bikin hidup on-call lo waras:**

1. **Alert berbasis gejala, bukan penyebab.** "User ngalamin error 5%" lebih baik dari "CPU 80%" kalau CPU 80% nggak bikin user rugi.
2. **Setiap alert harus bisa ditindaklanjuti.** Kalau nggak ada yang bisa lo lakuin, itu bukan alert — itu grafik.
3. **Tiap alert wajib punya runbook.** Kalau nggak ada runbook-nya, berarti lo nyiapin alarm tanpa petunjuk jalan keluar.
4. **Severity harus beda perlakuan**: page (bangunin orang) vs ticket (dilihat besok).

**Versi yang relevan:** Grafana 12.0 rilis 5 Mei **[2025]**; sekarang 13.x. **Promtail EOL 2 Maret 2026** — log pipeline pakai **Grafana Alloy**.

### Hari 2 · Praktik (2 jam)

Dashboard yang harus ada (RED + USE):

| Panel             | Query                                                   | Tipe                          |
| ----------------- | ------------------------------------------------------- | ----------------------------- |
| Request rate      | `sum by (route) (rate(http_server_requests_total[5m]))` | Time series                   |
| Error rate (%)    | error ratio × 100                                       | Time series + threshold merah |
| p95 / p99 latency | `histogram_quantile(0.95, ...)`                         | Time series                   |
| Request in-flight | `http_server_requests_in_flight`                        | Gauge                         |
| CPU container     | `rate(container_cpu_usage_seconds_total[5m])`           | Time series                   |
| Memory            | `container_memory_usage_bytes` vs limit                 | Time series                   |
| DB connections    | `pg_stat_activity_count`                                | Time series + threshold       |

**Provisioning as code** — ini yang bikin dashboard lo bisa di-review dan di-restore:

```yaml
# observability/grafana/provisioning/dashboards/default.yaml
apiVersion: 1
providers:
  - name: sre-core
    orgId: 1
    folder: SRE
    type: file
    disableDeletion: true
    updateIntervalSeconds: 30
    allowUiUpdates: false      # KUNCI: sumber kebenaran = Git, bukan klik-klik UI
    options:
      path: /var/lib/grafana/dashboards
```

Alert rules (Grafana-managed):

| Alert               | Kondisi                        | Severity | Runbook                       |
| ------------------- | ------------------------------ | -------- | ----------------------------- |
| `APIDown`           | `up{job="api"} == 0` selama 1m | **page** | `runbooks/api-down.md`        |
| `HighErrorRate`     | error ratio > 1% selama 5m     | **page** | `runbooks/high-error-rate.md` |
| `HighLatency`       | p95 > 300ms selama 10m         | ticket   | `runbooks/high-latency.md`    |
| `DBConnectionsHigh` | koneksi > 80% max selama 5m    | ticket   | `runbooks/db-saturation.md`   |

### Hari 3 · Eksperimen (2 jam)

```bash
docker compose stop api           # (1) matiin API
# catat: berapa detik dari mati sampai alert "firing"?

docker compose start api
# (2) naikin error rate: kirim request rusak terus-terusan
# (3) bikin latency tinggi: endpoint /slow?ms=2000, kirim beban
```

Ukur 3 hal untuk tiap alert, dan tulis hasilnya:

| Alert         | Berapa lama sampai firing? | Pesannya actionable? | False positive? |
| ------------- | -------------------------- | -------------------- | --------------- |
| APIDown       | ?                          | ya/tidak             | ?               |
| HighErrorRate | ?                          | ya/tidak             | ?               |

Kalau alert terlalu lama atau pesannya nggak jelas, **perbaiki aturannya sekarang** — bukan pas insiden beneran.

### Hari 4 · Review (2 jam)

Bikin **alert runbook** di `labs/week-10-grafana/runbooks/`. Format wajib:

```markdown
## Gejala
Apa yang terlihat user / apa yang muncul di alert.

## Kemungkinan penyebab (urut dari yang paling sering)
1. ...
2. ...

## Query yang dicek dulu
- ...

## Mitigasi cepat
1. ...

## Kapan eskalasi
Kalau ... dalam ... menit, escalate ke ...
```

### Jebakan umum minggu 10

- `allowUiUpdates: true` → dashboard berubah di UI, hilang pas redeploy. **Drift.**
- Alert tanpa `for:` → flapping (nyala-mati-nyala-mati), bikin on-call kebal.
- Satu contact point buat semua severity → yang penting tenggelam.
- Nggak ada **mute timings** buat maintenance window → alert noise pas maintenance.
- Alert "CPU > 80%" tanpa konteks dampak = alert yang nggak pernah ditindaklanjuti.

### Acceptance criteria minggu 10

1. Dashboard nunjukin traffic dan kondisi service **real time**, dan bisa di-reprovision dari nol.
2. Alert service-down kepicu dalam target waktu (misal **< 2 menit**).
3. Setiap alert penting punya instruksi tindakan yang jelas.
4. Pipeline log jalan lewat **Alloy**, bukan Promtail.

**Yang masuk GitHub:** `labs/week-10-grafana/` — dashboard JSON, alert rules, runbook.

---

## Minggu 11 · Logs dan distributed tracing

**Target:** nyambungin gejala di metrics ke penyebab di request tertentu.

### Konsep kunci: tiga pertanyaan, tiga alat

| Alat        | Njawab                          | Analogi         |
| ----------- | ------------------------------- | --------------- |
| **Metrics** | "Ada yang aneh nggak?"          | Alarm kebakaran |
| **Logs**    | "Apa persisnya yang terjadi?"   | CCTV            |
| **Traces**  | "Di titik mana waktunya habis?" | Peta rute kurir |

Urutannya: metrics bilang **ada masalah** → logs bilang **request mana** → trace bilang **di span mana**.

### Hari 1 · Teori (2 jam)

Structured logging, correlation ID, trace/span, context propagation, sampling.  
Referensi: [OpenTelemetry Documentation](https://opentelemetry.io/docs/).

**Status OTel yang perlu lo tahu:**

- **OTLP sudah stable** untuk traces, metrics, dan logs.
- **Semantic conventions**: HTTP spans dan DB client spans sudah **Stable**. Pakai env `OTEL_SEMCONV_STABILITY_OPT_IN=http,database` kalau mau metric nama baru yang stabil.
- **Propagasi konteks** pakai standar [W3C Trace Context](https://www.w3.org/TR/trace-context/) (`traceparent`) — ini yang bikin trace nyambung lintas service, bahkan lintas vendor.

### Hari 2 · Praktik (2 jam)

**Structured logging** — pakai `log/slog` bawaan Go (sejak Go 1.21), jangan tambah library kalau nggak perlu:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// di handler
span := trace.SpanFromContext(r.Context())
slog.InfoContext(r.Context(), "request_done",
    "method", r.Method,
    "route",  r.Pattern,
    "status", rw.status,
    "duration_ms", time.Since(start).Milliseconds(),
    "trace_id", span.SpanContext().TraceID().String(),
    "request_id", r.Header.Get("X-Request-Id"),
)
```

**Tracing dengan OTel:**

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

handler := otelhttp.NewHandler(mux, "sre-api")
```

**Yang paling sering bikin trace putus:** lupa nerusin `context.Context`.

```go
// SALAH — trace putus
func createOrder(ctx context.Context, o Order) error {
    return db.QueryRow("INSERT ...")            // context.Background() implicit? tergantung driver
}

// BENAR
func createOrder(ctx context.Context, o Order) error {
    return db.QueryRowContext(ctx, "INSERT ...")
}
```

**Pipeline log pakai Alloy** (bukan Promtail):

```river
// observability/otel/config.alloy
loki.source.docker "local" {
  host       = "unix:///var/run/docker.sock"
  targets    = [{__meta_docker_container_name = "sre-api", job = "api"}]
  forward_to = [loki.process.json.receiver]
}

loki.process "json" {
  stage.json {
    expressions = {level = "level", msg = "msg", trace_id = "trace_id", route = "route"}
  }
  stage.labels { values = {level = ""} }
  forward_to = [loki.write.default.receiver]
}

loki.write "default" {
  endpoint { url = "http://loki:3100/loki/api/v1/push" }
}

otelcol.receiver.otlp "default" {
  grpc {}
  output { traces = [otelcol.processor.tail_sampling.default.input] }
}

otelcol.processor.tail_sampling "default" {
  policy { latency = "500ms" }              // simpan trace lambat
  policy { status_code { status_codes = ["ERROR"] } }   // SELALU simpan trace error
  policy { probabilistic { sampling_percentage = 1 } }  // baseline 1%
  output { traces = [otelcol.exporter.otlp.tempo.input] }
}

otelcol.exporter.otlp "tempo" {
  client { endpoint = "tempo:4317" }
}
```

### Hari 3 · Eksperimen (2 jam)

1. Bikin endpoint yang lambat karena query database (misal `SELECT pg_sleep(1)` atau query tanpa index).
2. Kirim request, ambil `trace_id` dari response header atau log.
3. Cari log berdasarkan `trace_id` di Loki:

   ```logql
   {job="api"} | json | trace_id="<trace-id>"
   ```

4. Buka trace di Tempo, dan jawab: **dari total 1,2 detik, berapa yang habis di span database?**

Ini momen "oh, ternyata" yang paling sering kejadian minggu ini: lo kira aplikasinya lelet, ternyata 90% waktunya numpuk nunggu database.

### Hari 4 · Review (2 jam)

Tulis `labs/week-11-observability/` berisi panduan: **cara nTelusuri satu request dari dashboard ke log ke trace**, langkah demi langkah, dengan screenshot.

### Jebakan umum minggu 11

- **Trace putus** karena `context` nggak diteruskan ke fungsi turunan.
- **Correlation ID ada tapi nggak masuk log.** Harus di-inject tiap kali nulis log.
- **Tail sampling tanpa policy 100% untuk error** → trace yang paling penting justru kebuang.
- Ngira semua semconv sudah stabil — sebagian masih *Mixed*. Cek dulu.
- Log JSON tapi level-nya nggak konsisten (`level` vs `severity`) → query LogQL berantakan.

### Acceptance criteria minggu 11

1. Log terstruktur, bisa difilter berdasarkan request/trace ID.
2. Trace nunjukin durasi span aplikasi **dan** dependency.
3. Request lambat bisa ditelusuri tanpa nebak.
4. `traceparent` diteruskan dan trace nggak putus.

**Yang masuk GitHub:** `labs/week-11-observability/` — konfigurasi telemetry, contoh log tersanitasi, troubleshooting guide.

---

## Minggu 12 · SLI, SLO, error budget, dan incident response

**Target:** mulai mikir kayak SRE, bukan cuma operator monitoring.

### Hari 1 · Teori (2 jam)

| Istilah          | Definisi gampangnya                                    | Contoh                                                |
| ---------------- | ------------------------------------------------------ | ----------------------------------------------------- |
| **SLI**          | Angka yang diukur dari sisi user                       | "persentase request valid yang balas non-5xx"         |
| **SLO**          | Target dari angka itu                                  | "99,5% selama 30 hari rolling"                        |
| **SLA**          | Janji ke user, biasanya ada konsekuensi (denda/credit) | "99,9% per bulan, kalau melanggar dapat credit"       |
| **Error budget** | Sisa "jatah gagal" sebelum nabrak SLO                  | 99,5% dari 1 juta request = 5.000 request boleh gagal |

Baca: [Google SRE Workbook — Implementing SLOs](https://sre.google/workbook/implementing-slos/).

**Kenapa error budget itu ide yang brilian?** Karena dia ngubah perdebatan "boleh nggak kita rilis fitur baru?" dari perdebatan opini jadi **perdebatan angka**.

- Budget masih banyak → gas, rilis agresif.
- Budget habis → freeze rilis, kerjaan dialihin ke reliabilitas.

### Hari 2 · Praktik (2 jam)

Tulis `docs/slo.md`. Contoh buat lab:

```markdown
## SLI 1 — Availability
Definisi   : proporsi request valid yang tidak menghasilkan 5xx
Sumber data: http_server_requests_total (diukur di reverse proxy / load balancer)
Window     : 30 hari rolling
Eksklusi   : request dengan path /health* dan /metrics; request dari load generator internal
Catatan    : diukur dari sisi klien (proxy), bukan dari internal aplikasi

## SLO 1 — Availability
Target: 99,5% selama 30 hari rolling
Error budget: 0,5% × total request ≈ 5.000 request gagal per 1 juta
```

```markdown
## SLI 2 — Latency
Definisi   : proporsi request valid yang selesai < 300 ms
Kondisi    : beban ≤ 50 VU, diukur pakai k6
Window     : 30 hari rolling

## SLO 2 — Latency
Target: 95% request valid < 300 ms
```

**Penting:** tulis dengan jujur bahwa ini **target latihan**, bukan klaim performa.

### Hari 3 · Eksperimen (2 jam) — Alerting on SLO

Di sinilah teori ketemu praktik. Pakai **multiwindow, multi-burn-rate** dari [Google SRE Workbook — Alerting on SLOs](https://sre.google/workbook/alerting-on-slos/):

| Jendela panjang | Jendela pendek | Burn rate | Error budget habis | Aksi     |
| --------------- | -------------- | --------- | ------------------ | -------- |
| 1 jam           | 5 menit        | **14,4**  | 2% dalam 1 jam     | **Page** |
| 6 jam           | 30 menit       | **6**     | 5% dalam 6 jam     | **Page** |
| 3 hari          | 6 jam          | **1**     | 10% dalam 3 hari   | Ticket   |

Contoh rule (untuk SLO 99,9% → budget 0,1%):

```yaml
groups:
- name: slo-recording
  interval: 30s
  rules:
  - record: job:slo_errors_per_request:ratio_rate5m
    expr: |
      sum(rate(http_server_requests_total{status=~"5.."}[5m]))
      / sum(rate(http_server_requests_total[5m]))
  - record: job:slo_errors_per_request:ratio_rate1h
    expr: |
      sum(rate(http_server_requests_total{status=~"5.."}[1h]))
      / sum(rate(http_server_requests_total[1h]))

- name: slo-alerting
  rules:
  - alert: HighErrorBudgetBurn
    expr: |
      (job:slo_errors_per_request:ratio_rate1h > (14.4 * 0.001)
        and job:slo_errors_per_request:ratio_rate5m > (14.4 * 0.001))
      or
      (job:slo_errors_per_request:ratio_rate6h > (6 * 0.001)
        and job:slo_errors_per_request:ratio_rate30m > (6 * 0.001))
    for: 2m
    labels: { severity: page }
    annotations:
      summary: "Burn rate tinggi — budget 30 hari bisa habis dalam hitungan jam"
      runbook_url: "https://github.com/<user>/production-reliability-lab/blob/main/runbooks/high-error-rate.md"
```

Uji dengan:

```bash
promtool check rules rules/slo.yml          # sintaks
promtool test rules rules/slo_test.yml      # unit test aturan — INI yang sering kelewat
```

**Simulasi insiden** (2 jam):

1. Mulai dari traffic normal (k6 ringan).
2. Rusak sesuatu (matiin Redis, bikin query lambat).
3. Catat **timeline**: jam berapa mulai, jam berapa alert nyala, jam berapa mitigasi, jam berapa pulih.
4. Hitung berapa error budget yang habis.

### Hari 4 · Review (2 jam) — Postmortem blameless

Bikin `postmortems/incident-001.md`:

```markdown
# Incident 001 — <judul singkat>

## Metadata
Severity: SEV2 | Durasi: 34 menit | Penulis: <lo> | Tanggal: YYYY-MM-DD

## Impact
- 12% request gagal selama 34 menit
- ~4.200 request terdampak
- Error budget 30 hari terpakai 8%

## Timeline (semua waktu UTC)
- 10:14 — deploy versi sha-abc123
- 10:31 — alert HighErrorBudgetBurn firing
- 10:33 — on-call mulai investigasi
- 10:41 — hipotesis: connection pool habis
- 10:47 — rollback ke sha-9f8e21
- 10:48 — error rate normal

## Root cause
<satu kalimat yang jujur>

## Contributing factors (5-whys)
1. Kenapa? ... → kenapa? ... (sampai ketemu akar proses, bukan akar orang)

## What went well
- Alert kepicu dalam 17 menit
- Rollback jalan tanpa drama

## What went wrong
- Dashboard nggak nunjukin saturasi connection pool
- Nggak ada alert buat saturasi

## Action items
| # | Tindakan | Owner | Tenggat | Prioritas |
| 1 | Tambah panel connection pool + alert | lo | +1 minggu | P1 |
```

**Blameless artinya:** fokus ke "sistem mana yang memungkinkan ini terjadi", bukan "siapa yang salah". Kalau postmortem lo nyebut nama orang sebagai penyebab, postmortem itu gagal.

### Jebakan umum minggu 12

- **SLO diukur dari metrik internal aplikasi**, bukan dari sisi user. Aplikasi yang crash sebelum nyatat metrik nggak pernah kehitung sebagai error — padahal itu error terburuk.
- **Alert statis "error rate > 1%"** tanpa burn rate → banyak false positive, terutama di traffic rendah.
- **Service dengan traffic rendah** butuh aturan khusus: 1 request gagal per jam bisa jadi burn rate 1000×.
- Postmortem tanpa action item ber-owner dan tenggat = curhatan, bukan postmortem.
- Severity nggak ditetapkan di 5 menit pertama → respons jadi kacau.

### Acceptance criteria minggu 12

1. SLI bisa dihitung dari metrics yang udah dikumpulkan (sumber data + window jelas).
2. SLO punya periode evaluasi dan aturan pengukuran yang jelas.
3. Alert burn rate terpasang dan diuji (`promtool test rules`).
4. Laporan insiden berisi bukti, dampak, dan tindakan pencegahan.

**Yang masuk GitHub:** `docs/slo.md`, `docs/incident-response.md`, `postmortems/incident-001.md`.

---

# Bulan 4: Automation, Cloud, dan Reliability Engineering

Tujuan bulan ini: buktiin lo bisa bangun ulang semuanya, ukur kapasitasnya, dan pulihkan saat rusak.

---

## Minggu 13 · Terraform dan Infrastructure as Code

**Target:** infrastruktur reproducible dan bisa direview lewat Git.

### Hari 1 · Teori (2 jam)

Provider, resource, variable, output, state, dependency graph, drift.  
Referensi: [Terraform Tutorials](https://developer.hashicorp.com/terraform/tutorials).

**Konteks yang wajib lo tahu sebelum milih:**

|             | Terraform                                           | OpenTofu                                                            |
| ----------- | --------------------------------------------------- | ------------------------------------------------------------------- |
| Lisensi     | BUSL 1.1 (bukan open source) sejak 2023             | MPL-2.0 (open source, Linux Foundation)                             |
| Kepemilikan | HashiCorp — **diakuisisi IBM, selesai 27 Feb 2025** | Komunitas                                                           |
| Ekosistem   | Registry utama, paling lengkap                      | Registry sendiri (3.900+ provider)                                  |
| Fitur khas  | HCP Terraform, Sentinel                             | State encryption fleksibel, OCI registry (sejak 1.10, Jun **2025**) |

**Rekomendasi gue:** pakai **OpenTofu** buat lab (gratis, bebas, sintaksnya sama). Tapi **pahami bahwa di kantor lo kemungkinan besar ketemu Terraform**, dan dua-duanya bisa hidup berdampingan. Yang penting lo ngerti konsepnya: **state, plan, apply, drift.**

### Hari 2 · Praktik (2 jam)

```hcl
# infra/terraform/main.tf
terraform {
  required_version = "~> 1.9"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"          # PINNING WAJIB: cegah breaking change dadakan
    }
  }
  backend "s3" {
    bucket         = "sre-lab-tfstate"
    key            = "lab/network.tfstate"
    region         = "ap-southeast-1"
    dynamodb_table = "tf-locks"   # locking: cegah dua apply nabrak
    encrypt        = true         # state bisa berisi secret!
  }
}

variable "environment" {
  type    = string
  default = "lab"
}

variable "allowed_ssh_cidr" {
  type        = string
  description = "CIDR yang boleh SSH. Jangan pernah 0.0.0.0/0 di produksi."
}

resource "aws_vpc" "main" {
  cidr_block           = "10.20.0.0/16"
  enable_dns_hostnames = true
  tags = { Name = "${var.environment}-vpc", Env = var.environment }
}

output "vpc_id" {
  value       = aws_vpc.main.id
  description = "ID VPC lab"
}
```

```bash
terraform fmt -check       # atau: tofu fmt -check
terraform init
terraform validate
terraform plan -out=tfplan # SIMPAN PLAN-nya
terraform show -json tfplan > plan.json
```

### Hari 3 · Eksperimen (2 jam)

```bash
# 1. Ubah konfigurasi, lihat bedanya
terraform plan -out=tfplan2
terraform show -json tfplan2 | jq '.resource_changes[].change.actions'

# 2. Drift detection: ubah manual di cloud, lalu plan lagi
terraform plan -refresh-only

# 3. Import resource yang sudah ada (bukan terraform import CLI)
#    import block tersedia sejak TF 1.5
```

```hcl
import {
  to = aws_vpc.main
  id = "vpc-0abc123"
}
```

```bash
# 4. Scan keamanan IaC (tfsec sudah dilebur ke Trivy)
trivy config .

# 5. Cek apakah state berisi rahasia
terraform state list
terraform state show aws_db_instance.main | grep -i password
```

**Jangan pernah:** commit file state (`*.tfstate`) ke Git. Isinya bisa berisi password database dalam teks biasa.

### Hari 4 · Review (2 jam)

Tulis `docs/infrastructure.md`:

| Bagian       | Isi                                                                       |
| ------------ | ------------------------------------------------------------------------- |
| Provisioning | Perintah init→plan→apply, plus prasyarat (kredensial, backend)            |
| Perubahan    | Alur review: PR → `plan -out` → review → `apply tfplan`                   |
| Cleanup      | `terraform destroy` dan **resource apa yang bisa jadi biaya tak terduga** |
| Biaya        | Estimasi per resource, dan budget alert                                   |

### Jebakan umum minggu 13

- `terraform apply` **tanpa `-out`** → rencana yang dieksekusi bisa beda dari yang lo review.
- Pakai **workspace** buat bedain dev/prod dalam satu state → blast radius gede banget. Pakai **directory per environment** + backend terpisah.
- State di repo Git atau lokal → secret bocor plaintext.
- Provider tanpa version pin → `init` ulang di CI bisa narik versi baru yang nabrak.
- Lupa `terraform destroy` → VPS dan NAT gateway jadi tagihan bulanan.

### Acceptance criteria minggu 13

1. `terraform validate` (atau `tofu validate`) lolos, `fmt -check` bersih.
2. Perubahan bisa direview lewat `plan -out` sebelum apply.
3. Provisioning dan cleanup terdokumentasi; biaya dipahami sebelum apply.
4. Nggak ada state file atau secret ke-commit.

**Yang masuk GitHub:** `infra/terraform/` — file `.tf`, contoh variable, `docs/infrastructure.md`.

---

## Minggu 14 · Cloud networking dan security

**Target:** paham gimana service jalan di jaringan cloud yang aman.

### Hari 1 · Teori (2 jam)

VPC, subnet, route table, security group/firewall, IAM, load balancer, object storage, managed database.  
**Pilih SATU provider.** Jangan coba AWS + GCP + Azure sekaligus — lo cuma bakal hapal nama produknya, bukan konsepnya.

### Hari 2 · Praktik (2 jam)

Bikin diagram jaringan dulu **sebelum** nulis konfigurasi. Pola yang benar:

```
                 Internet
                     │
              [Load Balancer]        ← publik: cuma 80/443
                     │
        ┌────────────┴────────────┐
   Public Subnet             Private Subnet
   (NAT Gateway)             ┌──────────────┐
                             │  App (Go)    │  ← nggak punya public IP
                             │  Postgres    │  ← nggak punya public IP
                             │  Redis       │  ← nggak punya public IP
                             └──────────────┘
                     │
              Object Storage (backup, terenkripsi)
```

Aturannya:

| Yang boleh publik      | Yang tidak boleh                  |
| ---------------------- | --------------------------------- |
| Load balancer: 80/443  | Database (5432, 6379)             |
| (opsional) bastion/SSM | SSH ke app server langsung        |
| —                      | Port metrik internal (9090, 3100) |

```hcl
resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.small"
  subnet_id     = aws_subnet.private.id            # private, bukan public
  iam_instance_profile = aws_iam_instance_profile.app.name  # role, BUKAN access key

  metadata_options {                                # IMDSv2 WAJIB
    http_tokens   = "required"
    http_endpoint = "enabled"
  }

  vpc_security_group_ids = [aws_security_group.app.id]
}

resource "aws_security_group" "app" {
  name   = "sre-lab-app"
  vpc_id = aws_vpc.main.id

  ingress {                       # cuma dari load balancer
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.lb.id]
  }

  egress {                        # keluar lewat NAT
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

### Hari 3 · Eksperimen (2 jam)

```bash
# 1. Buktikan port DB nggak bisa diakses dari internet
nmap -Pn -p 5432 <public-ip-app>        # harus filtered/closed

# 2. Buktikan port publik bisa
curl -fsS -o /dev/null -w '%{http_code}\n' https://lab.domain-lo/

# 3. Cek DNS dan TLS
dig +short lab.domain-lo
openssl s_client -connect lab.domain-lo:443 -servername lab.domain-lo -brief

# 4. Akses admin tanpa SSH key permanen
aws ssm start-session --target i-0abc   # zero standing access

# 5. Budget alert — pasang SEBELUM deploy
aws budgets create-budget --account-id $ACC --budget file://budget.json \
  --notifications-with-subscribers file://notify.json
```

### Hari 4 · Review (2 jam)

Tulis `docs/security-review.md`:

| Area                  | Pertanyaan yang dijawab                                |
| --------------------- | ------------------------------------------------------ |
| **Attack surface**    | Port apa yang terbuka dari internet? Siapa yang perlu? |
| **Hak akses**         | Pakai role atau user permanen? Apa scope izinnya?      |
| **Eksposur jaringan** | Ada resource berharga di subnet publik?                |
| **Backup**            | Di mana, terenkripsi, bisa di-restore?                 |
| **Biaya**             | Berapa per bulan? Ada budget alert?                    |

**Fakta yang harus bikin lo waspada:** AWS Free Tier berubah **15 Juli 2025** — bukan lagi "12 bulan gratis", melainkan kredit $100 (+$100 tambahan) yang berlaku **6 bulan**. Habis itu, tagihan jalan. NAT gateway (~$0,045/jam + biaya data) adalah tersangka utama tagihan lab yang ditinggal nyala.

### Jebakan umum minggu 14

- Node di public subnet "biar gampang SSH" → pakai private subnet + SSM/NAT.
- IAM policy `Action: "*"` / `Resource: "*"` pas debugging, lupa dibalikin.
- Nggak pasang budget alert sebelum deploy.
- NAT gateway ditinggal nyala seminggu.
- Security group inbound `0.0.0.0/0` ke port database.

### Acceptance criteria minggu 14

1. Cuma port yang dibutuhkan yang bisa diakses dari luar.
2. Akses administrasi dibatasi (role + least privilege, bukan user permanen).
3. Diagram jaringan sesuai konfigurasi aktual.
4. Budget alert aktif.

**Yang masuk GitHub:** `labs/week-14-cloud/` — diagram jaringan, Terraform, `docs/security-review.md`.

---

## Minggu 15 · Kubernetes dan orchestration

**Target:** paham konsep orchestration, dan **kapan K8s itu perlu**.

### Hari 1 · Teori (2 jam)

Pod, Deployment, ReplicaSet, Service, ConfigMap, Secret, probe, requests/limits, rolling update.  
Referensi: [Kubernetes Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/).

**Versi & fakta yang relevan:**

- **Kubernetes 1.34 rilis 27 Agustus 2025** (58 enhancement). Yang penting buat lo: `spec.podReplacementPolicy` stable, Pod-level resources beta, **dukungan containerd 1.x berakhir di v1.35**.
- **Ingress NGINX di-retire**: maintenance best-effort cuma sampai Maret 2026, setelahnya nggak ada security fix. Rekomendasi resmi: pindah ke **Gateway API**.
- **Pod Security Admission** (built-in sejak 1.25) menggantikan PodSecurityPolicy yang udah dihapus.

### Hari 2 · Praktik (2 jam)

```bash
# kind = paling dekat ke upstream dan cepat buat CI
kind create cluster --name sre --image kindest/node:v1.34.0
kubectl cluster-info
kubectl get nodes
```

Roadmap lo mulai dari `nginx` buat kenalan workflow — boleh, tapi **jangan berhenti di situ**. Deploy image API lo sendiri:

```yaml
# kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sre-api
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate: { maxSurge: 1, maxUnavailable: 0 }   # nol downtime
  selector:
    matchLabels: { app: sre-api }
  template:
    metadata:
      labels: { app: sre-api }
    spec:
      terminationGracePeriodSeconds: 30
      securityContext:
        runAsNonRoot: true
        seccompProfile: { type: RuntimeDefault }
      containers:
        - name: api
          image: ghcr.io/<user>/sre-api:sha-abc123    # BUKAN :latest
          ports: [{ containerPort: 8080 }]
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits:   { cpu: 500m, memory: 256Mi }
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities: { drop: ["ALL"] }
          startupProbe:
            httpGet: { path: /healthz, port: 8080 }
            failureThreshold: 30
            periodSeconds: 2
          readinessProbe:
            httpGet: { path: /readyz, port: 8080 }
            periodSeconds: 5
            failureThreshold: 2
          livenessProbe:
            httpGet: { path: /healthz, port: 8080 }    # Cuma cek proses, BUKAN dependency
            periodSeconds: 10
            failureThreshold: 3
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata: { name: sre-api }
spec:
  minAvailable: 2
  selector: { matchLabels: { app: sre-api } }
```

```bash
kubectl apply -f kubernetes/
kubectl rollout status deployment/sre-api
kubectl get pods -w
kubectl port-forward svc/sre-api 8080:80
curl -fsS http://localhost:8080/readyz
```

### Hari 3 · Eksperimen (2 jam)

```bash
kubectl scale deployment sre-api --replicas=5
kubectl delete pod <pod-name>              # lihat dia diganti otomatis
kubectl rollout history deployment/sre-api
kubectl set image deployment/sre-api api=ghcr.io/<user>/sre-api:sha-bad123
kubectl rollout status deployment/sre-api  # harus gagal kalau probe bener
kubectl rollout undo deployment/sre-api    # balikin
kubectl describe pod <pod-name>            # baca Events — ini sumber kebenaran
kubectl logs <pod-name> --previous         # log dari pod yang crash
kubectl debug -it <pod-name> --image=busybox:1.36 --target=api   # ephemeral container
kubectl top pod --sort-by=memory
```

**Eksperimen yang paling berharga:** set `readinessProbe` ke path yang nggak ada, lalu deploy. Lo bakal lihat rollout mandek tanpa bikin service down. Itu contoh nyata kenapa probe itu pelindung, bukan formalitas.

### Hari 4 · Review (2 jam)

Tulis perbandingan yang jujur:

| Hal              | Docker Compose            | Kubernetes                                           |
| ---------------- | ------------------------- | ---------------------------------------------------- |
| Setup            | 1 file, 1 perintah        | Banyak objek, learning curve curam                   |
| Restart otomatis | `restart: unless-stopped` | ReplicaSet jaga desired state                        |
| Scaling          | manual / Swarm            | `scale`, HPA otomatis                                |
| Rollout          | **nggak ada** (recreate)  | rolling update + undo                                |
| Self-healing     | cuma restart container    | reschedule ke node lain                              |
| Networking       | DNS internal compose      | Service + Gateway API                                |
| **Overhead**     | hampir nol                | harus paham 15+ objek, resource, storage, networking |

**Kapan K8s itu benar-benar perlu:** banyak service, butuh auto-scaling, butuh multi-node self-healing, tim sudah punya platform engineer.  
**Kapan K8s itu beban:** satu service kecil, traffic rendah, tim 2 orang. Compose + satu VPS lebih murah dan lebih sedikit insiden.

Jawaban yang matang di interview itu bukan "K8s keren", tapi "K8s menyelesaikan X dengan biaya Y; di kasus ini Y belum sepadan, dan ini alasannya."

### Jebakan umum minggu 15

- `limits` tanpa `requests` (atau sebaliknya) → QoS jadi Burstable/BestEffort, jadi korban pertama saat node kehabisan memori.
- Nggak ada **startupProbe** buat aplikasi yang booting lama → CrashLoopBackOff palsu.
- **Liveness = readiness** → restart massal saat cuma overload.
- Masih pakai Ingress NGINX buat cluster baru (udah nggak dapat security fix).
- Image `:latest` → `rollout undo` nggak deterministik.

### Acceptance criteria minggu 15

1. Deployment **dan** Service jalan dengan image API milik lab.
2. Pod yang dihapus diganti otomatis sesuai desired state.
3. Rollout, rollback, dan probe failure bisa diamati dan dijelaskan.
4. Non-root + `readOnlyRootFilesystem` + requests/limits terpasang; PSA `restricted` aktif.

**Yang masuk GitHub:** `labs/week-15-kubernetes/` — manifest, diagram, hasil eksperimen.

---

## Minggu 16 · Load testing, chaos, dan final incident simulation

**Target:** buktiin lo bisa ukur kapasitas, nemu bottleneck, dan pulihin sistem.

### Hari 1 · Teori (2 jam)

Capacity planning, Little's Law, queueing, saturation, retry storm, exponential backoff, timeout budget, circuit breaker.  
Referensi: [Google SRE Workbook](https://sre.google/workbook/).

**Rumus yang wajib lo pakai minggu ini:**

> **Little's Law: L = λ × W**  
> Jumlah request dalam sistem (L) = kedatangan per detik (λ) × waktu yang dihabiskan di sistem (W).

Kenapa penting? Karena dia ngasih tahu **berapa banyak concurrency yang lo butuh** buat capai throughput tertentu. Kalau target 500 req/detik dengan latency 200ms, lo butuh minimal `500 × 0,2 = 100` request in-flight.

> **Utilisasi: ρ = λ / μ**  
> Saat ρ mendekati 1, **latency meledak bukan linear, tapi eksponensial**. Ini definisi **saturasi**.

Ini alasan kenapa "server kita kuat 800 req/detik" itu omong kosong kalau nggak nyebut latency di angka itu.

**Tiga pelindung yang harus lo terapkan:**

| Pola                             | Masalah yang dicegah                         | Nilai awal yang masuk akal                                       |
| -------------------------------- | -------------------------------------------- | ---------------------------------------------------------------- |
| **Timeout budget**               | Request numpuk nunggu dependency mati        | Total 2s: proxy 2s, app→DB 1s, app→Redis 200ms (menurun per hop) |
| **Exponential backoff + jitter** | Retry storm: semua client retry bareng       | `min(2^n × base, cap)` + jitter acak                             |
| **Circuit breaker**              | Terus nembak dependency yang udah jelas mati | Buka setelah 50% error dalam 10s, setengah-buka setelah 30s      |

Kenapa jitter itu penting: tanpa jitter, semua client retry **pada detik yang sama** → traffic jadi "gerombolan" yang makin nabrak.

### Hari 2 · Praktik (2 jam)

```js
// loadtest/loadtest.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    ramp: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50 },    // naik pelan
        { duration: '3m', target: 50 },    // tahan (steady state)
        { duration: '1m', target: 0 },     // turun
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<300'],     // GAGAL kalau p95 > 300ms
    http_req_failed:   ['rate<0.01'],     // GAGAL kalau error > 1%
  },
};

export default function () {
  const r = http.get('http://localhost/orders', { timeout: '2s' });  // timeout budget
  check(r, { 'status 200': (x) => x.status === 200 });
  sleep(1);
}
```

```bash
k6 run loadtest/loadtest.js
k6 run --summary-mode=full --out json=loadtest/results.json loadtest/loadtest.js
```

> **Catatan versi:** k6 **v1.0.0 rilis 6 Mei 2025** (native TypeScript, SemVer, 2 tahun support). Sekarang udah **v2.x** — **pin versi di CI**, karena beda major bisa ngubah perilaku.

Yang diukur dan dicatat per tahap beban:

| VU  | Throughput (req/s) | p50 | p95 | p99 | Error rate | CPU | Memori | Koneksi DB |
| --- | ------------------ | --- | --- | --- | ---------- | --- | ------ | ---------- |
| 10  |                    |     |     |     |            |     |        |            |
| 50  |                    |     |     |     |            |     |        |            |
| 100 |                    |     |     |     |            |     |        |            |
| 200 |                    |     |     |     |            |     |        |            |

**Yang lo cari:** titik di mana p95 mulai naik tajam sementara throughput datar. Itu kapasitas saturasi lo.

### Hari 3 · Eksperimen (2 jam) — Game day

Satu simulasi insiden terencana, 6 langkah:

1. **Traffic normal** — k6 jalan di 30 VU, catat baseline.
2. **Rusak** — pilih salah satu: bikin DB lambat (`SELECT pg_sleep(0.5)`), matiin Redis, atau inject latency 200ms pakai [toxiproxy](https://github.com/Shopify/toxiproxy).
3. **Amati** — alert apa yang nyala? Berapa lama? Dashboard apa yang paling kepakai?
4. **Diagnosis** — lewat metrics → logs → trace. Catat berapa menit yang lo butuh.
5. **Mitigasi & pulih** — rollback / restart / failover. Catat dampaknya.
6. **Ulangi setelah perbaikan** — dan bandingkan angkanya.

Bikin chaos experiment yang **punya hipotesis tertulis**:

```markdown
## Hipotesis
Kalau Redis mati, API tetap balas dalam < 500ms dan error rate < 1%,
karena cache miss jatuh ke database yang masih sanggup.

## Metrik steady-state
- p95 latency < 300ms
- error rate < 1%

## Blast radius
Cuma container Redis di environment lab.

## Abort condition
Kalau error rate > 20% selama 2 menit, eksperimen dihentikan otomatis
dan Redis dinyalakan lagi.
```

> **Chaos tanpa hipotesis dan abort condition itu namanya bukan eksperimen — itu sabotase.** Bedanya cuma di disiplin ini.

### Hari 4 · Final review (2 jam)

1. **Incident report** lengkap (format postmortem minggu 12).
2. **Benchmark sebelum vs sesudah perbaikan** — tabel di atas, dua versi.
3. **Batas kapasitas yang terukur**: "sistem ini sanggup X req/detik dengan p95 < 300ms di spesifikasi A. Bottleneck-nya B, dan ini buktinya."
4. **Rapikan semuanya**: README utama, diagram final, runbook, instruksi reproduksi.

Bikin **README utama** yang bikin reviewer langsung paham dalam 30 detik:

```markdown
# Production Reliability Lab

Satu repository berisi production-like service, automated deployment,
monitoring, alerting, simulasi insiden, dan dokumentasi reliability.

## Cara jalanin dari nol
git clone ... && cd production-reliability-lab
cp .env.example .env
make up
curl http://localhost/health

## Yang ada di sini
- Go HTTP API + PostgreSQL + Redis di balik Nginx
- CI/CD GitHub Actions dengan image ber-tag commit SHA
- Prometheus + Grafana + Loki/Alloy + OpenTelemetry
- SLO, error budget, dan postmortem insiden
- Terraform/OpenTofu, Kubernetes (kind), load test k6

## Bukti yang bisa dicek
| Minggu | Apa yang dibuktikan | Di mana |
| 1 | Diagnosis CPU/OOM/down | labs/week-01-linux/ |
...
| 16 | Kapasitas & pemulihan insiden | labs/week-16-final-capstone/ |
```

### Jebakan umum minggu 16

- Cuma ngukur **rata-rata latency** — yang penting p95/p99 dan error rate.
- Naikin VU terus tanpa cek utilisasi → lo cuma ngukur antrean, bukan kapasitas.
- Load test dari laptop → yang kebatas bandwidth lo sendiri.
- Chaos tanpa abort condition → malah bikin insiden beneran.
- Retry di 3 lapis (client, LB, aplikasi) tanpa jitter → amplifikasi sampai 27×.

### Acceptance criteria minggu 16

1. Load test bisa diulang dan punya threshold yang ngegagalkan CI.
2. Bottleneck didukung data, bukan dugaan.
3. Insiden bisa dideteksi, didiagnosis, dimitigasi, dipulihkan.
4. Hasil sebelum/sesudah bisa dibandingkan kuantitatif.

**Yang masuk GitHub:** `labs/week-16-final-capstone/` — script k6, incident report, hasil benchmark, diagram final.

---

# Lampiran A · AI Agent Workflow Products yang Relevan Buat SRE

> Lo nyebut "AI agent workflow products" di awal. Gue nggak bikin ini jadi fokus utama (roadmap lo udah bagus), tapi gue kasih lampiran pendek karena **ini area yang beneran lagi bergerak cepat** dan bakal ngaruh ke kerjaan SRE.

## Kenapa ini relevan buat SRE, bukan cuma buat "AI engineer"

AI agent workflow itu intinya: **serangkaian langkah di mana LLM bisa manggil tools, baca state, dan ambil keputusan**. Kalau lo ganti "tools" dengan "perintah kubectl, query PromQL, API PagerDuty", hasilnya adalah... asisten on-call.

Tiga kelas produk yang muncul:

| Kelas                          | Contoh                                   | Yang dikerjakan                                              | Status realistis 2025–2026                                             |
| ------------------------------ | ---------------------------------------- | ------------------------------------------------------------ | ---------------------------------------------------------------------- |
| **Platform insiden + AI**      | incident.io, Rootly, PagerDuty           | Ringkas insiden, auto-draft postmortem, COPILOT buat on-call | Sudah dipakai produksi; Rootly bikin API-nya AI-agent-first (2025)     |
| **AIOps di observability**     | Dynatrace, Datadog Bits AI, New Relic AI | Korelasi metric/log/trace, usulin root cause                 | Bagus buat hipotesis awal, **jangan dipakai sebagai kesimpulan akhir** |
| **Workflow automation klasik** | StackStorm, Rundeck, n8n, Temporal       | Runbook otomatis, remidiasi ber-aturan                       | Paling bisa dipercaya karena deterministik                             |

## Perubahan yang perlu lo antisipasi

- **Opsgenie end-of-sale 4 Juni 2025**, dukungan berakhir 2027 → banyak tim migrasi ke Jira Service Management. Kalau lo kebetulan ketemu Opsgenie, tahu konteksnya.
- Platform insiden makin "AI-agent-first": bukan cuma notifikasi, tapi agent yang bisa **nanya ke metrics** dan **nulis draft postmortem**.

## Aturan pakai yang gue saranin

1. **AI bagus buat hipotesis, bukan buat vonis.** Biarkan dia ngusulin "kayaknya connection pool", tapi lo tetap harus buktiin dengan query.
2. **Jangan kasih agent akses tulis ke produksi tanpa approval.** Tool yang bisa `kubectl delete` itu harus ada human-in-the-loop.
3. **Runbook tetap harus bisa dijalanin manual.** Kalau cuma agent yang ngerti, saat agent-nya error lo nggak punya apa-apa.
4. **Tulis audit trail.** Tiap tindakan agent harus ke-log seperti tindakan manusia.

Sumber buat mulai baca: [incident.io — 5 AI-powered SRE tools](https://incident.io/blog/sre-ai-tools-transform-devops-2025), [The New Stack — AI DevOps vs SRE agents](https://thenewstack.io/ai-devops-vs-sre-agents-compare-ai-incident-response-tools/).

> Pendapat gue: buat 16 minggu ini, **jangan sisipin AI agent ke lab**. Kuasai dulu fundamentals-nya. Agent yang jalan di atas observability yang buruk cuma bakal nge-hasilin hipotesis yang salah dengan lebih cepat.

---

# Lampiran B · Checklist Kelulusan Pribadi

Roadmap lo bilang: *"Jangan tandai minggu selesai hanya karena tutorialnya sudah ditonton."* Ini versi yang bisa lo pakai buat nguji diri sendiri.

## Per minggu — tanya 4 hal ini sebelum centang

1. **Bisa reproduksi?** Ada perintahnya, dan perintah itu jalan dari repo yang baru di-clone.
2. **Ada buktinya?** Output, screenshot, angka — tersimpan di `evidence/` dan sudah disanitasi.
3. **Bisa dijelaskan?** Lo bisa nerangin ke orang lain tanpa buka catatan. Kalau nggak bisa, belum selesai.
4. **Ada runbook-nya?** Kalau ini rusak lagi, ada yang bisa lo (atau orang lain) lakuin tanpa nebak.

## Per bulan — pertanyaan level atas

| Bulan | Kalau lo bisa jawab ini, bulan itu lulus                                                                                               |
| ----- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **1** | "Service lambat. Tanpa lihat dashboard, apa 5 perintah pertama yang lo jalanin, dan kenapa urutannya begitu?"                          |
| **2** | "Deploy barusan bikin error naik. Apa langkah rollback lo, dan berapa lama sampai normal?"                                             |
| **3** | "Berapa availability layanan lo 30 hari terakhir, dan dari mana angka itu?"                                                            |
| **4** | "Sistem ini sanggup berapa req/detik sebelum p95 rusak? Bottleneck-nya apa? Kalau rusak jam 3 pagi, apa yang bikin dia pulih sendiri?" |

## Tanda bahaya bahwa lo cuma nonton

- Semua README isinya **salinan dokumentasi**, nggak ada output dari mesin lo sendiri.
- `evidence/` kosong atau cuma ada screenshot halaman dokumentasi.
- Bisa jalanin `kubectl apply` tapi nggak bisa jelasin apa yang terjadi setelahnya.
- Nggak pernah sengaja ngerusak apa pun. (Ini tanda paling jelas.)

---

# Lampiran C · Referensi Utama per Topik

| Topik              | Sumber primer                                                                                                                                                                                                                                                                |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Linux & kernel     | [man7.org](https://man7.org/linux/man-pages/), [PSI docs](https://www.kernel.org/doc/html/latest/accounting/psi.html), [cgroup v2](https://docs.kernel.org/admin-guide/cgroup-v2.html)                                                                                       |
| Networking         | [Cloudflare Learning Center](https://www.cloudflare.com/learning/network-layer/), [Cloudflare Radar](https://blog.cloudflare.com/radar-2025-year-in-review/)                                                                                                                 |
| Docker             | [Docker Get Started](https://docs.docker.com/get-started/), [Compose file reference](https://docs.docker.com/reference/compose-file/services/)                                                                                                                               |
| Nginx              | [Beginner's Guide](https://nginx.org/en/docs/beginners_guide.html), [ngx\_http\_proxy\_module](https://nginx.org/en/docs/http/ngx_http_proxy_module.html)                                                                                                                    |
| CI/CD              | [GitHub Actions docs](https://docs.github.com/en/actions), [Artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations)                                                                                                               |
| Go shutdown        | [signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext), [http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)                                                                                                                                      |
| PostgreSQL         | [Backup & Restore](https://www.postgresql.org/docs/current/backup.html), [Continuous Archiving](https://www.postgresql.org/docs/current/continuous-archiving.html), [pgBackRest](https://pgbackrest.org/)                                                                    |
| PRR                | [Evolving SRE Engagement Model](https://sre.google/sre-book/evolving-sre-engagement-model/), [Launch Checklist](https://sre.google/sre-book/launch-checklist/)                                                                                                               |
| Prometheus         | [Getting Started](https://prometheus.io/docs/tutorials/getting_started/), [Histogram practices](https://prometheus.io/docs/practices/histograms/), [Native histograms](https://prometheus.io/docs/specs/native_histograms/)                                                  |
| Grafana            | [Alerting rules](https://grafana.com/docs/grafana/latest/alerting/alerting-rules/), [Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/), [Promtail → Alloy migration](https://grafana.com/docs/loki/latest/setup/migrate/migrate-to-alloy/) |
| OpenTelemetry      | [OTel docs](https://opentelemetry.io/docs/), [Semconv HTTP](https://opentelemetry.io/docs/specs/semconv/http/http-metrics/), [W3C Trace Context](https://www.w3.org/TR/trace-context/)                                                                                       |
| SLO                | [Implementing SLOs](https://sre.google/workbook/implementing-slos/), [Alerting on SLOs](https://sre.google/workbook/alerting-on-slos/), [Error budget policy](https://sre.google/workbook/error-budget-policy/)                                                              |
| Postmortem         | [Postmortem culture](https://sre.google/sre-book/postmortem-culture/)                                                                                                                                                                                                        |
| Terraform/OpenTofu | [Terraform tutorials](https://developer.hashicorp.com/terraform/tutorials), [OpenTofu](https://opentofu.org/)                                                                                                                                                                |
| Kubernetes         | [K8s basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/), [v1.34 release](https://kubernetes.io/blog/2025/08/27/kubernetes-v1-34-release/), [Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)     |
| Load test & chaos  | [k6 docs](https://grafana.com/docs/k6/latest/), [Chaos Mesh](https://chaos-mesh.org/), [toxiproxy](https://github.com/Shopify/toxiproxy)                                                                                                                                     |

---

**Penutup.** Target 16 minggu ini bukan hapal 14 tools. Targetnya: **kebiasaan** — ukur, deteksi, diagnosis, mitigasi, pulihkan, lalu cegah hal yang sama terulang. Tools-nya bakal ganti (Promtail aja udah EOL). Kebiasaannya nggak.
