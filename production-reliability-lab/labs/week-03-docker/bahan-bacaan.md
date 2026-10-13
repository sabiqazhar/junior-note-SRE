# Minggu 03 · Docker production fundamentals — Bahan Bacaan

> Dokumen ini berdiri sendiri. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham konsepnya. Baru setelah itu kerjakan lab.md.

---

## 1. Kenapa minggu ini penting

Lo udah jago bikin Dockerfile dan `docker run`. Tapi di produksi, beda tipis antara "jalan di laptop" dan "jalan dengan aman". Contoh nyata: lo deploy API, besok pagi user lapor "server sering restart sendiri". Lo buka dashboard, lihat container `restarting (1)` terus. Kenapa? Karena lo nggak kasih **healthcheck**, nggak kasih **resource limit**, dan aplikasi lo **hang** (bukan balas 503) pas Redis mati.

Minggu ini ngajarin operasi container yang terkontrol: lifecycle, resource, network, dan—paling penting—**degradasi terkendali**. Artinya: kalau dependency mati, aplikasi lo harus gagal *cepat dan jelas*, bukan nggantung sampai timeout 60 detik. Itu yang membedakan "lab" dan "calon insiden".

Fakta versi 2025–2026 yang wajib lo tahu: `docker-compose` (v1, Python) **sudah dihapus** — sekarang `docker compose` (v2, Go). PostgreSQL **18 GA 25 Sep 2025** (default `data_checksum ON`, ada AIO `io_method`, `md5` deprecated). Redis pakai `7-alpine`. Semua itu yang lo pakai di lab.

---

## 2. Konsep inti

### 2.1 Image layer & multi-stage build

Image Docker itu tumpukan layer yang read-only. Tiap `RUN`, `COPY`, `ADD` bikin layer baru. Cache layer ini yang bikin build cepat — tapi juga jebakan: `COPY . .` di awal bikin cache hancur tiap lo ubah kode.

**Multi-stage build** = pakai image besar buat compile, salin cuma binary ke image kecil (distroless/alpine). Hasilnya image ringan, attack surface kecil, start cepat.

### 2.2 Container lifecycle & restart policy

Siklus: `created → running → paused/stopped → removed`. `restart: unless-stopped` bikin container hidup lagi kalau host reboot atau proses mati — tapi HATI-HATI: ini nggak sama dengan "sehat". Container bisa restart terus (`restarting`) kalau aplikasinya panic saat start.

### 2.3 Bridge network & DNS internal Compose

Tiap `docker compose` bikin bridge network sendiri, plus **DNS internal**: service bisa panggil service lain pakai nama service (`api` → `redis`). Tapi ini cuma di dalam jaringan itu. Laptop lo nggak bisa resolve `redis`. Makanya `postgres` & `redis` **nggak usah punya `ports:`** — mereka cuma reachable dari jaringan internal.

### 2.4 Volume & persistensi

Data di dalam container itu ilang kalau container dihapus. `volume` (atau bind mount) nyimpen di host. Trap terbesar: `docker compose down -v` = volume ikut dihapus = **data hilang**. `down` tanpa `-v` aman.

### 2.5 Healthcheck & depends_on condition

Dua mekanisme beda yang sering disamain:
- **`healthcheck`** → "apakah *proses ini* sehat?" (mis. `pg_isready`, `redis-cli ping`).
- **`depends_on: condition: service_healthy`** → "jangan start aku sebelum dependency-nya *siap*".

`depends_on` bentuk pendek (`depends_on: [redis]`) itu jebakan: dia cuma nunggu container **start**, bukan **siap**. Postgres butuh beberapa detik buat init — API bisa kena connection error kalau start bareng.

### 2.6 cgroups v2 & resource limit

Container itu cuma proses di bawah cgroup. `deploy.resources.limits` (di Compose non-Swarm) ngasih batas CPU/memori/pids. Tanpa ini, satu container bisa makan semua RAM host → host yang kena OOM, bukan container. Di minggu 1 lo udah lihat `memory.events`; di sini lo yang pasang batasnya.

### 2.7 Controlled degradation (inti minggu ini)

Ini yang paling penting. Saat Redis mati, API lo harus:
- **fail fast** → balas 503 dalam milidetik, atau
- **degrade** → lanjut tanpa cache, atau
- **jangan hang** → jangan nunggu Redis sampai `proxy_read_timeout` habis.

Kalau API lo *hang* pas Redis mati, berarti ada bug nyata di error handling. Temuin itu minggu ini, perbaiki minggu 6.

### 2.8 `docker compose` vs Swarm

`deploy.replicas` & `update_config` itu fitur **Swarm**. Di Compose biasa mereka **diabaikan**. Yang tetap jalan di Compose non-Swarm: `deploy.resources.limits`. Jangan kaget kalau `replicas: 3` nggak nambah container.

---

## 3. Mekanisme di balik layar: dari `up` sampai container mati

1. **`docker compose up -d --build --wait`** → Compose baca `compose.yaml`, build image, buat network & volume, start container, lalu `--wait` nunggu **healthcheck OK**.
2. **Network dibikin** → bridge + DNS internal. `api` bisa resolve `redis` jadi IP internal.
3. **depends_on dievaluasi** → `api` nggak start sampai `postgres` & `redis` `service_healthy`.
4. **Healthcheck jalan** → Compose ngecek tiap `interval`. Status masuk ke `docker compose ps` kolom Health.
5. **Aplikasi jalan** → `sre-api` coba connect Postgres/Redis. Kalau sehat, oke.
6. **Redis mati (eksperimen)** → `docker compose stop redis`. API yang nggak punya timeout di client bakal **hang** di tiap request yang butuh Redis.
7. **OOM (eksperimen)** → `docker run --memory=64m`. cgroup kena batas → kernel nembak proses → container exit, lalu `restart: unless-stopped` nyalain lagi → loop `restarting`.
8. **Bukti** → `docker inspect` (state, exit code), `docker events`, `docker stats`, `docker compose logs`.

