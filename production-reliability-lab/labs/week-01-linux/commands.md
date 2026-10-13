# Perintah — Week 01 · Linux observability & troubleshooting

Format: `perintah` → output yang diharap → interpretasi. Output di bawah sudah disanitasi (PID, IP, waktu diganti placeholder). Jalankan di Ubuntu 22.04/24.04 dengan aplikasi lab (`sre-api`) jalan.

Anggaplah `sre-api` jalan di port `8080` dengan PID contoh `9001`. Ganti dengan PID asli lo (`pgrep -f sre-api`).

---

## Hari 2 (kondisi normal)

**1. Identitas kernel & arsitektur**
```bash
uname -a
```
→ `Linux lab-vm 6.8.0-xx-generic #yy-Ubuntu SMP x86_64 GNU/Linux`
→ Kernel 6.x (Ubuntu 24.04 bawa HWE). Versi kernel penting karena PSI butuh >= 4.20, cgroup v2 butuh >= 4.5.

**2. Load average 1/5/15 menit**
```bash
uptime
```
→ `up 3 days, load average: 0.12, 0.08, 0.05`
→ Angka di bawah jumlah core = sistem idle. Bukan persen CPU.

**3. Memori: lihat `available`, BUKAN `free`**
```bash
free -h
```
→ `Mem: total 7.7Gi, free 280Mi, buff/cache 4.1Gi, available 6.9Gi`
→ `available` 6.9Gi artinya masih lega. `free` kecil itu wajar (Linux pakai RAM buat page cache).

**4. Disk & inode terpisah**
```bash
df -h && df -i
```
→ `Use%` (kapasitas) dan `IUse%` (jumlah file). Keduanya harus <90%.
→ "Disk penuh" bisa jadi inode habis (`df -i` 100%) padahal kapasitas masih 40%.

**5. Proses pemakan CPU teratas**
```bash
ps aux --sort=-%cpu | head -10
```
→ Baris teratas nunjukin COMMAND & %CPU tertinggi, mis. `sre-api  85.0`.
→ Buat ngecek siapa yang lagi "panas" di level proses.

**6. Proses pemakan memori teratas**
```bash
ps aux --sort=-%mem | head -10
```
→ Sama, diurutkan memori. Bedakan VSZ (virtual) vs RSS (fisik, yang makan RAM beneran).

**7. Service sistem yang gagal**
```bash
systemctl --failed
```
→ `0 loaded units listed` → sehat. Kalau ada unit, itu titik mulai investigasi.

**8. Error journal sejak boot**
```bash
journalctl -p err -b --no-pager
```
→ Baris merah = error. Kosong berarti tidak ada error level err sejak boot.

**9. Cari PID aplikasi + command lengkap**
```bash
pgrep -af 'sre-api|go run'
```
→ `9001 /app/sre-api` — ini PID yang lo pakai di langkah berikut.

**10. Siapa yang listen di 8080**
```bash
ss -lntp | grep 8080
```
→ `LISTEN 0 4096 *:8080 users:(("sre-api",pid=9001,fd=3))`
→ Bukti: proses Go lo benar-benar listen, dan punya PID 9001.

**11. File & koneksi yang dibuka proses**
```bash
lsof -p 9001 | head -30
```
→ Baris `IPv4 ... TCP localhost:8080 (LISTEN)`, koneksi ke `5432` (Postgres) & `6379` (Redis).
→ Bukti dependency terhubung: ada socket ke Postgres & Redis.

**12. Koneksi di port 8080 (hakim waktu port bentrok)**
```bash
lsof -i :8080
```
→ Sama dengan `ss` di atas tapi dari sisi koneksi. Kalau ini kosong tapi `ss -lntp` ada, berarti port cuma listen lokal.

**13. Thread & RSS dari /proc**
```bash
cat /proc/9001/status | grep -E 'Threads|VmRSS'
```
→ `Threads: 18` `VmRSS: 28400 kB`
→ 18 thread OS (goroutine Go dipetakan ke sini), pakai ~28 MB RAM fisik.

**14. Jumlah file descriptor**
```bash
ls /proc/9001/fd | wc -l
```
→ `43`
→ Tiap koneksi HTTP + DB + file = 1 fd. Naik terus = bocor koneksi (selidiki di minggu 6).

---

## Hari 3 (eksperimen kegagalan)

**15. Bikin CPU tinggi lalu lihat per-thread**
```bash
yes > /dev/null & BAD=$!
top -H -p $BAD
```
→ `top -H` nunjukin thread `yes` di 100% CPU, bukan cuma proses.
→ `kill -TERM $BAD` lalu `kill -KILL $BAD` kalau bandel.

**16. CPU+mem+IO per proses, interval 1s**
```bash
pidstat -urd -p 9001 1 5
```
→ Baris ke-2 dst: `%cpu`, `minflt/s`, `kbmemfree`. Baris ke-1 itu rata-rata sejak boot.
→ Tunggu baris ke-2 sebelum simpulkan.

**17. Status systemd unit yang sengaja rusak**
```bash
systemctl status sre-lab.service
```
→ `failed (Result: exit-code)` + pesan error path binary.
→ Ini cara bedain "aplikasi error" vs "sistem error".

**18. Log spesifik unit**
```bash
journalctl -u sre-lab.service -n 100 --no-pager
```
→ Pesan `execve: No such file or directory` (path salah).
→ Journal = sumber kebenaran, lebih jujur dari log aplikasi.

**19. Tiga kondisi: proses mati**
```bash
pgrep -f sre-api      # kosong
ss -lntp | grep 8080  # kosong
curl -v --max-time 3 http://localhost:8080/health
```
→ `curl: (7) Failed to connect ... Connection refused`
→ `exit 7` = tidak ada yang listen.

**20. Tiga kondisi: hang (listen tapi nggak balas)**
```bash
curl -v --max-time 3 http://localhost:8080/slow
echo "exit code: $?"
```
→ `curl: (28) Operation timed out` → `exit 28`.
→ `ss` nemu listen, tapi aplikasi nggak balas (biasanya thread pool / DB habis).

**21. Bukti exit code (simpan trik ini)**
```bash
curl -sS -o /dev/null -w 'code=%{http_code} exit=%{exitcode}\n' http://localhost:8080/health
```
→ `code=200 exit=0` (ok), `code=000 exit=7` (refused), `code=000 exit=28` (timeout).
→ `http_code=000` = curl gagal sebelum dapat response HTTP.

**22. PSI: sistem lagi nunggu resource?**
```bash
cat /proc/pressure/memory
```
→ `some avg10=0.00 avg60=0.00 avg300=2.31 total=1234567`
→ `avg300` naik = beberapa proses sering nunggu memori. `full` tinggi = thrashing.

**23. OOM cgroup (bila eksperimen Hari 3.3 jalan)**
```bash
cat /sys/fs/cgroup/memory.events
```
→ `oom 1` `oom_kill 1` → kernel pernah nembak proses di cgroup ini.
→ Bukti OOM tanpa harus bongkar `dmesg`.
