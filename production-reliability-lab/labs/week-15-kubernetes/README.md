# Week 15 · Kubernetes dan orchestration

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai kubernetes dan orchestration sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Pod, Deployment, ReplicaSet, Service
- ConfigMap & Secret
- Probe: startup/readiness/liveness
- requests/limits & QoS
- Rolling update & rollback
- Gateway API vs Ingress

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Deployment + Service jalan di cluster lokal, lalu diganti dengan image API milik lab.
2. Pod yang dihapus digantikan otomatis sesuai desired state.
3. Rollout, rollback, dan probe failure bisa diamati dan dijelaskan.
4. Workload non-root, ada requests/limits, dan PSA restricted aktif.

## Struktur folder

```
labs/week-15-kubernetes/
├── README.md        <- file ini
├── notes.md         <- catatan konsep Hari 1
├── commands.md      <- semua perintah yang dijalankan + output yang disanitasi
├── runbooks/        <- cara memperbaiki kegagalan (Hari 4)
└── evidence/        <- bukti: screenshot, log, hasil pengukuran
```

## Aturan bukti

1. Setiap acceptance criteria di atas harus punya **bukti yang bisa direproduksi** (perintah + output).
2. Output yang mengandung IP internal, token, password, atau data pribadi **wajib disanitasi** sebelum di-commit.
3. Jangan tandai minggu ini selesai hanya karena tutorialnya sudah ditonton.

## Referensi

- [Kubernetes Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/) — Pod, Deployment, Service
- [Kubernetes v1.34 Release](https://kubernetes.io/blog/2025/08/27/kubernetes-v1-34-release/) — fakta versi terbaru
- [Configure Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/) — startup/readiness/liveness
- [Gateway API](https://gateway-api.sigs.k8s.io/) — pengganti Ingress NGINX (retire 2026)
- [Terraform/OpenTofu Docs](https://developer.hashicorp.com/terraform/tutorials) — konteks IaC lab
- [SRE Workbook](https://sre.google/workbook/) — capacity & reliability konteks

