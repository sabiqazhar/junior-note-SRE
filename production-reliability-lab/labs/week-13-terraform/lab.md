# Lab — Minggu 13 · Terraform dan Infrastructure as Code

> Ikuti BERURUTAN, jangan lompat. Setiap langkah formatnya: perintah → Yang harus terjadi → Kalau gagal → Catat notes.

## Persiapan — BACA INI SEBELUM APPLY

⚠️ **Peringatan biaya:** Terraform di lab ini bikin resource AWS NYATA yang bisa ditagih. Yang paling sering bikin kaget: **NAT Gateway (~$0,045/jam + biaya data processing)** dan **Elastic IP/volume yang kesisa**. EC2 `t3.small` ~$0,02/jam — matikan di luar jam lab.

⚠️ **AWS Free Tier sudah berubah per 15 Juli 2025:** bukan lagi "12 bulan gratis", tapi **kredit $100 (+$100) berlaku 6 bulan**. Habis itu, tagihan jalan otomatis.

🔔 **Pasang budget alert SEBELUM apply pertama** (bukan sesudahnya):

```bash
# 1. Catat account ID
aws sts get-caller-identity --query Account --output text

# 2. Buat budget $20 (atau batas nyaman lo)
cat > budget.json <<'EOF'
{
  "BudgetLimit": {"Amount":"20","Unit":"USD"},
  "TimeUnit":"MONTHLY",
  "BudgetType":"COST",
  "BudgetName":"sre-lab-budget"
}
EOF
cat > notify.json <<'EOF'
[{
  "Notification": {"NotificationType":"ACTUAL","ComparisonOperator":"GREATER_THAN","Threshold":50,"ThresholdType":"PERCENTAGE"},
  "Subscribers": [{"SubscriptionType":"EMAIL","Address":"lo@example.com"}]
}]
EOF
aws budgets create-budget --account-id $ACC \
  --budget file://budget.json \
  --notifications-with-subscribers file://notify.json
# Yang harus terjadi: "Budget sre-lab-budget created."
# Kalau gagal: cek IAM permission budgets:*, dan pastikan $ACC ter-set.
```

- [ ] Budget alert aktif di ≤ 50% batas nyaman.
- [ ] Alarm terpisah untuk "pengeluaran naik tajam" (optional tapi bagus).
- [ ] Sudah siap jalankan `tofu destroy` di Hari 3 & 4.

---

## Hari 1 · Teori (2 jam)

1. **Baca** `bahan-bacaan.md` sampai selesai → **Yang harus terjadi:** lo paham state/plan/drift. → **Kalau gagal:** balik ke section 1 & 2. → **Catat notes:** tulis analogi "resep vs foto" di `notes.md`.
2. **Buka** `infra/terraform/main.tf` → **Yang harus terjadi:** paham `required_version`, `required_providers`, `aws_vpc`, `metadata_options` (IMDSv2). → **Kalau gagal:** baca Terraform Language Docs. → **Catat notes:** apa arti `version = "~> 6.0"`.
3. **Bedah** tabel OpenTofu vs Terraform → **Yang harus terjadi:** lo bisa jelasin bedanya tanpa buka catatan. → **Kalau gagal:** ulang section 2.6. → **Catat notes:** kapan pakai yang mana.
4. **Latihan mental:** bayangin server mati, tulis urutan lo rebuild dari `tofu apply`. → **Yang harus terjadi:** urutan runtut (init→plan→apply). → **Kalau gagal:** ulang section 3. → **Catat notes:** kenapa state penting.

---

## Hari 2 · Praktik (2 jam)

