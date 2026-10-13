# Ringkasan — Minggu 14 · Cloud Networking dan Security

> Satu halaman untuk dibaca ulang sebelum ujian/wawancara. Bukan pengganti bahan-bacaan.md.

## Inti 5 kalimat

1. Di cloud, batas jaringan itu nyata: public subnet buat pintu (LB 80/443), private subnet buat brankas (app/DB tanpa IP publik).
2. Security group harus **merujuk SG lain, bukan `0.0.0.0/0`** — database nggak boleh kebuka dari internet.
3. Akses admin pakai **IAM role + SSM**, bukan access key permanen atau SSH publik (zero standing access).
4. **IMDSv2 wajib** (`http_tokens=required`) buat tutup celah SSRF curi kredensial role.
5. Lab cloud bikin tagihan nyata — **NAT Gateway** dan lupa `destroy` adalah tersangka utama; pasang budget alert SEBELUM apply.

## Peta konsep

```
Internet ──443──> [Load Balancer]  (SG lb: 80/443 dari 0.0.0.0/0)
                       │
            Public subnet (NAT GW keluar)
                       │
            Private subnet ──8080(SG app ← SG lb)──> App (Go)
                                                       │ ├─5432 (SG db ← SG app)──> Postgres
                                                       │ ├─6379 (SG app)──> Redis
                                                       │ └─IMDSv2 (token)──> IAM role ─> S3
            Admin: aws ssm start-session (bukan port 22 publik)
```

## Perintah kunci + output

```bash
# 1. Build infra
$ tofu apply tfplan
# output: Apply complete! Resources: N added.
# interpretasi: jaringan lab jadi.

# 2. Buktikan port DB tertutup dari luar
$ nmap -Pn -p 5432 <public-ip-app>
# output: 5432/tcp filtered
# interpretasi: database nggak reachable dari internet. BENAR.

# 3. Buktikan LB kebuka
$ curl -fsS -o /dev/null -w '%{http_code}\n' https://lab.domain-lo
# output: 200
# interpretasi: pintu publik jalan.

# 4. Cek DNS
$ dig +short lab.domain-lo
# output: 52.x.x.x
# interpretasi: A record menunjuk ke LB.

# 5. Cek TLS
$ openssl s_client -connect lab.domain-lo:443 -servername lab.domain-lo -brief
# output: Verify return code: 0 (ok)
# interpretasi: sertifikat valid.

# 6. Admin tanpa SSH
$ aws ssm start-session --target i-0abc
# output: Starting session with SessionId: ...
# interpretasi: akses tanpa port 22 publik (zero standing access).

# 7. Cek IMDSv2 aktif
$ aws ec2 describe-instances --instance-ids i-0abc \
    --query 'Reservations[].Instances[].MetadataOptions.HttpTokens'
# output: "required"
# interpretasi: IMDSv2 wajib, SSRF nggak bisa ambil token.

# 8. Di instance: IMDSv1 tanpa token harus 401
$ curl -s -o /dev/null -w '%{http_code}\n' http://169.254.169.254/latest/meta-data/
# output: 401
# interpretasi: butuh PUT token dulu = IMDSv2 bekerja.

# 9. Lihat SG db
$ aws ec2 describe-security-groups --group-ids <db-sg>
# output: UserIdGroupPairs: [{GroupId: sg-app}]
# interpretasi: DB cuma buka dari SG app, bukan 0.0.0.0/0.

# 10. Budget alert
$ aws budgets describe-budgets --account-id $ACC --query 'Budgets[].BudgetName'
# output: ["sre-lab-budget"]
# interpretasi: guardrail aktif.

# 11. Cek tagihan awal
$ aws ce get-cost-and-usage --time-period Start=2025-08-01,End=2025-08-02 \
    --granularity DAILY --metrics UnblendedCost
# output: UnblendedCost { Amount: "0.05", Unit: "USD" }
# interpretasi: kecil, tapi naik kalau NAT ditinggal nyala.

# 12. CLEANUP wajib
$ tofu destroy
# output: Destroy complete! Resources: N destroyed.
# interpretasi: semua lab hilang; cek sisa NAT/EIP/volume.
```

## Failure matrix

