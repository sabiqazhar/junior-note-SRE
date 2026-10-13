# Week 05 · CI/CD dan release management

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai ci/cd dan release management sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- CI vs CD
- Immutable artifact
- Versioning & traceability
- Environment & secrets di CI
- Rollback berbasis digest

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Test + lint jalan otomatis tiap PR dan selesai <10 menit.
2. Setiap commit punya image dengan tag berbasis SHA yang bisa dilacak.
3. Test sengaja digagalkan → pipeline menolak release (bukti: screenshot/log run merah).
4. Tidak ada secret tercetak di log/artefak; `permissions:` ditulis eksplisit.

## Struktur folder

```
labs/week-05-cicd/
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

- [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions) — `permissions`, `concurrency`, `needs`, `if`. Dokumentasi resmi, paling otoritatif buat nulis `ci.yml`.
- [Configuring OpenID Connect in cloud providers](https://docs.github.com/en/actions/security-for-github-actions/security-hardening-your-deployments/configuring-openid-connect-in-cloud-providers) — cara "login" ke cloud tanpa secret statis (`id-token: write`).
- [Using artifact attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations) — provenance: apa yang diklaim attestation dan gimana verifikasinya.
- [`gh attestation verify` — GitHub CLI manual](https://cli.github.com/manual/gh_attestation_verify) — semua flag verifikasi (`--signer-workflow`, `-R`).
- [About security hardening with GitHub Actions](https://docs.github.com/en/actions/security-for-github-actions/security-hardening-your-deployments/about-security-hardening-with-openid-connect) — bahaya `pull_request_target` + checklist script injection.
- [gitleaks](https://github.com/gitleaks/gitleaks) — secret scanner buat CI dan history repo.

