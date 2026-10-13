# Perintah — Week 14 · Cloud Networking & Security

Format: perintah → output → interpretasi.

## Hari 2 (kondisi normal)

```bash
# 1. Init + validate
$ tofu init && tofu validate
# output: ...initialized! / Success! The configuration is valid.
# interpretasi: provider siap, HCL valid.

# 2. Plan tersimpan
$ tofu plan -out=tfplan
# output: Plan: 12 to add, 0 to change, 0 to destroy.
# interpretasi: 12 resource (VPC, subnet, SG, instance, role).

# 3. Apply
$ tofu apply tfplan
# output: Apply complete! Resources: 12 added.
# interpretasi: jaringan lab jadi.

# 4. Output penting
$ tofu output private_subnet_id
# output: "subnet-0xyz"
# interpretasi: subnet tempat app (tanpa IP publik).

# 5. Lihat SG db (harus cuma dari SG app)
$ aws ec2 describe-security-groups --group-ids $(tofu output -raw db_sg_id)
# output: ...UserIdGroupPairs: [{GroupId: sg-app}]  (tanpa 0.0.0.0/0)
# interpretasi: database terisolasi dari internet.
```

## Hari 3 (eksperimen kegagalan)

```bash
# 6. Bukti port DB tertutup dari luar
$ nmap -Pn -p 5432 <public-ip-app>
# output: 5432/tcp filtered
# interpretasi: nggak reachable dari internet. BENAR.

# 7. Bukti LB kebuka
$ curl -fsS -o /dev/null -w '%{http_code}\n' https://lab.domain-lo
# output: 200
# interpretasi: satu-satunya pintu publik jalan.

# 8. Cek DNS
$ dig +short lab.domain-lo
# output: 52.x.x.x
# interpretasi: A record menunjuk ke LB.

# 9. Cek TLS
$ openssl s_client -connect lab.domain-lo:443 -servername lab.domain-lo -brief
# output: Verify return code: 0 (ok)
# interpretasi: sertifikat valid.

# 10. Admin via SSM (bukan SSH)
$ aws ssm start-session --target i-0abc
# output: Starting session with SessionId: x-....
# interpretasi: akses tanpa port 22 publik (zero standing access).

# 11. Cek IMDSv2 aktif
$ aws ec2 describe-instances --instance-ids i-0abc \
    --query 'Reservations[].Instances[].MetadataOptions.HttpTokens'
# output: "required"
# interpretasi: IMDSv2 wajib, SSRF nggak bisa ambil token.

# 12. Di dalam instance: IMDSv1 tanpa token harus gagal
$ curl -s -o /dev/null -w '%{http_code}\n' http://169.254.169.254/latest/meta-data/
# output: 401
# interpretasi: butuh PUT token dulu = IMDSv2 bekerja.

# 13. Budget alert
$ aws budgets describe-budgets --account-id $ACC --query 'Budgets[].BudgetName'
# output: ["sre-lab-budget"]
# interpretasi: guardrail aktif.

# 14. Cek tagihan awal
$ aws ce get-cost-and-usage --time-period Start=2025-08-01,End=2025-08-02 \
    --granularity DAILY --metrics UnblendedCost
# output: ...UnblendedCost { Amount: "0.05", Unit: "USD" }...
# interpretasi: kecil, tapi bakal naik kalau NAT ditinggal nyala.

# 15. CLEANUP wajib
$ tofu destroy
# output: Destroy complete! Resources: 12 destroyed.
# interpretasi: semua lab hilang. Cek console buat sisa NAT/EIP/volume.

# 16. Verifikasi nggak ada state ke-commit
$ git status --porcelain | grep tfstate
# output: (kosong)
# interpretasi: aman.
```
