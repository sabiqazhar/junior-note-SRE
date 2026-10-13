# Production Readiness Checklist

Dokumen ini diisi **minggu 8**, lalu direvisi di minggu 16.
Pertanyaan yang dijawab PRR: kalau sistem ini rusak jam 3 pagi, apakah orang yang jaga bisa mengerti dan memperbaikinya tanpa bertanya?

Cara isi: `TODO` (belum), `PARTIAL` (sebagian), `DONE` (teruji, ada bukti).
Jangan tulis `DONE` kalau belum pernah diuji.

## 1. Security

| Item | Status | Bukti |
| --- | --- | --- |
| Tidak ada secret di repo/image | | hasil `gitleaks detect` |
| Container jalan non-root | | `docker inspect` / manifest k8s |
| Image di-scan, tidak ada HIGH/CRITICAL | | hasil `trivy image` |
| Port database & Redis tidak terekspos | | `ss -lntp`, `nmap` dari luar |
| TLS aktif + auto-renewal teruji | | `openssl s_client`, log renewal |
| Security group tidak ada `0.0.0.0/0` ke port internal | | `terraform plan` / console |
| IMDSv2 enforced (cloud) | | `metadata_options.http_tokens = required` |

## 2. Availability

| Item | Status | Bukti |
| --- | --- | --- |
| > 1 instance (atau SPOF dinyatakan & diterima) | | `docker compose ps` / `kubectl get pods` |
| Healthcheck + readiness terpisah | | `labs/week-06-deployment/` |
| Graceful shutdown teruji (0 5xx saat restart) | | hasil `hey` saat SIGTERM |
| Rollback teruji | | `docs/deployment.md` |
| PDB + HPA (kalau K8s) | | `kubernetes/deployment.yaml` |

## 3. Observability

| Item | Status | Bukti |
| --- | --- | --- |
| Metrics RED tersedia | | `labs/week-09-prometheus/` |
| Log terstruktur + correlation/trace ID | | `labs/week-11-observability/` |
| Trace menyambung ke dependency | | screenshot Tempo |
| Alert "service down" terpicu < 2 menit | | `labs/week-10-grafana/` |
| Dashboard bisa di-reprovision dari Git | | `observability/grafana/provisioning/` |
| Setiap alert punya runbook | | `runbooks/` |

## 4. Recovery

| Item | Status | Bukti |
| --- | --- | --- |
| Backup otomatis jalan | | `pgbackrest info` |
| Restore pernah diuji ke instance terpisah | | `labs/week-07-backup-recovery/` |
| RPO/RTO tertulis dan diukur | | `docs/disaster-recovery.md` |
| Runbook untuk 3 skenario kegagalan terbesar | | `runbooks/` |
| Sistem pulih sendiri setelah reboot | | hasil reboot minggu 8 |

## 5. Deployment

| Item | Status | Bukti |
| --- | --- | --- |
| Pipeline otomatis (test + build) | | `.github/workflows/ci.yml` |
| Artefak immutable ber-tag commit SHA | | daftar tag di registry |
| Tidak ada langkah manual wajib | | `README.md` → "cara jalanin dari nol" |
| Rollback teruji | | `docs/deployment.md` |
| Alerting tidak dalam kondisi berisik | | daftar alert aktif 30 hari |

## Dependency map

| Komponen | Failure mode | Dampak | Deteksi | Mitigasi | SPOF? |
| --- | --- | --- | --- | --- | --- |
| PostgreSQL | Mati / lambat | Semua endpoint tulis gagal | readiness + latency query | Pool + timeout + backup + PITR | **Ya** |
| Redis | Mati | Cache miss → DB kebanjiran | hit ratio metric | Fallback: anggap cache kosong | Ya, ada fallback |
| Nginx | Mati | Semua akses gagal | healthcheck eksternal | — | **Ya** |
| Object storage | Mati | Backup gagal | alert job backup | Alert + retry | Tidak langsung |
| Provider auth (eksternal) | Timeout | Login gagal | latency outbound | Timeout + circuit breaker | Ya |

## Risiko yang diterima secara eksplisit

Tulis di sini, jangan disembunyikan:

1. Single instance database — belum ada replika/standby. RTO saat DB rusak total = waktu restore penuh.
2. Tidak ada multi-AZ; satu kegagalan zona = layanan tidak tersedia.
3. Kapasitas belum diuji di atas titik saturasi produksi (baru level lab).

Ditulis: ____ · Diperbarui: ____