1. **`tofu fmt -check`** → **Yang harus terjadi:** `main.tf` rapi, nol error. → **Kalau gagal:** `tofu fmt` lalu commit. → **Catat notes:** bedanya fmt vs validate.
2. **`tofu init`** → **Yang harus terjadi:** "Terraform has been successfully initialized!". → **Kalau gagal:** cek kredensial AWS (`aws configure`), cek `required_providers`. → **Catat notes:** apa yang didownload (provider di `.terraform/`).
3. **`tofu validate`** → **Yang harus terjadi:** "Success! The configuration is valid." → **Kalau gagal:** baca error baris berapa. → **Catat notes:** validate cek apa (sintaks, bukan keadaan cloud).
4. **`tofu plan -out=tfplan`** → **Yang harus terjadi:** "Plan: N to add, 0 to change, 0 to destroy." → **Kalau gagal:** cek region, cek quota. → **Catat notes:** simpan `tfplan`.
5. **`tofu show -json tfplan | jq '.resource_changes[].change.actions'`** → **Yang harus terjadi:** array `["create"]`. → **Kalau gagal:** pastikan `tfplan` ada. → **Catat notes:** ini bukti review otomatis.
6. **`tofu apply tfplan`** → **Yang harus terjadi:** "Apply complete! Resources: N added." → **Kalau gagal:** baca error (sering: quota/permission). → **Catat notes:** ID VPC dari `tofu output vpc_id`.
7. **Verifikasi output:** **`tofu output`** → **Yang harus terjadi:** daftar vpc_id, private_subnet_id, app_instance_id. → **Kalau gagal:** cek `output` block. → **Catat notes:** simpan ke `evidence/`.

---

## Hari 3 · Eksperimen (2 jam) — sengaja rusak

1. **Drift manual:** ubah sesuatu lewat console (misal ganti tag VPC) → **`tofu plan -refresh-only`** → **Yang harus terjadi:** Terraform laporkan perubahan. → **Kalau gagal:** pastikan state remote nyambung. → **Catat notes:** ini cara lo tahu ada drift.
2. **Ubah kode, bandingkan:** edit `main.tf` (tambah tag) → **`tofu plan -out=tfplan2`** → **`tofu show -json tfplan2 | jq '.resource_changes[].change.actions'`** → **Yang harus terjadi:** `["update"]` bukan recreate. → **Kalau gagal:** cek apa yang bikin force-replace. → **Catat notes:** bedanya update vs replace.
3. **Import block:** asumsikan ada VPC luar → tulis `import { to = aws_vpc.main id = "vpc-0abc" }` → **`tofu plan`** → **Yang harus terjadi:** Terraform tunjukin apa yang akan ditulis ke config. → **Kalau gagal:** pastikan ID benar. → **Catat notes:** kenapa `import` block lebih aman dari CLI.
4. **Scan:** **`trivy config .`** → **Yang harus terjadi:** daftar temuan (atau "no vulnerabilities"). → **Kalau gagal:** install trivy. → **Catat notes:** temuan paling mencolok.
5. **Cek secret di state:** **`tofu state show aws_instance.app | grep -i password`** → **Yang harus terjadi:** kosong. → **Kalau gagal:** rotate & pindah ke Secret Manager. → **Catat notes:** bukti state bersih.

🧹 **HARI 3 WAJIB — CLEANUP (lab disposable):** semua resource minggu 13 cuma buat belajar, bukan produksi.

```bash
tofu destroy
# Yang harus terjadi: "Destroy complete! Resources: N destroyed."
# Kalau gagal: baca resource yang nolak dihapus (misal ada dependency), hapus manual di console lalu destroy lagi.
# SETELAH destroy: cek console AWS untuk resource sisa (NAT gateway, EIP, volume) — itu yang sering tinggal & ditagih.
```

---

## Hari 4 · Review (2 jam)

