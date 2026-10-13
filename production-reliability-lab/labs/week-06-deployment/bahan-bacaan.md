# Minggu 06 · Secrets, konfigurasi, dan deployment — Bahan Bacaan

> Dokumen ini berdiri sendiri. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham kenapa deployment itu soal kepercayaan, bukan cuma "jalankan perintah". Baru setelah itu kerjakan lab.md.

## 1. Kenapa minggu ini penting

Bayangin jam 2 pagi. Pager nyala: aplikasi di produksi ngerespons 500 buat semua request. Lo cek log, ternyata koneksi ke Redis putus tiap 30 detik, dan tiap kali itu terjadi, Kubernetes ngerasa prosesnya "mati" lalu merestart semua pod. Sekarang lo punya: Redis lagi flaky + semua pod restart barengan + connection storm pas Redis balik. Satu detik gangguan jadi outage 5 menit. Ini bukan skenario langka — ini salah konfigurasi yang paling sering bikin insiden: **liveness probe yang ngecek dependency.**

Atau yang lebih senyap: lo deploy versi baru, tapi lupa kalau password database disimpen plaintext di environment variable. Besoknya ada orang screenshot `docker inspect` dan password-nya kebocoran di Slack. Atau lo deploy pakai `:latest`, rollback-nya "pakai versi sebelumnya" tapi nggak ingat versi sebelumnya itu apa. Semua ini satu akar: **deployment dianggap sekadar "menjalankan perintah", padahal dia soal kepercayaan** — kepercayaan bahwa versi yang jalan persis seperti yang lo uji, bahwa proses berhenti pelan-pelan nggak nebang request di tengah jalan, dan bahwa secret nggak pernah kelihatan di tempat yang salah.

Analogi yang pas: deployment itu kayak pindah rumah. Barangnya sama, tapi cara lo bungkus (konfigurasi), cara lo angkat (rolling vs sekaligus), dan di mana lo naruh surat berharga (secret) menentukan apakah pas sampai tujuan semuanya utuh. Kalau lo cuma "pindahin tanpa rencana", yang hilang biasanya bukan kursi — tapi dokumen penting.

## 2. Konsep inti

### 2.1 Konfigurasi itu bukan rahasia, rahasia itu bukan konfigurasi

Dua hal sering dicampur: **config** (port, nama database, batas koneksi — aman diliat orang) dan **secret** (password, API key, token — nggak boleh bocor). Aturan 12-factor: config lewat environment variable, tapi secret jangan ditaruh sebagai env var biasa yang gampang bocor ke `docker inspect`, crash report, atau log.

Solusinya di lab ini: **SOPS + age**. File `.env` terenkripsi, cuma bisa dibuka pakai private key `age` lo. Di CI, private key disuntikkan sebagai secret, didekripsi di memori, nggak pernah ditulis ke disk sebagai plaintext.

```bash
# enkripsi file env (age key)
sops --encrypt --age <AGE_PUBLIC_KEY> .env.example > .env.enc
# dekripsi di CI / lokal (butuh AGE_SECRET_KEY)
sops --decrypt .env.enc > .env
```

### 2.2 Tiga probe: startup, readiness, liveness

Ini sumber insiden paling klasik kalau disalahpahami.

| Probe | Pertanyaan | Kalau gagal | Boleh ngecek dependency? |
| --- | --- | --- | --- |
| **startupProbe** | "udah selesai booting?" | tunggu, jangan restart | ya (wajar) |
| **readinessProbe** | "boleh terima traffic?" | dicabut dari load balancer | **ya, ini tempatnya** |
| **livenessProbe** | "proses masih waras?" | restart | **jangan** — cuma cek proses responsif |

Liveness yang ngecek DB = mesin pembuat outage (lihat cerita awal). Di `app/main.go`, `/healthz` cuma bilang "proses hidup", sedangkan `/readyz` yang ngecek Postgres + Redis. Itu sebabnya Dockerfile pakai `/healthz` buat `HEALTHCHECK`, bukan `/readyz`.

