# Lab — Minggu 14 · Cloud Networking dan Security

> Ikuti BERURUTAN, jangan lompat. Setiap langkah formatnya: perintah → Yang harus terjadi → Kalau gagal → Catat notes.

## Persiapan — BACA INI SEBELUM APPLY

⚠️ **Peringatan biaya:** minggu ini bikin resource AWS NYATA. Tersangka tagihan: **NAT Gateway (~$0,045/jam + biaya data processing)**, EC2 `t3.small` (~$0,02/jam), managed DB, object storage per GB.

⚠️ **AWS Free Tier berubah 15 Juli 2025:** kredit **$100 (+$100) berlaku 6 bulan**, bukan lagi "12 bulan gratis". Habis itu tagihan jalan otomatis.

🔔 **Pasang budget alert SEBELUM apply pertama:**

```bash
aws sts get-caller-identity --query Account --output text   # catat $ACC
cat > budget.json <<'EOF'
{ "BudgetLimit":{"Amount":"20","Unit":"USD"}, "TimeUnit":"MONTHLY",
  "BudgetType":"COST", "BudgetName":"sre-lab-budget" }
EOF
cat > notify.json <<'EOF'
[{ "Notification":{"NotificationType":"ACTUAL","ComparisonOperator":"GREATER_THAN",
   "Threshold":50,"ThresholdType":"PERCENTAGE"},
   "Subscribers":[{"SubscriptionType":"EMAIL","Address":"lo@example.com"}] }]
EOF
aws budgets create-budget --account-id $ACC \
  --budget file://budget.json --notifications-with-subscribers file://notify.json
# Yang harus terjadi: "Budget sre-lab-budget created."
# Kalau gagal: cek permission budgets:*, pastikan $ACC ter-set.
```

- [ ] Budget alert aktif di ≤ 50% batas nyaman.
- [ ] Siap jalankan `tofu destroy` setelah lab.

---

## Hari 1 · Teori (2 jam)

1. **Baca** `bahan-bacaan.md` section 1–2 → **Yang harus terjadi:** paham public vs private, SG rujuk SG. → **Kalau gagal:** ulang section 2.1–2.2. → **Catat notes:** gambar diagram di `notes.md`.
2. **Buka** `infra/terraform/main.tf` → **Yang harus terjadi:** temukan `aws_security_group.db` cuma buka dari `aws_security_group.app`. → **Kalau gagal:** baca section 2.2. → **Catat notes:** kenapa nggak `0.0.0.0/0`.
3. **Pahami** IAM role + IMDSv2 → **Yang harus terjadi:** jelasin celah SSRF. → **Kalau gagal:** ulang section 2.3–2.4. → **Catat notes:** beda IMDSv1 vs v2.
4. **Gambar diagram dulu** (sebelum nulis config) → **Yang harus terjadi:** ada sketsa VPC→subnet→SG→instance. → **Kalau gagal:** ulang section 2.1. → **Catat notes:** ini kebiasaan yang cegah salah pasang.

---

## Hari 2 · Praktik (2 jam)

1. **`tofu init && tofu validate`** → **Yang harus terjadi:** valid. → **Kalau gagal:** cek `required_providers`. → **Catat notes:** output validate.
2. **`tofu plan -out=tfplan`** → **Yang harus terjadi:** "Plan: N to add". → **Kalau gagal:** cek quota/region. → **Catat notes:** simpan tfplan.
3. **`tofu apply tfplan`** → **Yang harus terjadi:** "Apply complete!". → **Kalau gagal:** baca error (permission/quota). → **Catat notes:** `tofu output` (vpc_id, private_subnet_id, app_instance_id).
4. **Gambar diagram jaringan** di `evidence/` → **Yang harus terjadi:** sesuai config aktual. → **Kalau gagal:** bandingkan dengan `tofu output`. → **Catat notes:** beda tebakan vs aktual.
5. **Verifikasi SG db:** **`aws ec2 describe-security-groups --group-ids <db-sg>`** → **Yang harus terjadi:** `UserIdGroupPairs` isi SG app, nggak ada `0.0.0.0/0`. → **Kalau gagal:** edit `main.tf`, `tofu apply`. → **Catat notes:** bukti least exposure.

---

## Hari 3 · Eksperimen (2 jam) — sengaja rusak

