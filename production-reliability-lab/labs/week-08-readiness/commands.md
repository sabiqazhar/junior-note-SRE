# Perintah — Minggu 8 · Production Readiness & Reliability Baseline

Format: perintah → output yang diharapkan (contoh, disanitasi) → interpretasi.

## Hari 2 (kondisi normal)

```bash
docker compose config
```
→ (dump konfigurasi valid, tanpa error)
→ Validasi compose sebelum build. Kalau error, `up` bakal gagal.

```bash
docker compose up -d --build --wait
```
→ `Container api  Healthy`, `Container postgres  Healthy`, `Container redis  Healthy`, `Container nginx  Started`
→ Semua naik dengan urutan benar (dependency healthy dulu).

```bash
docker compose ps
```
→ `NAME` / `STATUS` semua `healthy` / `running`
→ Bukti sistem jalan dari repo bersih.

```bash
curl -fsS http://localhost/health && curl -fsS http://localhost/readyz
```
→ (keluar 0, tidak ada output error) artinya dua-duanya `200`
→ `-f` membuat curl gagal kalau status >= 400 → cocok buat health check.

## Hari 3 (eksperimen kegagalan / recovery)

```bash
sudo reboot
```
→ mesin restart. Tunggu sampai bisa SSH/login lagi.

```bash
date +%s; curl -fsS http://localhost/readyz; date +%s
```
→ dua angka epoch; selisihnya = RTO lab (mis. `1717000000` ... `1717000045` → 45 detik)
→ Bukti sistem pulih sendiri tanpa campur tangan.

```bash
docker volume ls
```
→ `production-reliability-lab_pgdata` (atau nama volume Postgres)
→ Volume persist → data nggak hilang saat reboot.

```bash
docker inspect -f '{{ .HostConfig.RestartPolicy.Name }}' production-reliability-lab-api-1
```
→ `unless-stopped` (atau `always`)
→ Bukti container auto-restart. Kalau `no` → nggak bakal pulih otomatis.

```bash
docker compose down && docker compose up -d --wait
```
→ (alternatif reboot di environment yang nggak boleh reboot)
→ Simulasi mati total lalu balik.
