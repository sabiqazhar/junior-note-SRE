# Ringkasan — Minggu 03 · Docker production fundamentals

## Inti minggu ini dalam 5 kalimat

1. `docker compose` (v2) sudah menggantikan `docker-compose` (v1) yang sudah dihapus — jangan tulis strip `-` lagi.
2. `depends_on` bentuk pendek cuma nunggu container *start*; pakai `condition: service_healthy` buat nunggu *siap*.
3. `healthcheck` ngecek "apakah *proses ini* sehat", bukan "apakah dependency sehat" — dua hal beda.
4. Tanpa `deploy.resources.limits`, satu container bisa bikin **host** OOM, bukan cuma dirinya sendiri.
5. Inti minggu ini: kalau dependency mati, API harus **gagal cepat & jelas (503)**, bukan *hang* — itu namanya controlled degradation.

## Peta konsep (daftar bertingkat)

- Image & build
  - layer read-only + cache
  - multi-stage build (image kecil, distroless/alpine)
  - `USER` non-root
- Container lifecycle
  - created → running → stopped → removed
  - restart policy (`unless-stopped`)
  - `restarting (1)` loop = aplikasi panic saat start
- Networking
  - bridge network per stack
  - DNS internal Compose (`api` ↔ `redis`)
  - jangan kasih `ports:` ke postgres/redis
- Persistensi
  - volume vs bind mount
  - `down -v` = data hilang
- Kesiapan & dependensi
  - `healthcheck` (pg_isready, redis-cli ping)
  - `depends_on: condition: service_healthy`
- Resource & isolasi
  - cgroups v2
  - `deploy.resources.limits` (cpus, memory, pids)
- Controlled degradation
  - Redis mati → 503 / degrade, BUKAN hang
- Compose vs Swarm
  - `replicas`/`update_config` cuma Swarm
  - `deploy.resources.limits` jalan di Compose

## Perintah kunci + yang harus lo lihat di output

| Perintah | Output kunci | Interpretasi |
| --- | --- | --- |
| `docker compose config` | YAML tervalidasi tanpa error | Config benar sebelum jalan |
| `docker compose up -d --build --wait` | `Running 3/3` + sehat | `--wait` tunggu healthcheck OK |
| `docker compose ps --format 'table {{.Name}}\t{{.Status}}\t{{.Health}}'` | `api Up 2m healthy` | Status + health tiap service |
| `docker compose logs -f api` | log aplikasi real-time | Lihat panic/hang saat dep mati |
| `docker inspect <c> --format '{{json .State}}'` | `"ExitCode":0,"Status":"running"` | State & exit code container |
| `docker stats --no-stream` | CPU%, MEM, PIDS per container | Pemakaian resource nyata |
| `docker network ls` | list bridge termasuk compose stack | Jaringan internal ada |
| `docker events --since 10m` | event start/stop/die | Jejak kenapa container mati |
| `docker run --rm --memory=64m --cpus=.5 --pids-limit=100 <img>` | container mati/OOM | Batas cgroup kelewat |

## Failure matrix / decision tree

```
Container restart terus (restarting)
├─ docker inspect → ExitCode != 0 & Error panic
│   → aplikasi crash saat start. Cek logs -f api
├─ healthcheck selalu starting/unhealthy
│   → tes health salah (mis. curl di distroless). Perbaiki test
├─ docker events → OOM
│   → cgroup kena limit. Naikin memory / perbaiki leak
└─ depends_on salah → API start sebelum PG siap
    → pakai condition: service_healthy

Redis mati
├─ API balas 503 cepat  → controlled degradation OK
└─ API hang / timeout   → BUG di client (no timeout). Perbaiki minggu 6
```

| Mekanisme | Yang dijawab | Yang TIDAK dijawab |
| --- | --- | --- |
| `depends_on` | "container lain sudah distart?" | "dia sudah siap terima traffic?" |
| `healthcheck` | "proses ini sehat?" | "dependency sehat?" |
| App retry/backoff | "kalau dep sebentar mati?" | — (tanggung jawab aplikasi) |

## Jebakan umum

- Healthcheck pakai `curl` di image distroless → selalu unhealthy (nggak ada curl).
- Tanpa `USER` → container jalan sebagai root.
- Tanpa limit memori → host yang OOM, bukan container.
- `:latest` + cache → rollback mustahil; pin versi.
- `down -v` = data hilang; pahami beda `down` dan `down -v`.
- `replicas` di compose diabaikan (Swarm only).

## Acceptance criteria

1. Seluruh service hidup dari kondisi bersih dengan **satu perintah** (`docker compose up -d --build --wait`).
2. API **gagal secara terkendali** (503, bukan hang/panic) saat dependency nggak tersedia.
3. Setiap container punya limit resource masuk akal + healthcheck yang benar-benar menguji kesiapan.
4. Nggak ada password hardcoded; `.env` di `.gitignore`, `.env.example` di repo.

## Satu kalimat yang harus lo ingat

> Dependency mati bukan alasan buat hang — aplikasi lo harus gagal cepat (503) atau degrade, biar caller nggak nunggu sampai timeout proxy.
