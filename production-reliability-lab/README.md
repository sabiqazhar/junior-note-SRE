# Production Reliability Lab

Satu repository berisi **production-like service + automated deployment + monitoring + alerting + simulasi insiden + dokumentasi reliability**.

Tujuan: portofolio buat melamar **SRE / Platform Engineer**. Bukan kumpulan tutorial, tapi bukti bahwa gue bisa mengoperasikan sistem.

## Kenapa satu repo

Reviewer bisa ngikutin progres dari Linux dasar sampai incident response tanpa buka 16 repo kecil.

## Cara jalanin dari nol

```bash
git clone <repo> && cd production-reliability-lab
cp .env.example .env          # isi password yang aman
make up                       # build + jalanin semua service, tunggu sampai healthy
make health                   # cek /health dan /readyz
```

## Perintah yang tersedia

```bash
make up          # build dan jalankan semuanya (tunggu healthcheck)
make down        # stop, data tetap ada
make reset       # stop + hapus volume (DATA HILANG)
make logs        # log semua service
make test        # unit test
make test-int    # integration test (butuh service jalan)
make lint        # go vet + golangci-lint
make loadtest    # k6 load test
make backup      # backup database
```

## Stack

| Komponen | Teknologi |
| --- | --- |
| Aplikasi | Go HTTP API |
| Dependency | PostgreSQL + Redis |
| Runtime | Docker Compose |
| Reverse proxy | Nginx |
| CI/CD | GitHub Actions |
| Metrics | Prometheus |
| Dashboard & alert | Grafana |
| Logs | Loki + Grafana Alloy |
| Tracing | OpenTelemetry + Tempo |
| Load test | k6 |
| IaC | OpenTofu / Terraform |
| Orchestration | Kubernetes (kind) |

## Aturan bukti yang gue pegang

1. Setiap minggu punya **acceptance criteria** yang harus lolos dengan bukti, bukan perasaan.
2. Output yang mengandung IP internal, token, password, atau data pribadi **disanitasi** sebelum di-commit.
3. Setiap kegagalan yang pernah gue temui harus punya **runbook**.
4. Minggu dianggap selesai kalau: bisa direproduksi, ada buktinya, bisa dijelaskan, ada runbook-nya.

## Peta repository

```
app/                     Go HTTP API (graceful shutdown, /healthz, /readyz, /metrics, OTel)
docker/                  Dockerfile multi-stage + compose.yaml
nginx/                   konfigurasi reverse proxy
.github/workflows/       CI: test, lint, build image ber-tag commit SHA
observability/           prometheus, grafana provisioning, loki, alloy (OTel)
infra/terraform/         OpenTofu/Terraform buat lab cloud
kubernetes/              manifest Deployment, Service, HPA, PDB
loadtest/                skenario k6
labs/week-01 .. week-16  bukti per minggu: notes, commands, runbook, evidence
runbooks/                cara memperbaiki kegagalan
postmortems/             laporan insiden blameless
docs/                    architecture, slo, deployment, disaster-recovery, security-review
```

## Progress

| Bulan | Tema | Status |
| --- | --- | --- |
| 1 | Linux, networking, container, reverse proxy | 0/4 |
| 2 | CI/CD, secret, backup, production readiness | 0/4 |
| 3 | Metrics, alert, log, trace, SLO | 0/4 |
| 4 | IaC, cloud, Kubernetes, load test & chaos | 0/4 |

## Catatan keamanan lab

- Port database dan Redis **tidak diekspos** ke host; cuma Nginx yang punya `ports:`.
- `.env` ada di `.gitignore`; yang ke-commit cuma `.env.example`.
- Backup terenkripsi dan disimpan terpisah dari volume database.
- Eksperimen kegagalan cuma dilakukan di environment lab yang disposable.
