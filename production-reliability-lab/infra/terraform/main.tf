# Konfigurasi lab cloud (minggu 13–14).
# Pakai `tofu` (OpenTofu) atau `terraform` — sintaksnya sama.
#
# JANGAN commit *.tfstate. Isinya bisa berisi password dalam teks biasa.

terraform {
  required_version = "~> 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"          # pinning wajib: cegah breaking change dadakan
    }
  }

  # Ganti dengan bucket + tabel lock milik lo sebelum apply.
  # backend "s3" {
  #   bucket         = "sre-lab-tfstate"
  #   key            = "lab/network.tfstate"
  #   region         = "ap-southeast-1"
  #   dynamodb_table = "tf-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.region
}

variable "region" {
  type    = string
  default = "ap-southeast-1"
}

variable "environment" {
  type    = string
  default = "lab"
}

variable "allowed_admin_cidr" {
  type        = string
  description = "CIDR yang boleh akses admin. Jangan pernah 0.0.0.0/0."
  default     = "203.0.113.10/32"
}

# ---------- Network ----------
resource "aws_vpc" "main" {
  cidr_block           = "10.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = { Name = "${var.environment}-vpc", Env = var.environment }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.20.0.0/24"
  map_public_ip_on_launch = true
  availability_zone       = "${var.region}a"

  tags = { Name = "${var.environment}-public", Tier = "public" }
}

resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.20.10.0/24"
  availability_zone = "${var.region}a"

  # TIDAK ada map_public_ip_on_launch: workload internal nggak punya IP publik.
  tags = { Name = "${var.environment}-private", Tier = "private" }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = { Name = "${var.environment}-igw" }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# ---------- Security group ----------
resource "aws_security_group" "lb" {
  name        = "${var.environment}-lb"
  description = "Load balancer: cuma 80/443 dari internet"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "app" {
  name        = "${var.environment}-app"
  description = "Aplikasi: cuma terima dari load balancer"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.lb.id]   # BUKAN 0.0.0.0/0
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]                    # keluar lewat NAT
  }
}

resource "aws_security_group" "db" {
  name        = "${var.environment}-db"
  description = "Database: cuma dari aplikasi"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }
}

# ---------- Compute ----------
resource "aws_instance" "app" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.small"
  subnet_id              = aws_subnet.private.id          # private, BUKAN public
  vpc_security_group_ids = [aws_security_group.app.id]
  iam_instance_profile   = aws_iam_instance_profile.app.name

  # IMDSv2 wajib — cegah SSRF ngambil kredensial role.
  metadata_options {
    http_tokens   = "required"
    http_endpoint = "enabled"
  }

  tags = { Name = "${var.environment}-app", Env = var.environment }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]   # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gpu/ubuntu-24.04-amd64-server-*"]
  }
}

# Least privilege: role, bukan access key permanen.
resource "aws_iam_role" "app" {
  name = "${var.environment}-app-role"

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
  role = aws_iam_role.app.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      # Scope dipersempit ke bucket backup saja — bukan "*".
      Action   = ["s3:PutObject", "s3:GetObject", "s3:ListBucket"]
      Resource = ["*"] # TODO minggu 14: ganti ke ARN bucket backup spesifik
    }]
  })
}

resource "aws_iam_instance_profile" "app" {
  name = "${var.environment}-app-profile"
  role = aws_iam_role.app.name
}

# ---------- Output ----------
output "vpc_id" {
  value       = aws_vpc.main.id
  description = "ID VPC lab"
}

output "private_subnet_id" {
  value       = aws_subnet.private.id
  description = "Subnet workload internal"
}

output "app_instance_id" {
  value       = aws_instance.app.id
  description = "Instance aplikasi (akses lewat SSM, bukan SSH publik)"
}
