# Deployment & Rollback

## Prinsip

1. **Satu artefak, banyak environment.** Image dibangun sekali, diberi tag commit SHA, tidak pernah diubah.
2. **Rollback itu fitur, bukan rencana cadangan.** Kalau belum pernah diuji, berarti belum punya rollback.
3. **Tidak ada langkah manual yang wajib.** Kalau harus ingat sesuatu, itu bug.

## Build

```bash
make build            # -> sre-lab-api:sha-<12 char>
# di CI (push ke main): otomatis, plus attestation
gh attestation verify oci://ghcr.io/<user>/.../sre-api:sha-abc123 -R <user>/<repo>
```

## Deploy versi tertentu

Selalu dengan **digest**, bukan tag mutable:

```bash
docker pull ghcr.io/<user>/.../sre-api@sha256:<digest>
# lalu pakai digest itu di compose / manifest k8s
kubectl set image deployment/sre-api api=ghcr.io/<user>/.../sre-api@sha256:<digest>
kubectl rollout status deployment/sre-api --timeout=5m
```

## Checklist sebelum deploy

- [ ] CI hijau (unit + integration)
- [ ] Perubahan konfigurasi/secret sudah ada di environment target
- [ ] Ada rencana rollback: **versi sebelumnya diketahui dan tersedia**
- [ ] Tidak dalam masa maintenance window sistem lain yang jadi dependency
- [ ] Alerting sehat (bukan dalam kondisi banyak alert aktif)

## Checklist setelah deploy

- [ ] `/healthz` dan `/readyz` 200
- [ ] Error rate dalam 10 menit pertama tidak lebih buruk dari baseline
- [ ] p95 latency dalam rentang normal
- [ ] Tidak ada restart container yang tidak wajar (`docker compose ps`, `kubectl get pods`)
- [ ] Log tidak muncul error baru yang belum dikenal

## Kriteria rollback — jangan debat, lihat angka

| Gejala | Ambang | Tindakan |
| --- | --- | --- |
| Error rate | > 2% selama 5 menit | Rollback |
| p95 latency | > 2× baseline selama 10 menit | Rollback |
| CrashLoop / restart beruntun | ≥ 3 kali dalam 5 menit | Rollback |
| Readiness gagal terus | > 2 menit | Rollback |

## Prosedur rollback

```bash
# Kubernetes
kubectl rollout undo deployment/sre-api
kubectl rollout status deployment/sre-api

# Compose (keterbatasan: tidak ada perintah rollout; ini recreate ke image sebelumnya)
docker compose down
# ganti tag image ke versi sebelumnya
docker compose up -d --wait
```

**Perkiraan waktu:** K8s ~30–60 detik. Compose: ada jeda, tidak zero-downtime — keterbatasan ini dicatat eksplisit di `labs/week-06-deployment/`.

## Verifikasi setelah rollback

- [ ] Error rate kembali ke baseline
- [ ] `/readyz` 200
- [ ] Tidak ada request yang hilang (cek log proxy untuk 5xx selama jendela rollback)
- [ ] Catat: berapa lama dari keputusan rollback sampai sistem normal