1. **Tulis** bagian Provisioning di `docs/infrastructure.md` (init→plan→apply + prasyarat). → **Yang harus terjadi:** ada langkah reproduksi. → **Kalau gagal:** lihat template di `docs/infrastructure.md`.
2. **Isi** tabel Perubahan (PR → `plan -out` → review → `apply tfplan`). → **Yang harus terjadi:** alur review jelas.
3. **Isi** tabel Cleanup + Biaya (estimasi per resource, budget alert). → **Yang harus terjadi:** budget alert disebut.
4. **Pastikan** nggak ada `*.tfstate` ke-commit (cek `.gitignore`). → **Yang harus terjadi:** `git status` bersih dari tfstate.

---

## Cheat sheet per hari

| Hari | Fokus | Perintah inti |
| --- | --- | --- |
| 1 | Teori | baca `bahan-bacaan.md`, bedah `main.tf` |
| 2 | Praktik | `tofu fmt -check` → `init` → `validate` → `plan -out` → `apply tfplan` |
| 3 | Rusak | `plan -refresh-only`, `import {}`, `trivy config .`, `destroy` |
| 4 | Review | isi `docs/infrastructure.md`, cek `.gitignore` |

## Rubrik observasi per hari

| Hari | Yang lo amati | Tanda lulus |
| --- | --- | --- |
| 1 | Paham konsep state/plan/drift | Bisa jelasin tanpa buka catatan |
| 2 | Resource benar-benar jalan | `apply tfplan` sukses, `output` muncul |
| 3 | Kegagalan kecil terbukti | Drift & rollback & scan tercatat |
| 4 | Dokumentasi lengkap | `docs/infrastructure.md` + `.gitignore` ok |

## Troubleshooting cepat

| Error | Penyebab | Solusi |
| --- | --- | --- |
| `init` gagal | versi provider ke-pin nggak ada | longgarkan `~> 6.0`, cek network |
| `plan` bilang destroy+recreate | edit attribute yang force replace | cek diff, pisah perubahan |
| `apply` bikin dobel | state hilang | `import` block, jangan buat ulang |
| drift terus muncul | ada yang utak-atik console | kembalikan lewat kode |
| lock gagal | dua `apply` berbarengan | tunggu / 1 proses |
| tagihan meledak | lupa `destroy` | `tofu destroy` + cek sisa |

## Alat bantu wajib

| Alat | Kegunaan | Cara cek |
| --- | --- | --- |
| `tofu` / `terraform` | eksekusi IaC | `tofu version` |
| `jq` | parse JSON plan | `jq --version` |
| `trivy` | scan keamanan | `trivy --version` |
| `aws` CLI | budget alert | `aws --version` |

## Langkah pemulihan darurat

- State rusak/hilang: jangan `apply` buta — pakai `import` block satu per satu.
- `apply` gagal di tengah: `tofu plan` lagi untuk lihat sisa, lalu lanjut `apply tfplan`.
- Resource nolak `destroy`: cek dependency (`aws ec2 describe-*`), hapus manual lalu `destroy` lagi.

## Verifikasi akhir

- [ ] `git status` bersih dari `*.tfstate`.
- [ ] Budget alert tercatat aktif.
- [ ] Bukti `plan`/`apply`/`destroy` ada di `evidence/`.

## Daftar periksa

- [ ] `tofu validate` lolos, `fmt -check` bersih.
- [ ] Budget alert aktif SEBELUM apply.
- [ ] Tiap perubahan lewat `plan -out` → `apply tfplan`.
- [ ] Drift berhasil dideteksi hari 3.
- [ ] `trivy config .` pernah dijalankan.
- [ ] `tofu destroy` selesai + cek resource sisa di console.
- [ ] Nggak ada `*.tfstate` / secret ke-commit.

## Cara buktikan

Lo lulus minggu ini kalau: (1) `tofu apply tfplan` menghasilkan resource persis seperti `plan`, (2) `tofu plan -refresh-only` bisa nangkep drift yang lo bikin manual, (3) sesudah `tofu destroy` nggak ada resource AWS yang nyangkut (cek console). Bukti disimpan di `evidence/` dengan output yang sudah disanitasi (sensor IP/token).