---

## 4. Glosarium mini

| Istilah | Arti dengan bahasa sederhana | Contoh konkret |
| --- | --- | --- |
| **Image layer** | Lapisan read-only penyusun image | `COPY . .` bikin layer baru |
| **Multi-stage build** | Build di image besar, salin binary ke image kecil | `FROM golang:1.23 AS build` → `FROM gcr.io/distroless/static` |
| **Bridge network** | Jaringan privat antar container | `docker compose` bikin 1 bridge per stack |
| **DNS internal** | Nama service = hostname di dalam network | `api` resolve `redis` → `172.20.0.3` |
| **Volume** | Penyimpanan persistent di host | `pgdata:/var/lib/postgresql/data` |
| **Healthcheck** | Tes "apakah aku sehat?" | `pg_isready -U app` |
| **depends_on (condition)** | "jangan start sebelum dia sehat" | `condition: service_healthy` |
| **Resource limit** | Batas CPU/RAM/pids lewat cgroup | `deploy.resources.limits.memory: 256M` |
| **Controlled degradation** | Gagal cepat & jelas, bukan hang | Redis mati → API balas 503 |
| **restart policy** | Aturan hidup-mati container | `unless-stopped` |
| **Swarm vs Compose** | Orchestrator vs tool lokal | `replicas` cuma jalan di Swarm |
| **Postgres 18** | PG GA 25 Sep 2025, checksum ON | `initdb` default `data_checksum ON` |

---

## 5. Bahan bacaan wajib

| # | Sumber (link) | Bagian yang dibaca & apa yang dicari | Estimasi |
| --- | --- | --- | --- |
| 1 | [Docker Get Started](https://docs.docker.com/get-started/) | Pahami image vs container vs registry secara umum. Lo udah mahir, jadi cukup skim bagian "Container" & "Image". | 20 menit |
| 2 | [Compose file reference — services](https://docs.docker.com/reference/compose-file/services/) | Fokus `depends_on` (condition), `healthcheck`, `deploy.resources.limits`, `restart`. Ini sumber otoritatif buat compose.yaml lo. | 40 menit |
| 3 | [Dockerfile reference](https://docs.docker.com/reference/dockerfile/) | Fokus `FROM` (multi-stage), `COPY`, `USER`, `HEALTHCHECK`, `ENTRYPOINT` vs `CMD`. | 30 menit |
| 4 | [Control startup order in Compose](https://docs.docker.com/compose/how-tos/startup-order/) | Cara bikin service nunggu dependency siap pakai `condition: service_healthy`. Langsung dipakai di lab. | 15 menit |
| 5 | [Docker SDK/CLI — resource limits](https://docs.docker.com/reference/cli/docker/run/#memory) | Cek flag `--memory`, `--cpus`, `--pids-limit` buat eksperimen OOM di Hari 3. | 15 menit |
| 6 | [PostgreSQL 18 release notes](https://www.postgresql.org/about/news/postgresql-18-released-2965/) | Konfirmasi fitur PG 18: `data_checksum` default ON, AIO `io_method`, `md5` deprecated. | 15 menit |

---

## 6. Self-check (harus bisa dijawab tanpa buka Google)

1. Apa bedanya `depends_on` bentuk pendek dan `condition: service_healthy`? Kenapa yang pendek itu jebakan?
2. Apa bedanya `healthcheck` dan `depends_on`? Mana yang ngecek "dependency sehat"?
3. Tanpa `USER` di Dockerfile, container jalan sebagai siapa? Kenapa itu bahaya?
4. `docker compose down` vs `docker compose down -v` — mana yang bikin data hilang?
5. Lo pasang `deploy.resources.limits.memory: 256M`, tapi container tetep bikin host OOM. Kenapa bisa? (petunjuk: cgroup vs host)
6. Redis mati dan API lo malah *hang* (bukan 503). Itu gejala apa, dan harusnya gimana?
7. `replicas: 3` di compose.yaml lo nggak nambah container. Kenapa?

---

## 7. Yang sering salah paham

| Mitos | Kenyataannya |
| --- | --- |
| "`docker-compose` masih bisa dipakai." | v1 (Python) sudah dihapus. Pakai `docker compose` (v2). Tutorial lama dengan strip `-` usang. |
| "`depends_on` nunggu service siap." | Bentuk pendek cuma nunggu *start*. Pakai `condition: service_healthy` buat nunggu *siap*. |
| "Healthcheck `test: ["CMD","true"]` cukup." | Itu selalu sehat walau aplikasi mati. Tes yang beneran ngecek kesiapan. |
| "Image `:latest` + cache = aman." | `:latest` bikin rollback mustahil. Pin versi/tag SHA. |
| "Tanpa limit memori cuma container yang kena batas." | Tanpa batas, container makan RAM host → **host** yang OOM, semua container mati. |
| "`replicas` di compose = scaling." | Itu fitur Swarm. Di Compose biasa diabaikan; yang jalan cuma `deploy.resources.limits`. |
| "Redis mati = API error." | Salah kelola, API malah **hang** nunggu Redis. Itu bug, bukan fitur. |
