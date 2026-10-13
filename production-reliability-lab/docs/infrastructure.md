# Infrastructure as Code

Alat: **OpenTofu** (open source) atau **Terraform** (BUSL, di bawah HashiCorp/IBM). Sintaks sama; dokumentasi ini memakai perintah `tofu`, ganti jadi `terraform` kalau perlu.

## Kenapa dua-duanya perlu dipahami

| | Terraform | OpenTofu |
| --- | --- | --- |
| Lisensi | BUSL 1.1 sejak 2023 | MPL-2.0, Linux Foundation |
| Kepemilikan | HashiCorp — akuisisi IBM selesai 27 Feb 2025 | Komunitas |
| Fitur khas | HCP Terraform, Sentinel | State encryption fleksibel, OCI registry (sejak 1.10, Jun 2025) |

Buat lab: OpenTofu lebih bebas. Di kantor: kemungkinan besar Terraform. Konsepnya sama persis.

## Struktur

```
infra/terraform/
├── main.tf                  resource utama
├── variables.tf             input
├── outputs.tf               nilai yang dibutuhkan setelah apply
└── terraform.tfvars.example contoh input non-rahasia
```

## Alur kerja yang aman

```bash
tofu fmt -check                 # format
tofu init                       # provider + backend
tofu validate                   # sintaks
tofu plan -out=tfplan           # SIMPAN rencanya
tofu show -json tfplan > plan.json
# review plan.json (atau output plan), baru:
tofu apply tfplan               # pakai file plan, JANGAN apply ulang
```

**Kenapa `plan -out` itu wajib:** `tofu apply` tanpa file plan akan menghitung ulang rencana saat itu juga. Kalau ada yang berubah di antaranya (resource diubah manual, provider update), yang dieksekusi bukan yang sudah lo review.

## State

| Aturan | Alasan |
| --- | --- |
| Backend remote + enkripsi | State bisa berisi password database dalam teks biasa |
| Locking (DynamoDB/equivalent) | Cegah dua `apply` berbarengan |
| Satu state per environment | Jangan pakai workspace buat dev vs prod — blast radius terlalu besar |
| Jangan commit `*.tfstate` | Sudah ada di `.gitignore` |
| Rutin `tofu plan -refresh-only` | Deteksi drift: kondisi aktual vs konfigurasi |

## Deteksi drift

```bash
tofu plan -refresh-only            # cuma bandingkan, jangan ubah
tofu state list
tofu state show aws_instance.app
```

## Import resource yang sudah ada

Jangan pakai perintah `terraform import` (state langsung berubah tanpa review). Pakai `import` block sejak TF 1.5:

```hcl
import {
  to = aws_vpc.main
  id = "vpc-0abc123"
}
```

Lalu `tofu plan` untuk melihat apa yang akan ditulis ke konfigurasi.

## Scan keamanan

```bash
trivy config .                     # tfsec sudah dilebur ke Trivy
terraform state show ... | grep -i password   # cek kebocoran di state
```

## Cleanup — jangan sampai jadi tagihan bulanan

```bash
tofu destroy
```

Sebelum destroy, pastikan:
- [ ] Backup yang mau disimpan sudah dipindah ke luar resource ini
- [ ] Tidak ada resource yang dibuat manual di luar Terraform (itu yang sering tertinggal)
- [ ] Setelah destroy, cek console cloud untuk resource sisa (NAT gateway, elastic IP, volume)

## Biaya sebelum apply

| Resource | Perkiraan/jam | Catatan |
| --- | --- | --- |
| EC2 t3.small | ~$0,02 | Matikan di luar jam lab |
| NAT Gateway | ~$0,045 + data | Paling sering jadi biaya tak terduga |
| Managed DB kelas kecil | ~$0,02–0,05 | Plus storage |
| Object storage | per GB | Retensi backup |

Pasang **budget alert sebelum apply pertama**, bukan sesudahnya.
