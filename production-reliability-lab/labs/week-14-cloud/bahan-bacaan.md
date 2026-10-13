# Minggu 14 · Cloud Networking dan Security — Bahan Bacaan

> *"Jaringan cloud itu kayak sistem perpipaan gedung: kalau lo salah pasang katup, yang bocor bukan cuma air, tapi data pelanggan."* — prinsip minggu ini.

---

## 1. Kenapa penting

Waktu lo deploy aplikasi di laptop, "jaringan" cuma localhost dan port. Di cloud, batasnya nyata: ada yang boleh ngelihat service lo dari internet, ada yang nggak. Salah pasang security group, database lo bisa kebuka dari seluruh dunia dalam hitungan menit — dan bot scanner bakal nemuinya sebelum lo tidur. Ini bukan skenario teori: ribuan MongoDB/Postgres tanpa auth pernah ke-scrape gara-gara `0.0.0.0/0`.

Kenapa gue tekankan **pilih satu provider**? Karena minggu ini tujuannya bukan hapal nama produk (VPC, subnet, SG), tapi paham **pola**: public vs private, ingress vs egress, identity vs access. Pola ini sama di AWS, GCP, maupun Azure. Kalau lo kuasai polanya di satu tempat, pindah provider cuma soal ganti nama command.

Analognya: VPC itu **gedung/komplek**. Subnet public itu ruangan di lobi yang bisa diketuk dari jalan raya (load balancer, 80/443). Subnet private itu ruang brankas di belakang, nggak punya pintu ke jalan — cuma bisa diakses lewat lobi (SG merujuk SG) atau lewat telepon internal (NAT ke internet). Security group itu **daftar tamu di tiap pintu**: "yang boleh masuk lewat pintu 8080 cuma satpam (load balancer), bukan sembarang orang".

---

## 2. Konsep inti

### 2.1 VPC, subnet, route table

VPC (CIDR `10.20.0.0/16`) adalah batas isolasi jaringan lo. Di dalamnya ada subnet: **public** (ada rute ke Internet Gateway, `map_public_ip_on_launch=true`) dan **private** (nggak punya IP publik, egress lewat NAT). Route table nentuin kemana traffic pergi.

```hcl
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.20.0.0/24"
  map_public_ip_on_launch = true
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}
```

### 2.2 Security group: referensi SG, bukan `0.0.0.0/0`

SG itu firewall stateful per resource. Aturan emas: app cuma nerima dari SG load balancer, DB cuma nerima dari SG app. **Jangan** buka database ke `0.0.0.0/0`.

```hcl
resource "aws_security_group" "app" {
  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.lb.id]   # dari LB, BUKAN 0.0.0.0/0
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]                    # keluar lewat NAT
  }
}

resource "aws_security_group" "db" {
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]   # cuma dari app
  }
}
```

### 2.3 IAM least privilege: role, bukan access key

Jangan taruh access key permanen di instance. Pakai **IAM role** yang melekat ke instance (instance profile). Policy dipersempit ke resource spesifik, bukan `"Action":"*"`, `"Resource":"*"`.

```hcl
resource "aws_iam_role" "app" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "app_backup" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:PutObject","s3:GetObject","s3:ListBucket"]
      Resource = ["arn:aws:s3:::sre-lab-backup/*"]   # scope sempit, bukan "*"
    }]
  })
}
```

### 2.4 IMDSv2: tutup celah SSRF

IMDS (Instance Metadata Service) itu endpoint `169.254.169.254` di dalam instance buat ngambil kredensial role. Kalau app lo kena SSRF, attacker bisa njenguk IMDS dan curi token role. **IMDSv2 wajibkan token** (`http_tokens = "required"`) sehingga butuh request PUT dulu — SSRF simpel nggak bisa.

```hcl
resource "aws_instance" "app" {
  metadata_options {
    http_tokens   = "required"   # IMDSv2
    http_endpoint = "enabled"
  }
}
```

### 2.5 Load balancer & akses admin (SSM)

Satu-satunya pintu publik: load balancer (80/443). Admin nggak SSH pakai public IP — pakai **SSM Session Manager** (zero standing access, token umur pendek). `nmap` harus nunjukin cuma 80/443 yang terbuka.

### 2.6 Cost guardrail: budget alert

Ini minggu yang paling gampang bikin tagihan. **NAT Gateway (~$0,045/jam + data)** adalah tersangka utama. AWS Free Tier berubah **15 Juli 2025**: kredit $100 (+$100) berlaku 6 bulan, bukan 12 bulan gratis. Pasang budget alert SEBELUM deploy.

---

### 2.7 Diagram dulu, konfigurasi kemudian

