# Perintah — Week 03 · Docker production fundamentals

Format: `perintah` → output yang diharap → interpretasi. Output disanitasi. Jalankan di Ubuntu 22.04/24.04 + Docker Compose v2. Ganti `<c>` dengan nama/ID container asli (`docker compose ps`).

---

## Hari 2 (kondisi normal)

**1. Versi Compose (v2, bukan v1)**
```bash
docker compose version
```
→ `Docker Compose version v2.29.1`
→ Ada spasi = v2 (Go). `docker-compose` (strip) sudah dihapus.

**2. Validasi config sebelum jalan**
```bash
docker compose config
```
→ YAML tervalidasi, service `api/postgres/redis` muncul, nggak ada error.
→ Mencegah typo YAML bikin stack gagal start.

**3. Start dari bersih + tunggu healthcheck**
```bash
docker compose up -d --build --wait
```
→ `Container api ... Started` + `Running 3/3`.
→ `--wait` nunggu sampai SEMUA healthcheck OK, bukan cuma container up.

**4. Status + health tiap service**
```bash
docker compose ps --format 'table {{.Name}}\t{{.Status}}\t{{.Health}}'
```
→ `lab-api-1 Up 2m healthy` / `lab-postgres-1 Up 2m healthy` / `lab-redis-1 Up 2m healthy`
→ Kolom Health = hasil healthcheck.

**5. Log aplikasi real-time**
```bash
docker compose logs -f api
```
→ baris startup Go, koneksi ke PG/Redis sukses.
→ `Ctrl-C` untuk berhenti follow.

**6. Pemakaian resource nyata**
```bash
docker stats --no-stream
```
→ `NAME CPU% MEM USAGE / LIMIT PIDS`
→ Limit terlihat di kolom MEM USAGE / LIMIT (mis. `45MiB / 256MiB`).

**7. Inspect network & DNS internal**
```bash
docker network ls
docker compose exec api getent hosts redis
```
→ bridge stack ada; `api` resolve `redis` → `172.20.0.3 redis`.
→ Bukti DNS internal Compose jalan.

**8. Hanya port proxy yang terekspos**
```bash
ss -lntp | grep -E '5432|6379|8080'
```
→ cuma `:8080` (atau `:80` nginx) yang listen di host; `5432`/`6379` nggak ada.
→ PG/Redis aman di jaringan internal.

**9. Inspect state container**
```bash
docker inspect <c> --format '{{json .State}}'
```
→ `{"Status":"running","ExitCode":0,"Health":{"Status":"healthy"}}`
→ State + exit code + health dalam satu JSON.

---

## Hari 3 (eksperimen kegagalan)

**10. OOM karena batas memori cgroup**
```bash
docker run --rm --name oom-test --memory=64m --cpus=.5 --pids-limit=100 postgres:18
docker events --since 10m | grep -i oom
```
→ container exit OOM; event `oom` muncul.
→ cgroup kena batas, bukan host. `docker run --rm` bersih otomatis.

**11. Matiin Redis → lihat reaksi API**
```bash
docker compose stop redis
docker compose logs -f api
docker compose ps
```
→ lihat API **hang** (request nunggu Redis) — ini bug yang dicari.
→ **Kembali normal:** `docker compose start redis && docker compose ps` (tunggu healthy).

**12. Uji degradasi lewat curl (Redis mati)**
```bash
curl -sS -m 2 -o /dev/null -w 'code=%{http_code} time=%{time_total}\n' http://localhost:8080/health
```
→ `code=503 time=0.002` = fail fast (bagus) / `time=2.0` hang = bug.
→ Bukti controlled degradation.

**13. Container restart loop**
```bash
docker compose restart api
docker compose ps --format 'table {{.Name}}\t{{.Status}}'
```
→ kalau app panic: `Restarting`; kalau sehat: `Up`.
→ **Kembali normal:** `docker compose up -d` setelah perbaiki env.

**14. Cari kenapa restart terus**
```bash
docker inspect <c> --format '{{json .State}}' | jq
docker events --since 10m
```
→ `ExitCode` & `Error` ketemu; event `die`/`oom` muncul.
→ Jq opsional; bisa `python3 -m json.tool`.

**15. Healthcheck beneran vs `true`**
```bash
docker compose exec postgres pg_isready -U app
docker compose exec redis redis-cli ping
```
→ `pg_isready` → `/var/run/postgresql:5432 - accepting connections`; `redis-cli ping` → `PONG`.
→ Ini tes healthcheck asli, bukan `true`.

**16. Volume tidak hilang (beda down vs down -v)**
```bash
docker compose down          # data pgdata tetap
docker volume ls | grep pgdata
```
→ volume `pgdata` masih ada setelah `down`.
→ `down -v` (JANGAN di lab) yang bikin data hilang.

**17. Versi image sesuai jadwal**
```bash
docker compose images
```
→ `postgres 18`, `redis 7-alpine`, `nginx 1.29-alpine`.
→ Konfirmasi versi 2025–2026 dipakai.