| Gejala | Penyebab | Cek | Penyelesaian |
| --- | --- | --- | --- |
| `nmap` buka 5432 | SG db `0.0.0.0/0` | review SG db ingress | ganti ke SG app |
| App nggak reachable LB | SG app nggak rujuk SG lb | `describe-security-groups` | tambah ingress dari SG lb |
| App nggak bisa akses S3 | role/instance profile salah | `sts get-caller-identity` di instance | pasang `iam_instance_profile` |
| IMDSv1 masih jalan | `http_tokens` bukan required | `describe-instances` | set `metadata_options.http_tokens=required` |
| Tagihan meledak | NAT ditinggal nyala | AWS Cost Explorer | `tofu destroy`, matikan NAT |
| SSH masih bisa | port 22 dibuka publik | `nmap -Pn -p 22` | tutup, pakai SSM |

## Jebakan

- Node di public subnet "biar gampang SSH".
- IAM `Action:"*"/Resource:"*"` pas debug, lupa dibalikin.
- Nggak pasang budget alert sebelum deploy.
- NAT gateway ditinggal nyala seminggu.
- SG inbound `0.0.0.0/0` ke port database.

## Cheat sheet cepat

| Mau ngapain | Perintah |
| --- | --- |
| Build infra | `tofu apply tfplan` |
| Bukti DB tertutup | `nmap -Pn -p 5432 <ip>` |
| Bukti LB kebuka | `curl -fsS ... lab.domain` |
| Admin aman | `aws ssm start-session --target i-0abc` |
| Cek IMDSv2 | `describe-instances ... HttpTokens` |
| Budget | `aws budgets create-budget ...` |
| Cek tagihan | `aws ce get-cost-and-usage ...` |
| Bersih | `tofu destroy` |

## Tanya jawab kilat

- **Q: Kenapa SG rujuk SG, bukan CIDR?** A: pas app nambah node, SG otomatis ikut.
- **Q: IMDSv2 cegah apa?** A: SSRF ambil token role.
- **Q: Free Tier 2025?** A: kredit $100 (+$100) / 6 bulan, bukan 12 bulan gratis.
- **Q: Akses admin?** A: role + SSM, bukan SSH publik/access key.
- **Q: Tersangka tagihan?** A: NAT Gateway + lupa destroy.

## Acceptance

1. Cuma port yang dibutuhkan (80/443) yang kebuka dari luar; DB tertutup.
2. Akses admin via role + least privilege, bukan user permanen.
3. Diagram jaringan sesuai konfigurasi aktual (bukan tebakan).
4. Budget alert aktif di ≤ 50% batas nyaman.

## Studi kasus mini

- **Kasus A:** app ditaruh di public subnet "biar gampang SSH". Bot scanner nemuin port 22 + app expose → kompromi. Pelajaran: private subnet + SSM.
- **Kasus B:** SG db dibuka `0.0.0.0/0` buat "debug cepat". DB ke-scrape semalam. Pelajaran: SG rujuk SG app, bukan CIDR lebar.
- **Kasus C:** IAM `Action:"*"` pas debug, lupa dibalikin. Eksploitasi lateral terjadi. Pelajaran: tulis policy sempit dari awal.

## Istilah kunci (flashcard)

- **VPC** = jaringan terisolasi milik lo.
- **Subnet public** = ada rute ke IGW, dapat IP publik.
- **Subnet private** = tanpa IP publik, egress lewat NAT.
- **SG** = firewall stateful per resource.
- **IMDSv2** = metadata ber-token, cegah SSRF.
- **IAM role** = identitas sementara, bukan access key.
- **Least privilege** = izin sesedikit mungkin.
- **SSM** = admin tanpa SSH publik.

## Pertanyaan wawancara

1. Kenapa SG rujuk SG, bukan CIDR?
2. Apa celah yang ditutup IMDSv2?
3. Kenapa Free Tier 2025 beda?
4. Kapan pakai SSM vs SSH?
5. Tersangka tagihan lab?

## Pola yang sering kelewat

- NAT gateway ditinggal nyala.
- Budget alert dipasang sesudah, bukan sebelum.
- Port 22 dibuka publik.
- `0.0.0.0/0` ke port DB.

## Cek cepat sebelum apply

| Cek | Perintah |
| --- | --- |
| DB tertutup | `nmap -Pn -p 5432 <ip>` |
| LB kebuka | `curl ... lab.domain` |
| IMDSv2 | `describe-instances ... HttpTokens` |
| Budget | `aws budgets describe-budgets` |

## Satu kalimat

> Benteng cloud itu bertingkat: public cuma buat LB, private buat brankas, SG rujuk SG, dan budget alert dipasang sebelum, bukan sesudah, tombol apply ditekan.