Kebiasaan paling murah yang cegah insiden lab: **gambar diagram jaringan SEBELUM nulis satu baris HCL**. Banyak orang langsung `apply` karena "nanti juga jalan", lalu kaget pas DB kebuka publik atau app nggak bisa ngomong ke DB. Diagram nggak perlu rapi — cukup urutannya: internet → load balancer (80/443) → subnet public (NAT) → subnet private → app → DB/Redis. Kalau lo bisa gambarnya, lo bisa nulis `aws_security_group`-nya, karena tiap panah di diagram = satu aturan SG atau route table.

Blast radius juga keputusan diagram: seberapa jauh kerusakan kalau satu resource bocor? Kalau app ada di public subnet, satu miskonfigurasi = seluruh instance kebuka. Kalau di private + cuma LB yang publik, yang bisa diserang cuma LB. Prinsipnya: **kecilkan permukaan yang menghadap internet sebisa mungkin**, dan biarkan sisanya ngomong lewat referensi SG, bukan CIDR lebar.

## 3. Mekanisme di balik layar

Gimana traffic sampai ke app lo, step by step:

1. **User** → HTTPS ke load balancer (publik, 443).
2. **LB** ngecek SG-nya: dari internet oke, lalu forward ke app di **private subnet** lewat SG `app` (yang cuma nerima dari SG `lb`).
3. **App** butuh kredensial S3 → panggil **IMDSv2** (PUT token dulu, baru GET), dapet role token, akses S3.
4. **App** query **DB** di private subnet; DB SG cuma buka 5432 dari SG `app`.
5. **App** butuh update/ambail data eksternal → lewat **NAT gateway** (egress), bukan IP publik.
6. **SSM** buat admin: session di-router lewat endpoint AWS, nggak butuh port 22 publik.

Tiap hop dicek oleh SG masing-masing — itu sebabnya "reference SG, not CIDR" itu penting: saat app nambah node, SG otomatis ikut, nggak perlu edit CIDR manual.

---

## 4. Glosarium mini

| Istilah | Arti singkat |
| --- | --- |
| VPC | Virtual network terisolasi milik lo di cloud |
| Subnet | Pembagian IP di dalam VPC (public/private) |
| Route table | Aturan kemana traffic diarahkan (ke IGW/NAT/dll) |
| Internet Gateway | Pintu VPC ke internet publik |
| NAT Gateway | Jalan keluar privat ke internet (tanpa IP publik) |
| Security Group | Firewall stateful per resource (allow list) |
| IAM Role | Identitas sementara untuk service, bukan user permanen |
| IMDSv2 | Metadata instance ber-token, cegah SSRF ambil kredensial |
| Least privilege | Kasih izin sesedikit mungkin, cuma yang dibutuhkan |
| SSM | Session Manager: akses admin tanpa SSH publik |

---

## 5. Bahan bacaan wajib

1. **AWS VPC User Guide** — https://docs.aws.amazon.com/vpc/latest/userguide/what-is-amazon-vpc.html
2. **AWS IAM User Guide** — https://docs.aws.amazon.com/IAM/latest/UserGuide/introduction.html
3. **AWS Budgets** — https://docs.aws.amazon.com/cost-management/latest/userguide/budgets-managing-costs.html
4. **Terraform AWS Provider** — https://developer.hashicorp.com/terraform/language (referensi HCL/resource)
5. **OpenTofu Docs** — https://opentofu.org/ (alternatif IaC open source)
6. **SRE Book: Postmortem Culture** — https://sre.google/sre-book/postmortem-culture/ (disiplin review insiden)

---

## 6. Self-check

1. Bedanya subnet public dan private dari sisi route table dan IP publik?
2. Kenapa SG app harus rujuk SG load balancer, bukan `0.0.0.0/0`?
3. Apa itu IAM role + instance profile, dan kenapa lebih aman dari access key?
4. IMDSv2 cegah serangan apa, dan gimana mekanismenya?
5. Kenapa NAT Gateway sering jadi biaya tak terduga di lab?
6. Kapan lo pakai SSM daripada SSH publik, dan kenapa?
7. Fakta penting soal AWS Free Tier per 15 Juli 2025 — apa bedanya sama dulu?

---

## 7. Yang sering salah paham

1. **"Biar gampang, taruh app di public subnet."** Berarti app punya IP publik → surface serangan gede. Pakai private + SSM/NAT.
2. **"IAM `Action:*` pas debugging, nanti dibalikin."** Nanti = lupa. Selalu tulis policy sempit dari awal.
3. **"Budget alert nanti aja pas udah jalan."** Waktu lo sadar, tagihan udah jalan. Pasang SEBELUM apply.
4. **"NAT gateway diemin seminggu nggak apa-apa."** ~$0,045/jam × 168 jam ≈ $7,5 + data. Matikan kalau nggak dipakai.
5. **"Security group `0.0.0.0/0` ke 5432 cuma lab."** Bot scanner nggak peduli itu lab. Jangan.
6. **"IMDSv1 masih aman."** SSRF bisa langsung ambil token. Wajibkan IMDSv2 (`http_tokens=required`).
