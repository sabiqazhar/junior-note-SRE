# Ringkasan — Minggu 06 · Secrets, konfigurasi, dan deployment

> Satu halaman buat dibaca ulang sebelum ujian/wawancara. Bukan pengganti bahan-bacaan.md.

## Inti minggu ini dalam 5 kalimat

1. Config dan secret itu dua hal beda: config aman diliat, secret nggak — jangan taruh secret sebagai env var biasa yang gampang bocor ke `docker inspect`/log.
2. Tiga probe punya tugas beda: startup ("udah boot?"), readiness ("boleh terima traffic?"), liveness ("proses waras?"); **liveness jangan ngecek dependency** atau lo bikin outage berantai.
3. Graceful shutdown di Go = tangkap `SIGTERM` via `signal.NotifyContext`, selesaiin request in-flight, lalu tutup pool — dengan syarat timeout shutdown < grace period orchestrator.
4. Docker Compose **nggak punya** perintah `rollout`; zero-downtime di situ terbatas — catat keterbatasan ini jujur, jangan pura-pura.
5. Rollback itu fitur yang harus diuji: nunjuk **digest** versi sebelumnya, bukan `:latest`, dan cuma jalan kalau ambang error/latency terlampaui.

## Peta konsep (daftar bertingkat)

- **Konfigurasi & secret**
  - Config: port, nama DB, limit — aman diliat
  - Secret: password, API key — jangan bocor
  - 12-factor: config lewat env, secret dipisah & dienkripsi
  - SOPS + age: file `.env` terenkripsi, dekripsi di memori (CI/lama)
- **Probe**
  - startupProbe: tunggu boot selesai
  - readinessProbe → `/readyz`: cek DB+Redis, cabut dari LB kalau gagal
  - livenessProbe → `/healthz`: cuma cek proses, restart kalau gagal
  - Di repo: Dockerfile `HEALTHCHECK` pakai `/healthz` (bukan `/readyz`)
- **Graceful shutdown**
  - `signal.NotifyContext(SIGINT, SIGTERM)`
  - `srv.Shutdown(ctx)` — stop terima koneksi baru, selesaiin in-flight
  - `db.Close()` / `redis.Close()` — tutup pool rapi
  - Syarat: `shutdownTimeout < terminationGracePeriod`
  - Exec form / `init: true` biar sinyal nyampe ke PID 1
- **Deployment & rollback**
  - Rolling update (k8s: `rollout status`, `maxSurge`)
  - Compose: `up -d` — ada jeda, bukan zero-downtime
  - Rollback: `kubectl rollout undo` / tarik digest lama
  - Kriteria objektif: error >2% 5m, p95 >2× baseline 10m, CrashLoop ≥3/5m

## Perintah kunci + yang harus lo lihat di output

| Perintah | Yang dicari di output | Artinya kalau beda |
| --- | --- | --- |
| `sops --encrypt --age $K .env.example > .env.enc` | file terenkripsi, bukan plaintext | kalau masih plaintext → salah flag `--age` |
| `sops --decrypt .env.enc` | isi sama dengan aslinya | beda → key salah |
| `curl -w "%{http_code}" /health` | `200` | bukan 200 → proses nggak jalan |
| `curl -w "%{http_code}" /readyz` | `200` saat normal, `503` saat DB/Redis mati | `200` pas dependency mati → salah taruh cek di `/healthz` |
| `docker compose stop api` | log `shutdown selesai`, ~0 5xx | banyak 5xx → timeout shutdown > grace period |
| `grep -n 'ENTRYPOINT' docker/Dockerfile` | `["/app/sre-api"]` (exec form) | shell form → sinyal nggak nyampe |
| `docker inspect --format='{{index .RepoDigests 0}}' <svc>` | `sha256:...` | `nil` → image lokal belum di-push |
| `kubectl rollout status deployment/sre-api` | `successfully rolled out` | timeout → deploy gagal |
| `kubectl rollout undo deployment/sre-api` | pod baru dari ReplicaSet lama | error → ReplicaSet lama sudah dihapus |

## Failure matrix / decision tree

| Gejala | Kemungkinan penyebab | Cara buktikan | Mitigasi |
| --- | --- | --- | --- |
| Outage meluas saat DB flaky | liveness ngecek DB → restart massal | cek apakah `/healthz` ikut gagal saat DB mati | pindahkan cek dependency ke `/readyz` |
| 5xx saat SIGTERM | timeout shutdown > grace period | bandingkan `shutdownTimeout` vs `terminationGracePeriod` | kecilkan `shutdownTimeout` |
| Sinyal nggak memicu shutdown | PID 1 shell, bukan binary | `grep ENTRYPOINT Dockerfile` | pakai exec form / `init: true` |
| Secret bocor ke log | env var biasa / `echo $SECRET` | `docker inspect`, grep log | pakai SOPS/secret manager |
| `readyz=200` padahal DB mati | handler nggak ngecek DB beneran | matiin DB, cek `readyz` | fix `handleReadyz` di `app/main.go` |
| Rollback gagal: digest hilang | retensi image habis | `docker pull ...@sha256:lama` → `manifest unknown` | atur retensi registry |
| Downtime saat `up -d` | Compose ganti container sekaligus | `docker compose ps` saat ganti | catat keterbatasan; pakai k8s buat zero-downtime |

## Jebakan umum

- **Liveness ngecek dependency** → mesin pembuat cascade failure. Readiness tempatnya.
- **Secret di env var biasa** → bocor ke `docker inspect`, crash report, log.
- **PID 1 shell** → `SIGTERM` nggak nyampe, shutdown nggak graceful.
- **Timeout shutdown > grace period** → proses dipaksa mati, request putus.
- **`latest` buat rollback** → tag mutable, nggak tahu persis versi yang kemarin.
- **Compose dianggap zero-downtime** → `update_config` itu Swarm, Compose v2 nggak punya rollout.
- **Rollback cuma di kepala** → kalau belum diuji, berarti belum punya rollback.

## Acceptance criteria

1. Aplikasi berhenti graceful saat SIGTERM; request in-flight selesai, ~0 5xx.
2. `/readyz` 503 saat dependency wajib mati; `/healthz` tetap 200.
3. Startup gagal (password salah) → container nggak pernah healthy.
4. Nggak ada secret plaintext di Git; `.env.enc` terenkripsi (SOPS+age).
5. Rollback by digest diuji, atau retensi image dicatat sebagai temuan.
6. `docs/deployment.md` punya kriteria rollback objektif.

## Satu kalimat yang harus lo ingat

> Deployment yang bisa diandalkan itu soal kepercayaan: lo tahu persis versi apa yang jalan (digest), gimana dia berhenti (graceful), dan gimana balik kalau rusak (rollback by digest) — bukan sekadar "udah dijalankan perintahnya".
