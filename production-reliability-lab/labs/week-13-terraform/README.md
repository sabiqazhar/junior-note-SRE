# Week 13 · Terraform dan Infrastructure as Code

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai terraform dan infrastructure as code sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- Provider, resource, variable, output
- State, locking & drift
- Module & dependency graph
- Policy as code (Trivy/Checkov)
- Import block

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. `terraform validate` / `tofu validate` lolos dan `fmt -check` bersih.
2. Perubahan bisa direview lewat `plan -out` sebelum apply.
3. Provisioning dan cleanup terdokumentasi; biaya dipahami sebelum apply.
4. Tidak ada state file atau secret ter-commit (backend terenkripsi + .gitignore).

## Struktur folder

```
labs/week-13-terraform/
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

- [OpenTofu Documentation](https://opentofu.org/) — panduan `tofu` resmi (open source, MPL-2.0)
- [Terraform Tutorials](https://developer.hashicorp.com/terraform/tutorials) — lab resmi HashiCorp
- [Terraform Language Docs](https://developer.hashicorp.com/terraform/language) — referensi HCL: resource, state, backend
- [Trivy Configuration Scanning](https://trivy.dev/latest/docs/configuration/) — scan keamanan IaC (tfsec sudah dilebur)
- [AWS VPC Docs](https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html) — konteks resource cloud
- [SRE Book: Postmortem Culture](https://sre.google/sre-book/postmortem-culture/) — kenapa reproducibility penting