1. **Bukti DB tertutup:** **`nmap -Pn -p 5432 <public-ip-app>`** → **Yang harus terjadi:** `filtered/closed`. → **Kalau gagal:** SG db kebuka publik, perbaiki. → **Catat notes:** ini acceptance #1.
2. **Bukti LB kebuka:** **`curl -fsS -o /dev/null -w '%{http_code}\n' https://lab.domain-lo`** → **Yang harus terjadi:** `200`. → **Kalau gagal:** cek SG lb + DNS. → **Catat notes.**
3. **Cek TLS:** **`openssl s_client -connect lab.domain-lo:443 -servername lab.domain-lo -brief`** → **Yang harus terjadi:** `Verify return code: 0`. → **Kalau gagal:** cek sertifikat. → **Catat notes.**
4. **Admin tanpa SSH:** **`aws ssm start-session --target i-0abc`** → **Yang harus terjadi:** session masuk tanpa port 22. → **Kalau gagal:** cek SSM agent + role. → **Catat notes:** bukti zero standing access.
5. **Cek IMDSv2:** di dalam instance **`curl -s http://169.254.169.254/latest/meta-data/`** → **Yang harus terjadi:** 401 (butuh token) karena `http_tokens=required`. → **Kalau gagal:** set metadata_options. → **Catat notes.**
6. **Verifikasi dari luar:** **`nmap -Pn -p 22,80,443,5432,6379 <ip>`** → **Yang harus terjadi:** cuma 80/443 open. → **Kalau gagal:** tutup port lain. → **Catat notes:** attack surface terbatas.

🧹 **HARI 3 WAJIB — CLEANUP (lab disposable):**

```bash
tofu destroy
# Yang harus terjadi: "Destroy complete! Resources: N destroyed."
# SETELAH destroy: cek console AWS untuk sisa NAT gateway, EIP, volume — itu yang sering ditagih.
```

---

## Hari 4 · Review (2 jam)

1. **Tulis** `docs/security-review.md` (attack surface, hak akses, eksposur jaringan, biaya). → **Yang harus terjadi:** tabel terisi. → **Kalau gagal:** lihat template `docs/security-review.md`.
2. **Isi** temuan: IMDSv2, SG rujuk SG, private subnet, budget alert. → **Yang harus terjadi:** status "ya".
3. **Verifikasi** nggak ada `*.tfstate` ke-commit. → **Yang harus terjadi:** `git status` bersih.

---

## Cheat sheet per hari

| Hari | Fokus | Perintah inti |
| --- | --- | --- |
| 1 | Teori | bedah `main.tf`, gambar diagram |
| 2 | Praktik | `tofu apply tfplan`, `describe-security-groups` |
| 3 | Rusak | `nmap`, `curl`, `ssm start-session`, `destroy` |
| 4 | Review | isi `docs/security-review.md`, cek `.gitignore` |

## Rubrik observasi per hari

| Hari | Yang lo amati | Tanda lulus |
| --- | --- | --- |
| 1 | Paham VPC/SG/IAM | Bisa gambar diagram tanpa bantuan |
| 2 | Resource jalan rapi | `apply` sukses, output muncul |
| 3 | Kegagalan keamanan terbukti | DB tertutup, IMDSv2 aktif |
| 4 | Review lengkap | `docs/security-review.md` terisi |

## Troubleshooting cepat

| Error | Penyebab | Solusi |
| --- | --- | --- |
| `nmap` buka 5432 | SG db `0.0.0.0/0` | ganti ke SG app |
| app nggak reachable LB | SG app nggak rujuk SG lb | tambah ingress dari SG lb |
| app nggak akses S3 | role/profile salah | pasang `iam_instance_profile` |
| IMDSv1 masih jalan | `http_tokens` bukan required | set metadata_options |
| SSH masih bisa | port 22 publik | tutup, pakai SSM |
| tagihan meledak | NAT ditinggal nyala | `tofu destroy`, matikan NAT |

## Alat bantu wajib

| Alat | Kegunaan | Cara cek |
| --- | --- | --- |
| `tofu` / `terraform` | eksekusi IaC | `tofu version` |
| `aws` CLI | cek SG, IMDS, budget | `aws --version` |
| `nmap` | bukti port tertutup | `nmap --version` |
| `openssl` | cek TLS | `openssl version` |

## Langkah pemulihan darurat

- SG kebuka publik: edit `main.tf`, `tofu apply`, lalu `nmap` ulang buat bukti.
- IMDSv2 belum aktif: set `metadata_options.http_tokens=required`, `tofu apply`.
- Tagihan tak terduga: `tofu destroy` lalu cek console untuk NAT/EIP/volume sisa.

## Verifikasi akhir

- [ ] `git status` bersih dari `*.tfstate`.
- [ ] Budget alert tercatat aktif.
- [ ] Bukti `nmap`/`curl`/`ssm` ada di `evidence/`.

## Daftar periksa

- [ ] Budget alert aktif SEBELUM apply.
- [ ] `nmap` 5432 = filtered; curl 443 = 200.
- [ ] IMDSv2 `http_tokens=required` (curl 169.254... → 401).
- [ ] Admin lewat SSM, bukan SSH publik.
- [ ] Diagram jaringan sesuai `tofu output`.
- [ ] `tofu destroy` selesai + cek sisa resource di console.

## Cara buktikan

Lo lulus kalau: (1) `nmap -Pn -p 5432` menunjukkan DB tertutup, (2) `curl` LB 200, (3) `curl 169.254.169.254` di instance balikin 401 (IMDSv2 aktif), (4) budget alert tercatat aktif, dan (5) sesudah `tofu destroy` nggak ada resource nyangkut. Bukti di `evidence/` sudah disanitasi.
