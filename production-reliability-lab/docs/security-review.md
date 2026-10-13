# Security Review Lab

Review sederhana, diisi minggu 14, direvisi minggu 16. Bukan audit formal — tujuannya membiasakan diri memeriksa.

## 1. Attack surface

| Yang terbuka ke internet | Port | Perlu? | Catatan |
| --- | --- | --- | --- |
| Nginx / load balancer | 80, 443 | Ya | Satu-satunya pintu masuk |
| SSH ke app server | 22 | Tidak | Pakai SSM / bastion, bukan public IP |
| PostgreSQL | 5432 | Tidak | Cuma dari security group aplikasi |
| Redis | 6379 | Tidak | Cuma dari aplikasi |
| Prometheus / Grafana / Loki | 9090, 3000, 3100 | Tidak | Akses lewat port-forward |

Cara verifikasi:

```bash
nmap -Pn -p 22,80,443,5432,6379,9090,3000 <ip-publik>
# yang seharusnya terbuka cuma 80/443
```

## 2. Hak akses

| Prinsip | Penerapan di lab | Status |
| --- | --- | --- |
| Least privilege | IAM role dengan policy sempit, bukan access key permanen | |
| Tidak ada kredensial statis | OIDC dari CI ke cloud (token umur pendek) | |
| Zero standing access | SSM Session Manager, bukan SSH key yang menempel terus | |
| Secret tidak di repo | SOPS/age atau secret manager; `.env` di `.gitignore` | |
| Secret tidak di log | Pastikan error handler tidak mencetak DSN/token | |

## 3. Eksposur jaringan

- [ ] Workload internal di **private subnet**, tanpa public IP
- [ ] Egress lewat NAT, bukan public IP per instance
- [ ] IMDSv2 diwajibkan (`HttpTokens: required`)
- [ ] Security group merujuk ke security group lain, bukan `0.0.0.0/0`
- [ ] Route table private subnet tidak punya rute ke internet gateway

## 4. Data & backup

- [ ] Backup terenkripsi (cipher pass dari secret, bukan plaintext)
- [ ] Backup terpisah dari volume database
- [ ] Object storage tidak publik; akses lewat role
- [ ] Restore pernah diuji (lihat `docs/disaster-recovery.md`)
- [ ] Log tidak mengandung data pribadi; output yang di-commit sudah disanitasi

## 5. Supply chain

- [ ] Dependensi di-pin (`go.sum`, provider version)
- [ ] Image hasil build punya attestation (`gh attestation verify`)
- [ ] Image di-scan (`trivy image --severity HIGH,CRITICAL`)
- [ ] IaC di-scan (`trivy config .`)
- [ ] Tidak ada `latest` di deployment

## 6. Biaya operasional

**Fakta penting:** AWS Free Tier berubah 15 Juli 2025 — bukan lagi 12 bulan gratis, melainkan kredit $100 (+$100) yang berlaku 6 bulan.

| Sumber biaya | Perkiraan | Cara mengendalikan |
| --- | --- | --- |
| NAT Gateway | ~$0,045/jam + data processing | Tersangka utama tagihan lab; matikan saat tidak dipakai |
| Instance/NAT node | per jam | Matikan di luar jam lab |
| Managed database | per jam + storage | Pilih kelas terkecil |
| Object storage | per GB | Retensi backup jangan berlebihan |
| Egress | per GB | Jangan tarik backup besar melintasi region |

- [ ] Budget alert aktif di ≤ 50% dari batas yang bisa ditoleransi
- [ ] Alarm terpisah untuk "pengeluaran naik tajam"
- [ ] Rutinitas `terraform destroy` setelah lab selesai

## Temuan & tindakan

| # | Temuan | Tingkat | Tindakan | Status |
| --- | --- | --- | --- | --- |
| 1 | | | | |

Ditulis: ____ · Diperbarui: ____
