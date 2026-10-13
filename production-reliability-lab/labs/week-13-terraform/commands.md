# Perintah — Week 13 · Terraform / OpenTofu

Format: perintah → output → interpretasi.

## Hari 1–2 (kondisi normal)

```bash
# 1. Cek format
$ tofu fmt -check
# output: (kosong = rapi)  /  main.tf /path
# interpretasi: kosong berarti sesuai gaya HCL. Kalau ada path, jalankan `tofu fmt`.

# 2. Inisialisasi provider + backend
$ tofu init
# output: Terraform has been successfully initialized!
# interpretasi: provider AWS didownload ke .terraform/, backend (state) siap.

# 3. Validasi sintaks
$ tofu validate
# output: Success! The configuration is valid.
# interpretasi: struktur HCL benar. Bukan berarti resource bakal sukses di cloud.

# 4. Buat rencana tersimpan
$ tofu plan -out=tfplan
# output: Plan: 7 to add, 0 to change, 0 to destroy.
# interpretasi: 7 resource baru. File tfplan WAJIB dipakai di apply.

# 5. Review rencana sebagai JSON
$ tofu show -json tfplan | jq '.resource_changes | length'
# output: 7
# interpretasi: ada 7 resource berubah; bisa dipakai buat CI review otomatis.

# 6. Lihat aksi per resource
$ tofu show -json tfplan | jq '.resource_changes[].change.actions'
# output: ["create"]  (7 baris)
# interpretasi: semua create, nggak ada destroy/ replace — aman.

# 7. Eksekusi rencana
$ tofu apply tfplan
# output: Apply complete! Resources: 7 added, 0 changed, 0 destroyed.
# interpretasi: infrastruktur jadi persis seperti plan.

# 8. Ambil output
$ tofu output vpc_id
# output: "vpc-0abc123"
# interpretasi: ID VPC lab, dipakai buat langkah berikutnya.

# 9. Lihat state
$ tofu state list
# output: aws_vpc.main / aws_subnet.public / aws_instance.app / ...
# interpretasi: semua resource yang dikelola state.
```

## Hari 3 (eksperimen kegagalan)

```bash
# 10. Deteksi drift (ubah manual lewat console dulu)
$ tofu plan -refresh-only
# output: # aws_vpc.main: Refreshing state... / ~ tags...
# interpretasi: Terraform lihat perubaan di luar kode = drift.

# 11. Bandingkan dua rencana
$ tofu plan -out=tfplan2
$ tofu show -json tfplan2 | jq '.resource_changes[].change.actions'
# output: ["update"]
# interpretasi: cuma update (ganti tag), BUKAN recreate. Bagus.

# 12. Import resource luar (block, bukan CLI)
$ tofu plan
# output: Plan: 1 to import, 0 to add...
# interpretasi: Terraform akan tulis resource luar ke state lewat import block.

# 13. Scan keamanan IaC
$ trivy config .
# output: kubernetes.io/docs: 0 / terraform: 2 (AVD-aws-... "IMDSv2 not required")
# interpretasi: ada temuan; perbaiki sebelum apply (di lab ini sudah http_tokens="required").

# 14. Cek kebocoran secret di state
$ tofu state show aws_instance.app | grep -i password
# output: (kosong)
# interpretasi: nggak ada password plaintext di state resource ini.

# 15. Cek versi & provider
$ tofu version
$ tofu providers
# output: OpenTofu v1.9.x / aws v6.x.x
# interpretasi: sesuai required_version & pinning.

# 16. CLEANUP wajib
$ tofu destroy
# output: Destroy complete! Resources: 7 destroyed.
# interpretasi: semua resource lab hilang. Cek console buat sisa (NAT/EIP/volume).

# 17. Verifikasi nggak ada state ke-commit
$ git status --porcelain | grep tfstate
# output: (kosong)
# interpretasi: aman, nggak ada *.tfstate ke-commit.
```
