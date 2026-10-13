# Service Down — endpoint tidak bisa diakses

## Gejala
- Alert `APIDown` menyala, atau pengguna melapor "tidak bisa akses"
- `curl` mengembalikan: `Connection refused`, hang lalu timeout, atau 5xx

## Dampak
Semua pengguna tidak bisa memakai layanan. Anggap SEV1/SEV2 sampai terbukti sebaliknya.

## Deteksi — jalankan berurutan, jangan lompat

```bash
# 1. Apakah prosesnya hidup?
pgrep -af sre-api            # kosong = proses mati
docker compose ps            # lihat kolom Status dan Health

# 2. Apakah ada yang listen?
ss -lntp | grep 8080         # kosong = tidak ada yang listen

# 3. Apakah aplikasi menjawab?
curl -v --max-time 3 http://localhost:8080/healthz
echo "exit: $?"
#   7  = connection refused  -> lanjut ke cabang B
#   28 = timeout             -> lanjut ke cabang C
#   0  = aplikasi menjawab   -> lanjut ke cabang A

# 4. Apakah aplikasi SADAR bahwa dirinya tidak sehat?
curl -v --max-time 3 http://localhost:8080/readyz

# 5. Apa kata log terakhir?
docker compose logs --tail=100 api
journalctl -u sre-lab.service -n 100 --no-pager   # kalau pakai systemd
```

## Diagnosis

```
Proses tidak ada (pgrep kosong)
├─ Container exit?        → docker inspect --format '{{.State.ExitCode}}' <ctr>
│                           exit 137 = OOM/kill → lari ke runbook oom-symptoms.md
│                           exit 1   = panic/error startup → baca log
└─ Tidak pernah jalan?    → cek env/secret, cek dependency

Proses ada, tapi tidak ada yang listen (ss kosong)
└─ Aplikasi gagal binding → port bentrok? cek `lsof -i :8080`
                            env APP_PORT salah? cek .env

Proses ada, listen ada, tapi tidak menjawab (timeout)
├─ /healthz juga timeout   → aplikasi hang: thread pool/koneksi DB habis
│                            cek: in-flight requests, koneksi DB
└─ /healthz OK, /readyz 503 → dependency bermasalah → cek Postgres/Redis
```

## Mitigasi

1. **Kalau baru saja deploy → rollback dulu, analisis belakangan.**
   ```bash
   kubectl rollout undo deployment/sre-api      # K8s
   # Compose: turunkan, ganti tag image ke versi sebelumnya, naikkan lagi
   ```
2. Kalau OOM: naikkan limit memori **sementara**, lalu cari penyebabnya.
3. Kalau dependency mati: hidupkan dependency. Jangan restart aplikasi dulu —
   restart massal saat DB down justru memperpanjang insiden.
4. Kalau hang karena saturasi: kurangi beban (scale out atau batasi concurrency).

## Verifikasi pemulihan

- [ ] `curl /healthz` → 200
- [ ] `curl /readyz` → 200
- [ ] Error rate kembali ke baseline (lihat dashboard)
- [ ] Tidak ada restart container dalam 10 menit terakhir
- [ ] p95 latency kembali normal

## Pencegahan

- Pasang alert untuk sinyal leading: saturasi koneksi DB, in-flight request tinggi
- Pisahkan `/healthz` dan `/readyz` (lihat minggu 6)
- Pastikan `startupProbe` ada, supaya aplikasi lambat booting tidak dinyatakan mati

## Eskalasi

Kalau belum pulih dalam **30 menit**, atau dampak meluas ke sistem lain: naikkan severity dan minta bantuan pemilik dependency.