### 2.3 Graceful shutdown di Go

Proses nggak boleh "dibunuh tengah makan". Di Go, kita pakai `signal.NotifyContext` buat nangkep `SIGTERM`, lalu `srv.Shutdown(ctx)` yang menyelesaikan request yang lagi jalan (in-flight) sebelum nutup.

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
go func() { _ = srv.ListenAndServe() }()
<-ctx.Done()                         // dapet SIGTERM
c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()
_ = srv.Shutdown(c)                  // selesaiin request in-flight
db.Close(); redis.Close()            // tutup pool rapi
```

Aturan emas: **timeout shutdown HARUS lebih kecil dari `terminationGracePeriodSeconds`** di Kubernetes, atau orchestrator memaksa kill di tengah jalan. Di Compose, `stop_grace_period` peranannya sama.

### 2.4 PID 1 dan exec form

Di container, kalau PID 1-nya shell (`/bin/sh -c "..."`), sinyal `SIGTERM` kadang nggak nyampe ke binary lo. Pakai **exec form** di Dockerfile (`["/app/sre-api"]`, bukan `/app/sre-api`), atau `init: true` di Compose supaya ada init yang neruskans sinyal. Tanpa ini, graceful shutdown lo cuma harapan.

### 2.5 Rolling deployment & batas Docker Compose

Rolling = ganti versi satu per satu, versi lama tetap jalan sampai yang baru sehat. Kubernetes punya `kubectl rollout` + `maxSurge`/`maxUnavailable`. **Docker Compose nggak punya perintah rollout** — `update_config` itu fitur Swarm, bukan Compose v2. Jadi di lab ini, "zero-downtime" di Compose terbatas: `docker compose up -d` ganti container, tapi ada jendela kosong. Catat keterbatasan ini jujur, bukan pura-pura.

```bash
kubectl set image deployment/sre-api api=ghcr.io/<owner>/<repo>/sre-api@sha256:<digest>
kubectl rollout status deployment/sre-api --timeout=5m
# Compose (keterbatasan dicatat):
docker compose down && docker compose up -d --wait   # ada jeda, bukan zero-downtime
```

### 2.6 Kapan rollback, jangan debat

Rollback itu fitur, bukan rencana cadangan. Kalau belum diuji, berarti belum punya rollback. Kriteria objektif (dari `docs/deployment.md`):

| Gejala | Ambang | Tindakan |
| --- | --- | --- |
| Error rate | > 2% selama 5 menit | Rollback |
| p95 latency | > 2× baseline 10 menit | Rollback |
| CrashLoop | ≥ 3× dalam 5 menit | Rollback |

### 2.7 Satu artefak, banyak environment

Image dibangun sekali (minggu 5), diberi tag digest, nggak pernah diubah. Yang beda antar environment cuma konfigurasi + secret. Jangan rebuild di tiap environment — lo nggak akan nguji barang yang sama.

## 3. Mekanisme di balik layar

Dari `docker compose up` sampai traffic lewat, ini yang terjadi:

1. **Compose baca `docker/compose.yaml`** → validasi, resolve image by digest/tag, buat network & volume kalau belum ada.
2. **Container mulai booting** → `startupProbe` (kalau ada) nunggu sampai aplikasi selesai init; selama ini, pod dianggap belum siap.
3. **`/healthz` mulai 200** → liveness bilang proses waras. Tapi ini BELUM berarti bisa terima traffic.
4. **`/readyz` diuji** → cek `db.PingContext` + `redis.Ping`. Kalau keduanya OK, container masuk ke load balancer (readiness = true).
5. **Traffic masuk** → request dilayani. Di tengahnya, orchestrator kirim **`SIGTERM`** (mis. `docker compose stop`).
6. **`signal.NotifyContext` kebakar** → `srv.Shutdown(c)` jalan: stop terima koneksi BARU, selesaiin request lama dalam 20 detik.
7. **Pool ditutup rapi** (`db.Close()`, `redis.Close()`) → nggak ada koneksi gantung.
8. **Proses exit 0** → orchestrator lepas container. Kalau sampai `terminationGracePeriod` habis tapi proses belum exit, baru dipaksa `SIGKILL` (ini yang bikin 5xx).

## 4. Glosarium mini

| Istilah | Arti sederhana | Contoh |
| --- | --- | --- |
| **Config** | Pengaturan yang aman diliat orang | `APP_PORT`, nama DB |
| **Secret** | Kredensial yang nggak boleh bocor | password DB, API key |
| **SOPS** | Alat enkripsi file config/secret | `sops --encrypt .env` |
| **age** | Algoritma kunci publik buat enkripsi | `AGE_PUBLIC_KEY` / `AGE_SECRET_KEY` |
| **Liveness** | "proses masih waras?" | `/healthz` → 200 |
| **Readiness** | "boleh terima traffic?" | `/readyz` → cek DB+Redis |
| **Startup probe** | "udah selesai boot?" | tunggu init selesai |
| **Graceful shutdown** | berhenti pelan, selesaiin request | `srv.Shutdown(ctx)` |
| **SIGTERM** | sinyal "tolong berhenti pelan" | `docker compose stop` |
| **Rolling update** | ganti versi bertahap | `kubectl rollout` |
| **Rollback** | balik ke versi sebelumnya | `kubectl rollout undo` |
| **PID 1** | proses pertama di container | shell vs binary lo |

## 5. Bahan bacaan wajib

| # | Sumber | Yang dicari |
| --- | --- | --- |
| 1 | [Container probes — Kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#container-probes) | beda liveness/readiness/startup + efek `restartPolicy` |
| 2 | [Configure liveness/readiness/startup probes — Kubernetes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/) | contoh YAML tiap probe |
| 3 | [`signal.NotifyContext` — Go pkg.go.dev](https://pkg.go.dev/os/signal#NotifyContext) | gimana context kebakar pas dapat sinyal |
| 4 | [The Twelve-Factor App: Config](https://12factor.net/config) | kenapa config lewat env, secret dipisah |
| 5 | [SOPS — GitHub](https://github.com/getsops/sops) | enkripsi file secret dengan age/KMS/PGP |
| 6 | [Rolling updates & rollbacks — Kubernetes](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment) | `maxSurge`, `maxUnavailable`, `rollout undo` |

## 6. Self-check

1. Kenapa liveness probe yang ngecek database itu berbahaya, dan probe mana yang seharusnya ngecek dependency?
2. Apa beda `/healthz` dan `/readyz` di `app/main.go`, dan kenapa Dockerfile pakai `/healthz` buat HEALTHCHECK?
3. Di Go, apa yang terjadi kalau timeout `srv.Shutdown` lebih besar dari `terminationGracePeriodSeconds`?
4. Kenapa PID 1 di container sebaiknya binary lo (exec form), bukan shell?
5. Docker Compose punya perintah `rollout` seperti Kubernetes? Kalau nggak, apa batasnya?
6. Kapan lo harus rollback menurut tabel di `docs/deployment.md`?

## 7. Yang sering salah paham

| Mitos | Kenyataan |
| --- | --- |
| "Liveness cek DB biar tau kalau DB mati." | Liveness yang ngecek DB bikin restart massal saat DB flaky → outage makin parah. Pakai readiness. |
| "Secret di env var aman, kan cuma di dalam container." | Gampang bocor ke `docker inspect`, log error, crash report. Pakai SOPS/secret manager. |
| "Compose bisa zero-downtime kalau pakai `up -d`." | Compose v2 nggak punya rollout; ada jeda. `update_config` itu Swarm. Catat keterbatasan. |
| "Rollback cuma ganti tag balik." | Rollback harus nunjuk **digest** versi sebelumnya, bukan `latest` yang udah berubah. |
| "SIGTERM otomatis graceful." | Cuma kalau kode lo nangkep sinyal & shutdown pelan. Kalau PID 1 shell, sinyal nggak nyampe. |
