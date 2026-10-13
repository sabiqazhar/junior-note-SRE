# Week 06 · Secrets, konfigurasi, dan deployment

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai secrets, konfigurasi, dan deployment sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- 12-factor config
- Secret injection
- Readiness vs liveness vs startup probe
- Graceful shutdown di Go
- Rolling/blue-green & batas Compose

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Aplikasi berhenti graceful saat SIGTERM; request in-flight selesai, 0 5xx saat restart.
2. Readiness TIDAK mengatakan siap saat dependency wajib belum siap.
3. Deployment gagal tidak menimpa versi sehat tanpa rencana recovery.
4. Tidak ada secret plaintext di Git (SOPS/age atau secret manager).

## Struktur folder

```
labs/week-06-deployment/
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

- [Container probes — Kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#container-probes) — beda liveness/readiness/startup dan efeknya kalau disalahpahami.
- [Configure liveness/readiness/startup probes — Kubernetes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/) — contoh YAML tiap probe.
- [`signal.NotifyContext` — Go pkg.go.dev](https://pkg.go.dev/os/signal#NotifyContext) — gimana context kebakar pas dapat SIGINT/SIGTERM.
- [The Twelve-Factor App: Config](https://12factor.net/config) — kenapa config lewat env dan secret dipisah.
- [SOPS — GitHub](https://github.com/getsops/sops) — enkripsi file secret dengan age / KMS / PGP.
- [Rolling updates & rollbacks — Kubernetes](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment) — `maxSurge`, `maxUnavailable`, `rollout undo`.

