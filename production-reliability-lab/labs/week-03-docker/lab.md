# Lab — Minggu 03 · Docker production fundamentals

> Target: stack Go API + PostgreSQL 18 + Redis jalan dari satu perintah, dengan resource limit, healthcheck, dan degradasi terkendali.

## Persiapan (10–15 menit)

1. **Cek Docker Compose v2** → perintah: `docker compose version` → Yang harus terjadi: `Docker Compose version v2.x.x` (ada spasi, bukan `docker-compose`). → Kalau gagal: `docker compose` nggak ada = install Docker Engine terbaru. → Catat di notes.md.
2. **Cek file lab ada** → perintah: `ls docker/compose.yaml docker/Dockerfile .env.example` → Yang harus terjadi: ketiganya ada. → Kalau gagal: ambil dari repo utama. → Catat di notes.md.
3. **Siapin .env** → perintah: `cp .env.example .env && grep -E 'PG|REDIS' .env` → Yang harus terjadi: `.env` berisi `PGUSER`, `PGPASSWORD`, `PGDATABASE`, `REDIS_URL`. → Kalau gagal: edit manual. → Catat di notes.md (jangan commit .env!).

## Hari 1 · Teori (2 jam)

1. **Baca bahan bacaan** → perintah: `cat labs/week-03-docker/bahan-bacaan.md` → Yang harus terjadi: lo paham depends_on vs healthcheck vs limits. → Kalau gagal: ulang bagian #2. → Catat di notes.md.
2. **Baca compose reference** → perintah: buka [Compose services](https://docs.docker.com/reference/compose-file/services/) → Yang harus terjadi: lo tau `condition: service_healthy` & `deploy.resources.limits`. → Kalau gagal: fokus ke contoh di lab.md. → Catat di notes.md.
3. **Pahami multi-stage Dockerfile** → perintah: `cat docker/Dockerfile` → Yang harus terjadi: lo lihat `FROM golang AS build` lalu `FROM gcr.io/distroless/static`. → Kalau gagal: lihat Dockerfile reference. → Catat di notes.md.
4. **Bedah compose.yaml** → perintah: `docker compose -f docker/compose.yaml config` → Yang harus terjadi: YAML valid, service `api/postgres/redis` muncul. → Kalau gagal: baca error (biasanya indent YAML). → Catat di notes.md.
5. **Catat versi dependency** → perintah: `grep -E 'image:' docker/compose.yaml` → Yang harus terjadi: `postgres:18`, `redis:7-alpine`, `nginx:1.29-alpine`. → Kalau gagal: sesuaikan ke versi di jadwal. → Catat di notes.md.

## Hari 2 · Praktik (2 jam)

1. **Build & jalanin dari bersih** → perintah: `docker compose up -d --build --wait` → Yang harus terjadi: `Running 3/3` dan semua `healthy`. → Kalau gagal: `docker compose logs` cari panic. → Catat di notes.md.
2. **Cek status + health** → perintah: `docker compose ps --format 'table {{.Name}}\t{{.Status}}\t{{.Health}}'` → Yang harus terjadi: `api Up ... healthy`, `postgres ... healthy`, `redis ... healthy`. → Kalau gagal: lihat healthcheck tiap service. → Catat di notes.md.
3. **Validasi config** → perintah: `docker compose config | head -40` → Yang harus terjadi: limit & healthcheck kelihatan di output. → Kalau gagal: compose.yaml salah. → Catat di notes.md.
4. **Lihat resource nyata** → perintah: `docker stats --no-stream` → Yang harus terjadi: CPU/MEM/PIDS per container sesuai limit. → Kalau gagal: limit belum dipasang. → Catat di notes.md.
5. **Inspect network** → perintah: `docker network ls` & `docker compose exec api getent hosts redis` → Yang harus terjadi: bridge stack ada; `api` resolve `redis`. → Kalau gagal: cek `networks:` di compose. → Catat di notes.md.
6. **Cek port terekspos** → perintah: `ss -lntp | grep -E '5432|6379|8080'` → Yang harus terjadi: CUMA `8080` (atau `80` nginx) yang listen di host; `5432`/`6379` nggak ada. → Kalau gagal: postgres/redis kepasang `ports:`. → Catat di notes.md.

## Hari 3 · Eksperimen kegagalan (2 jam) — WAJIB ada cara kembali ke normal

1. **Batas memori mepet → OOM cgroup** → perintah:
   ```bash
   docker run --rm --name oom-test --memory=64m --cpus=.5 --pids-limit=100 postgres:18
   docker inspect --format '{{json .State}}' oom-test 2>/dev/null
   ```
   → Yang harus terjadi: container exit karena OOM; `docker events --since 10m` nunjukin `oom`. → Kalau gagal: postgres butuh >64m; naikin ke 128m. → Catat di notes.md. **Kembali normal:** container `--rm` otomatis bersih; jalankan `docker ps -a` buat pastiin nggak nyangkut.
2. **Matiin dependency → lihat reaksi API** → perintah:
   ```bash
   docker compose stop redis
   docker compose logs -f api        # amati: panic? hang? 503?
   docker compose ps                 # health api berubah?
   ```
   → Yang harus terjadi: lo lihat API **hang** (ini yang dicari — bug!). → Kalau gagal: API langsung 503 = sudah bagus, catat sebagai "controlled degradation OK". → Catat di notes.md. **Kembali normal (WAJIB):** `docker compose start redis && docker compose ps` → tunggu `redis` sehat sebelum lanjut.
3. **Container restart loop** → perintah: `docker compose restart api && docker compose ps --format 'table {{.Name}}\t{{.Status}}'` → Yang harus terjadi: kalau app panic, status `Restarting`. → Kalau gagal: app sehat, status `Up`. → Catat di notes.md. **Kembali normal:** pastiin app bener (perbaiki env) lalu `docker compose up -d`.
4. **Cari kenapa restart terus** → perintah: `docker inspect <container> --format '{{json .State}}' | jq` & `docker events --since 10m` → Yang harus terjadi: `ExitCode` & `Error` ketemu; event `die`/`oom` muncul. → Kalau gagal: jq belum ada, pakai `python3 -m json.tool`. → Catat di notes.md.
5. **Verifikasi degradasi terkendali** → perintah: `curl -sS -m 2 -o /dev/null -w '%{http_code} time=%{time_total}\n' http://localhost:8080/health` (waktu Redis mati) → Yang harus terjadi: 503 dalam <100ms (bukan hang 30s). → Kalau gagal: ini bukti bug; catat buat diperbaiki minggu 6. → Catat di notes.md.

## Hari 4 · Review & dokumentasi (2 jam)

1. **Healthcheck beneran** → perintah: buka `docker/compose.yaml`, cek `test:` tiap service nggak `true`. → Yang harus terjadi: PG = `pg_isready`, Redis = `redis-cli ping`. → Kalau gagal: ganti ke tes nyata. → Catat di notes.md.
2. **Tabel perbandingan mekanisme** → perintah: salin tabel "depends_on vs healthcheck vs app retry" ke `commands.md`. → Yang harus terjadi: tabel lengkap. → Kalau gagal: lihat ringkasan. → Catat di notes.md.
3. **Cek .env / .gitignore** → perintah: `grep -n 'POSTGRES_PASSWORD' docker/compose.yaml; grep -n '.env' .gitignore` → Yang harus terjadi: password di `.env` (lewat `env_file`), dan `.env` di `.gitignore`. → Kalau gagal: jangan hardcode password. → Catat di notes.md.
4. **Lengkapi commands.md** → perintah: isi Hari 2 & Hari 3 dengan output asli (≥15 perintah). → Yang harus terjadi: semua percobaan tercatat. → Kalau gagal: tambah eksperimen. → Catat di notes.md.
5. **Update README** → perintah: isi `## Referensi`. → Yang harus terjadi: 5–8 link primer. → Kalau gagal: lihat bahan bacaan #5. → Catat di notes.md.
6. **Commit** → perintah: `git add labs/week-03-docker docker .env.example && git commit -m "week 03 docker lab"` → Yang harus terjadi: commit masuk, `.env` tdk ikut. → Kalau gagal: cek `git status`. → Catat di notes.md.

## Daftar periksa "minggu ini selesai"

- [ ] `docker compose up -d --build --wait` jalan dari bersih, 3 service healthy.
- [ ] `ss -lntp` nunjukin 5432/6379 TIDAK terekspos ke host.
- [ ] Tiap container punya limit CPU/memori/pids + healthcheck nyata.
- [ ] Lo bisa reproduksi: Redis mati → API hang/503, lalu `start redis` balikin normal.
- [ ] `commands.md` ≥15 perintah + output; `notes.md` ada tabel & kesimpulan.
- [ ] State bersih: nggak ada container `--rm` nyangkut, Redis sudah `start`.

## Cara membuktikan ke reviewer

```bash
# 1. Start bersih satu perintah
docker compose up -d --build --wait
docker compose ps --format 'table {{.Name}}\t{{.Status}}\t{{.Health}}'

# 2. Bukti port internal doang
sudo ss -lntp | grep -E '5432|6379' && echo "BAHAYA: terekspos" || echo "aman: PG/Redis internal"

# 3. Bukti degradasi (matiin redis, amati)
docker compose stop redis
curl -sS -m 2 -o /dev/null -w 'code=%{http_code} time=%{time_total}\n' http://localhost:8080/health
docker compose start redis && docker compose ps

# 4. Bukti resource limit
docker stats --no-stream

# 5. State bersih
docker ps -a --filter status=exited
```
