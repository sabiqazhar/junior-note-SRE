# Minggu 13 · Terraform dan Infrastructure as Code — Bahan Bacaan

> *"Infrastruktur yang nggak bisa lo bangun ulang dari nol dalam 1 jam itu bukan infrastruktur, itu warisan budaya."* — filosofi lab ini.

---

## 1. Kenapa penting

Bayangin lo punya server produksi yang diutak-atik manual lewat console cloud selama setahun. Suatu hari server itu mati. Lo buka console, dan nggak ada satupun orang yang ingat persis: port apa saja yang dibuka, security group mana yang nyolok ke mana, subnet-nya private atau public. Lo cuma punya firasat. Itulah mimpi buruk yang disebut **configuration drift** — keadaan aktual di cloud udah nggak cocok lagi sama apa yang lo pikir lo punya.

Infrastructure as Code (IaC) itu cara lo nerjemahin "bentuk infrastruktur" jadi teks yang bisa di-commit ke Git. Kalau server mati, lo tinggal `apply` lagi dan dapet infrastruktur yang persis sama. Lo bisa review perubahan lewat Pull Request, persis kayak review kode. Dan lo bisa rollback ke versi minggu lalu kalau ternyata perubahan kemarin bikin rusak.

Analognya gini: Terraform/OpenTofu itu kayak **resep masak**, bukan foto hasil masakan. Foto cuma nunjukin "tampaknya begini", tapi resep nunjukin "langkah pasti biar hasilnya selalu sama". Kalau dapurnya kebakar, lo tinggal masak ulang dari resep — bukan nebak-nebak "tadi garamnya segenggam apa segini ya?". State file itu catatan "kemarin masaknya sampe langkah berapa", supaya besok lo lanjut, bukan mulai dari nol atau malah bikin dua porsi dobel.

---

## 2. Konsep inti

### 2.1 Provider, resource, variable, output

Provider itu "driver" ke cloud lo (AWS, GCP, dll). Resource itu satuan infrastruktur: satu VPC, satu instance, satu security group. Variable buat input (region, environment), output buat nilai yang lo butuh setelah `apply` (ID VPC, IP).

```hcl
resource "aws_vpc" "main" {
  cidr_block           = "10.20.0.0/16"
  enable_dns_hostnames = true
  tags = { Name = "${var.environment}-vpc", Env = var.environment }
}

output "vpc_id" {
  value       = aws_vpc.main.id
  description = "ID VPC lab"
}
```

### 2.2 State: sumber kebenaran lokal

State (`terraform.tfstate`) nyimpen pemetaan "resource di kode → objek asli di cloud". Tanpa state, Terraform nggak tahu mana yang sudah dibuat dan mana yang belum. **Bahaya:** state bisa berisi password database dalam teks biasa. Makanya state harus di backend **remote + terenkripsi**, dan **jangan pernah di-commit ke Git**.

### 2.3 Drift: keadaan aktual vs harapan

Drift terjadi kalau ada yang ngubah infrastruktur **di luar** Terraform — misal lo hapus security group lewat console. `tofu plan` bakal ngedeteksi: "resource ini ada di state, tapi di cloud bentuknya beda". Deteksi rutin pakai `tofu plan -refresh-only`.

```bash
tofu plan -refresh-only        # cuma bandingkan, jangan ubah
tofu state list                 # lihat resource yang dikelola state
tofu state show aws_instance.app
```

### 2.4 Plan: kontrak sebelum apply

`plan` itu simulasi "apa yang bakal berubah". Simpan hasilnya ke file (`-out=tfplan`) lalu `apply` pakai file itu. Kenapa? Karena `apply` tanpa `-out` **menghitung ulang rencana saat itu juga** — kalau di antaranya ada yang berubah (provider update, resource diubah manual), yang dieksekusi bukan yang lo review.

```bash
tofu fmt -check
tofu init
tofu validate
tofu plan -out=tfplan
tofu show -json tfplan > plan.json    # review pakai jq
tofu apply tfplan                     # JANGAN apply ulang tanpa file ini
```

### 2.5 Provider pinning: cegah breaking change dadakan

Pin versi provider dan Terraform itu wajib. `init` ulang di CI bisa narik versi baru yang nabrak konfigurasi lo.

```hcl
terraform {
  required_version = "~> 1.9"        # cuma terima 1.9.x
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"             # pinning wajib
    }
  }
}
```

### 2.6 OpenTofu vs Terraform: pilih mana?

| | Terraform | OpenTofu |
| --- | --- | --- |
| Lisensi | BUSL 1.1 (bukan open source) sejak 2023 | MPL-2.0 (open source, Linux Foundation) |
| Kepemilikan | HashiCorp — diakuisisi IBM selesai 27 Feb 2025 | Komunitas |
| Fitur khas | HCP Terraform, Sentinel | State encryption fleksibel, OCI registry (sejak 1.10, Jun 2025) |

Buat lab: OpenTofu lebih bebas. Di kantor: kemungkinan besar Terraform. **Sintaks keduanya identik** — perintah `tofu` ganti jadi `terraform`, sisanya sama.

