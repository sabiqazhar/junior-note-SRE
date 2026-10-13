# Ringkasan — Minggu 13 · Terraform dan Infrastructure as Code

> Satu halaman untuk dibaca ulang sebelum ujian/wawancara. Bukan pengganti bahan-bacaan.md.

## Inti 5 kalimat

1. Infrastructure as Code (IaC) nerjemahin bentuk infrastruktur jadi teks di Git, supaya lo bisa bangun ulang persis sama kapan saja.
2. **State** adalah peta "kode ↔ objek asli di cloud"; dia bukan log, tapi input buat `plan` berikutnya — makanya nggak boleh bocor/di-commit.
3. **Plan** adalah kontrak: selalu `plan -out` lalu `apply tfplan`, jangan `apply` telanjang yang bisa jalanin rencana beda dari yang lo review.
4. **Drift** muncul kalau ada yang utak-atik di luar Terraform; deteksi rutin pakai `tofu plan -refresh-only`.
5. OpenTofu dan Terraform punya sintaks identik — bedanya cuma lisensi (MPL-2.0 vs BUSL) dan kepemilikan; pin versi provider biar nggak ada kejutan.

## Peta konsep

```
kode .tf ──init──> provider + backend
   │
   ├─ plan ──> graph (siapa nunggu siapa) + refresh state + diff
   │              │
   │              └─ -out=tfplan ──> apply tfplan ──> update STATE (remote+encrypt)
   │
   ├─ plan -refresh-only ──> DETEKSI DRIFT
   ├─ import block ──> masukin resource luar ke state (review dulu)
   └─ trivy config . ──> scan keamanan IaC

OpenTofu (MPL)  ≈  Terraform (BUSL)   →  sintaks SAMA, ganti `tofu`↔`terraform`
```

## Perintah kunci + output

```bash
# 1. Format & init
$ tofu fmt -check && tofu init
# output: Terraform has been successfully initialized!
# interpretasi: provider AWS didownload, backend (state) siap.

# 2. Validasi sintaks
$ tofu validate
# output: Success! The configuration is valid.
# interpretasi: struktur HCL benar; BUKAN berarti resource sukses di cloud.

# 3. Plan tersimpan
$ tofu plan -out=tfplan
# output: Plan: 7 to add, 0 to change, 0 to destroy.
# interpretasi: 7 resource baru; file tfplan WAJIB dipakai di apply.

# 4. Lihat JSON plan (review otomatis)
$ tofu show -json tfplan | jq '.resource_changes | length'
# output: 7
# interpretasi: ada 7 resource berubah; bisa dipakai CI review.

# 5. Lihat aksi per resource
$ tofu show -json tfplan | jq '.resource_changes[].change.actions'
# output: ["create"] (7 baris)
# interpretasi: semua create, nggak ada replace — aman.

# 6. Apply pakai file plan
$ tofu apply tfplan
# output: Apply complete! Resources: 7 added.
# interpretasi: infrastruktur jadi persis seperti plan.

# 7. Ambil output
$ tofu output vpc_id
# output: "vpc-0abc123"
# interpretasi: ID VPC lab, dipakai langkah berikutnya.

# 8. State list
$ tofu state list
# output: aws_vpc.main, aws_subnet.public, aws_instance.app, ...
# interpretasi: resource yang dikelola state.

# 9. Deteksi drift
$ tofu plan -refresh-only
# output: # aws_instance.app: Refreshing state... ~ tags
# interpretasi: Terraform lihat perubahan di luar kode = drift.

# 10. Import block (bukan CLI)
$ tofu plan
# output: Plan: 1 to import, 0 to add
# interpretasi: resource luar akan masuk state lewat import block.

# 11. Scan keamanan
$ trivy config .
# output: terraform: 0 (atau daftar temuan)
# interpretasi: IaC lolos / ada yang perlu diperbaiki.

# 12. Cek kebocoran secret
$ tofu state show aws_instance.app | grep -i password
# output: (kosong)
# interpretasi: nggak ada password plaintext di state.

# 13. CLEANUP wajib
$ tofu destroy
# output: Destroy complete! Resources: 7 destroyed.
# interpretasi: semua resource lab hilang; cek sisa di console.

# 14. Pastikan nggak ada state ke-commit
$ git status --porcelain | grep tfstate
# output: (kosong)
# interpretasi: aman, nggak ada *.tfstate ke-commit.
```

## Failure matrix

