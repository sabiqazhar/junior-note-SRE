# Week 14 · Cloud networking dan security

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai cloud networking dan security sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- VPC, subnet, route table
- Security group/firewall
- IAM least privilege
- Load balancer & managed DB
- Cost guardrail

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Hanya port yang dibutuhkan yang terbuka dari internet; port DB tidak bisa diakses publik.
2. Akses administrasi pakai least privilege (role, bukan user permanen).
3. Diagram jaringan sesuai konfigurasi aktual (bukan hasil tebakan).
4. Budget alert aktif supaya lab tidak jadi tagihan kejutan.

## Struktur folder

```
labs/week-14-cloud/
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

- [AWS VPC User Guide](https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html) — VPC, subnet, route table
- [AWS IAM User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/introduction.html) — role, policy, least privilege
- [AWS Budgets](https://docs.aws.amazon.com/cost-management/latest/userguide/budgets-managing-costs.html) — cost guardrail
- [Terraform Language Docs](https://developer.hashicorp.com/terraform/language) — referensi HCL/resource AWS
- [OpenTofu Docs](https://opentofu.org/) — alternatif IaC open source
- [SRE Book: Postmortem Culture](https://sre.google/sre-book/postmortem-culture/) — disiplin review insiden