### 2.7 Import block & scan keamanan

Jangan pakai `terraform import` CLI (state langsung berubah tanpa review). Pakai `import` block sejak TF 1.5, lalu `plan` untuk lihat apa yang ditulis:

```hcl
import {
  to = aws_vpc.main
  id = "vpc-0abc123"
}
```

Scan keamanan IaC (tfsec sudah dilebur ke Trivy):

```bash
trivy config .
terraform state show ... | grep -i password
```

---

## 3. Mekanisme di balik layar

Gimana sebenarnya `tofu apply` jalan, step by step:

1. **`init`** — unduh provider (AWS) dan setup backend. Provider disimpen di `.terraform/`, bukan di state.
2. **Build dependency graph** — Terraform baca semua `.tf`, bangun graf siapa nunggu siapa (VPC harus ada dulu sebelum subnet).
3. **Refresh state** — bandingkan state lama vs keadaan aktual di cloud (nah di sinilah drift kelihatan).
4. **Plan / diff** — hitung resource yang perlu `create`, `update`, atau `destroy`. Untuk update, Terraform cari **diff minimal** (misal cuma ganti tag, bukan bikin ulang).
5. **Apply** — eksekusi graph sesuai urutan topological. Tiap resource di-lock biar nggak dobel jalan (kalau backend pakai DynamoDB lock).
6. **Update state** — tulis hasil akhir ke state (remote+encrypt) supaya besok jadi acuan refresh.

Kuncinya: state bukan sekadar "log", tapi **input untuk langkah berikutnya**. Makanya state rusak = bencana, karena Terraform kehilangan peta.

---

## 4. Glosarium mini

| Istilah | Arti singkat |
| --- | --- |
| Provider | Plugin penghubung ke cloud/platform (aws, gcp, kubernetes) |
| Resource | Satu objek infrastruktur yang dikelola (vpc, instance) |
| State | File pemetaan resource kode ↔ objek asli di cloud |
| Drift | Selisih antara keadaan aktual cloud vs konfigurasi |
| Plan | Simulasi perubahan sebelum dieksekusi |
| Apply | Eksekusi rencana untuk bikin/ubah/destroy resource |
| Backend | Tempat state disimpan (lokal atau remote, misal S3) |
| Module | Kumpulan resource yang bisa dipakai ulang |
| Pinning | Mengunci versi provider/Terraform biar stabil |
| Import | Masukin resource yang sudah ada ke dalam state |

---

## 5. Bahan bacaan wajib

1. **OpenTofu Documentation** — https://opentofu.org/ (open source fork, panduan `tofu` resmi)
2. **Terraform Tutorials** — https://developer.hashicorp.com/terraform/tutorials (lab resmi HashiCorp)
3. **Terraform Language Docs** — https://developer.hashicorp.com/terraform/language (referensi HCL: resource, state, backend)
4. **Trivy Configuration Scanning** — https://trivy.dev/latest/docs/configuration/ (scan keamanan IaC)
5. **AWS VPC Docs** — https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html (konteks resource cloud)
6. **SRE Book: Postmortem Culture** — https://sre.google/sre-book/postmortem-culture/ (kenapa reproducibility penting)

---

## 6. Self-check

1. Apa bedanya `plan` dan `apply`, dan kenapa `apply` tanpa `-out` bahaya?
2. Kenapa state nggak boleh di-commit ke Git? Apa risikonya?
3. Gimana caranya lo deteksi drift yang terjadi karena ada yang hapus resource lewat console?
4. Apa fungsi `required_providers { version = "~> 6.0" }` dan apa yang terjadi kalau lo hapus?
5. Kapan lo pakai `import` block, dan kenapa nggak `terraform import` CLI?
6. Bedanya OpenTofu vs Terraform dari sisi lisensi — dan kenapa di kantor lo bisa ketemu yang satu padahal di lab pakai yang lain?
7. Kenapa nggak boleh pakai workspace buat pisah dev/prod dalam satu state?

---

## 7. Yang sering salah paham

1. **"State itu cuma log."** Salah. State adalah input kritis untuk `plan` berikutnya. State rusak/hilang = Terraform ngerasa semua resource belum dibuat → bikin duplikat atau gagal.
2. **"`apply` aja langsung, plan mah lambat."** Ini penyebab insiden nomor satu. Rencana yang lo lihat bisa beda dari yang dieksekusi kalau ada perubahan di antaranya.
3. **"Workspace buat pisah environment."** Blast radius-nya gede banget — satu state rusak, dev DAN prod kena. Pakai directory per environment + backend terpisah.
4. **"Provider tanpa pin aman, toh versi stabil."** CI yang `init` ulang bisa narik minor version baru yang nabrak. Pin itu pagar pengaman, bukan birokrasi.
5. **"Lupa `destroy`, besok aja."** VPC dan NAT gateway jadi tagihan bulanan. Matikan pas lab selesai.
6. **"OpenTofu dan Terraform beda banget."** Secara konsep dan sintaks nyaris identik. Lo cukup paham satu, bisa pakai dua.