| Gejala | Kemungkinan penyebab | Cek | Penyelesaian |
| --- | --- | --- | --- |
| `init` gagal download provider | versi provider ke-pin nggak ada / network | `tofu version`, `required_providers` | longgarkan `~> 6.0`, cek akses registry |
| `apply` bikin resource dobel | state hilang/rusak | `tofu state list` | `import` block, jangan buat ulang |
| Rencana beda dari review | `apply` tanpa `-out` | banding `tfplan` vs `show` | selalu `apply tfplan` |
| Drift detector ribut terus | ada yang utak-atik console | `tofu plan -refresh-only` | kembalikan lewat kode, bukan console |
| `lock` gagal (DynamoDB) | dua `apply` berbarengan | error "lock" | tunggu / pastikan 1 proses |
| Secret ketahuan di state | output `password` | `tofu state show ... \| grep -i password` | pindah ke Secret Manager, rotate |
| Tagihan meledak | lupa `destroy` | AWS Cost Explorer | `tofu destroy`, cek NAT/EIP/volume |

## Jebakan

- `apply` tanpa `-out` → rencana beda dari yang direview.
- Workspace buat dev/prod dalam 1 state → blast radius gede.
- State di Git/lokal → secret bocor plaintext.
- Provider tanpa pin → `init` CI narik versi nabrak.
- Lupa `destroy` → NAT gateway jadi tagihan bulanan.
- Pakai `terraform import` CLI → state berubah tanpa review.

## Cheat sheet cepat

| Mau ngapain | Perintah |
| --- | --- |
| Cek rapi | `tofu fmt -check` |
| Cek valid | `tofu validate` |
| Simulasi aman | `tofu plan -out=tfplan` |
| Eksekusi pasti | `tofu apply tfplan` |
| Cek drift | `tofu plan -refresh-only` |
| Lihat state | `tofu state list` / `tofu state show <x>` |
| Masukin resource luar | `import { to = ... id = ... }` |
| Scan | `trivy config .` |
| Bersih | `tofu destroy` |

## Tanya jawab kilat

- **Q: State boleh di-commit?** A: TIDAK. Isinya bisa berisi secret plaintext.
- **Q: `apply` langsung vs `apply tfplan`?** A: Selalu `tfplan` biar yang dieksekusi = yang direview.
- **Q: Drift itu apa?** A: selisih keadaan aktual cloud vs konfigurasi.
- **Q: OpenTofu vs Terraform?** A: sintaks sama, lisensi beda (MPL vs BUSL).
- **Q: Kenapa pin provider?** A: cegah breaking change dadakan di CI.

## Acceptance

1. `tofu validate` lolos, `fmt -check` bersih.
2. Tiap perubahan direview lewat `plan -out` sebelum `apply`.
3. Provisioning + cleanup terdokumentasi; biaya dipahami sebelum `apply`.
4. Nggak ada `*.tfstate` atau secret ke-commit (backend terenkripsi + `.gitignore`).

## Studi kasus mini

- **Kasus A:** lo `apply` langsung tanpa `-out`, dan di antaranya rekan lo nambah SG lewat console. Pas `apply`, yang dieksekusi bukan plan lo → SG dobel. Pelajaran: selalu `plan -out` lalu `apply tfplan`.
- **Kasus B:** state di-commit ke Git, ternyata `aws_db_instance` keluarkan password plaintext → bocor di PR. Pelajaran: backend remote + encrypt, `.gitignore` `*.tfstate`.
- **Kasus C:** `init` di CI narik provider `7.0` padahal kode nulis `~> 6.0` → resource baru nabrak. Pelajaran: pin versi di `required_providers`.

## Istilah kunci (flashcard)

- **State** = peta resource kode ↔ objek cloud; input buat plan berikutnya.
- **Drift** = selisih keadaan aktual vs konfigurasi.
- **Plan** = simulasi perubahan; simpan ke `-out`.
- **Apply** = eksekusi plan; pakai file plan, bukan telanjang.
- **Backend** = tempat state (remote + encrypt disarankan).
- **Pinning** = kunci versi provider/Terraform.
- **Import block** = masukin resource luar ke state (review dulu).
- **Module** = kumpulan resource reusable.

## Pertanyaan wawancara

1. Kenapa `apply` tanpa `-out` bahaya?
2. Apa risiko state di-commit ke Git?
3. Bagaimana detect drift?
4. Kapan pakai `import` block?
5. Bedanya OpenTofu vs Terraform?

## Pola yang sering kelewat

- Lupa `destroy` → tagihan bulanan.
- Pakai workspace buat dev/prod (blast radius gede).
- Provider tanpa pin.
- Scan IaC (`trivy config .`) dilewatin.

## Cek cepat sebelum commit

| Cek | Perintah |
| --- | --- |
| Rapi | `tofu fmt -check` |
| Valid | `tofu validate` |
| Review aman | `tofu plan -out=tfplan` |
| Nggak ada state | `git status \| grep tfstate` |

## Satu kalimat

> IaC itu resep, bukan foto: state adalah peta yang harus dijaga, dan `plan -out` → `apply tfplan` adalah satu-satunya cara aman ngeksekusi perubahan.
