# Lab — Minggu 01 · Linux observability dan troubleshooting

> Ikuti **BERURUTAN**, jangan lompat. Setiap langkah formatnya: perintah → Yang harus terjadi → Kalau gagal.
> Catatan: semua perintah diasumsikan dijalankan dari root repo `production-reliability-lab/`.

---

## Persiapan (10–15 menit)

Ceklist yang harus sudah beres sebelum mulai:

- [ ] Ubuntu 22.04 atau 24.04 (boleh VPS, boleh WSL2, boleh VM lokal). **Jangan** pakai macOS buat minggu ini — `/proc` dan `systemd`-nya beda.
- [ ] Akses `sudo`.
- [ ] Go 1.22+ terpasang.
- [ ] Tool observability terpasang (lihat verifikasi di bawah).

**Verifikasi environment.** Jalankan semua, dan pastikan nggak ada yang error:

```bash
lsb_release -a                 # Ubuntu 22.04.x atau 24.04.x
uname -r                       # kernel 5.15 (22.04) atau 6.x (24.04)
go version                     # go1.22 atau lebih baru
sudo apt-get update -y
sudo apt-get install -y sysstat lsof procps iproute2 curl jq
which pidstat lsof ss curl jq  # keempatnya harus nemu path
```

**Yang harus terjadi:** `go version` keluar `go1.2x.x linux/amd64`, dan `which` ngasih 4 path (mis. `/usr/bin/pidstat`, `/usr/bin/lsof`, `/usr/bin/ss`, `/usr/bin/curl`).

**Kalau gagal:**
- `pidstat: command not found` → `sudo apt-get install -y sysstat`, lalu **aktifkan**: `sudo sed -i 's/ENABLED="false"/ENABLED="true"/' /etc/default/sysstat && sudo systemctl restart sysstat`.
- `ss: command not found` → `sudo apt-get install -y iproute2`.
- Di WSL2, `systemd` mungkin nggak jalan. Cek dengan `systemctl is-system-running`. Kalau outputnya `offline`, pakai VM atau VPS buat Langkah 3.x.

**Cek dulu apakah cgroup v2 aktif** — ini nentuin apakah eksperimen OOM cgroup lo bisa jalan:

```bash
stat -fc %T /sys/fs/cgroup/     # harusnya "cgroup2fs"
cat /sys/fs/cgroup/memory.max   # harusnya keluar angka atau "max"
```

**Kalau gagal:** kalau outputnya `tmpfs`, berarti masih cgroup v1. Tambahin `systemd.unified_cgroup_hierarchy=1` ke boot param, atau pakai Ubuntu 24.04 bersih.

---

## Hari 1 · Teori (2 jam)

### Langkah 1.1 — Baca bahan bacaan sampai selesai

```bash
# Buka file ini dan baca dari atas ke bawah. Serius, jangan discan.
$EDITOR labs/week-01-linux/bahan-bacaan.md
```

**Yang harus terjadi:** lo selesai baca dan bisa jawab 7 pertanyaan di bagian *Self-check* tanpa buka Google.

**Kalau gagal:** kalau masih ada yang nggak bisa lo jawab, baca ulang bagian 2 dan 3 aja. Jangan lanjut ke Hari 2 sebelum *Self-check* nomor 1–4 kejawab.

**Catat di notes.md:** tulis ulang 8 konsep inti pakai kalimat lo sendiri, satu paragraf per konsep. Kalau lo nggak bisa nulis ulang, berarti belum paham.

### Langkah 1.2 — Baca sumber primer (cgroup v2 + PSI)

Buka dua halaman ini, fokus ke bagian yang disebut di tabel `bahan-bacaan.md` §5:

- [Control Group v2 — Memory](https://docs.kernel.org/admin-guide/cgroup-v2.html)
- [PSI — Pressure Stall Information](https://docs.kernel.org/accounting/psi.html)

**Yang harus terjadi:** lo bisa sebutin minimal 4 file antarmuka cgroup (`memory.max`, `memory.current`, `memory.events`, `memory.pressure`) dan bisa jelasin beda `some` vs `full`.

**Kalau gagal:** cari contoh output tiap file di mesin lo sendiri dengan `cat`. Angka nyata lebih gampang nempel daripada teks dokumen.

**Catat di notes.md:** daftar file `/sys/fs/cgroup/memory.*` yang lo temuin di mesin lo, plus arti satu kalimat tiap file.

### Langkah 1.3 — Pahami urutan OOM

Baca bagian 3 (`Mekanisme di balik layar`) bahan bacaan, lalu gambar ulang urutannya di kertas atau ASCII:

```
malloc → overcommit check → charge ke cgroup → lewat memory.max? → reklamasi
  → masih lewat? → OOM cgroup → oom_badness() → skor tertinggi → SIGKILL
```

**Yang harus terjadi:** lo bisa jelasin ke orang lain kenapa OOM bisa muncul *tiba-tiba* padahal memori kelihatannya masih ada.

**Kalau gagal:** ingat satu kalimat ini: *kernel mengiyakan dulu (overcommit), nagih belakangan (OOM).*

**Catat di notes.md:** urutan 6 langkah versi lo.

### Langkah 1.4 — Kenali sinyal lewat man page

```bash
man 7 signal | grep -A4 -E 'SIGTERM|SIGKILL'
# atau kalau man page belum ada:
curl -sS https://man7.org/linux/man-pages/man7/signal.7.html | sed 's/<[^>]*>//g' | grep -E 'SIGTERM|SIGKILL' | head
```

**Yang harus terjadi:** lo nemu bahwa SIGTERM = 15 dan SIGKILL = 9, dan bahwa SIGKILL **tidak bisa di-*catch*, di-*block*, atau di-*ignore***.

**Kalau gagal:** `sudo apt-get install -y manpages manpages-dev` lalu ulangi.

**Catat di notes.md:** nomor + arti SIGTERM, SIGKILL, SIGINT, SIGHUP, SIGQUIT.

### Langkah 1.5 — Kenali journalctl

```bash
journalctl -b --no-pager | head -20
journalctl -k --no-pager | head -10
journalctl -p err -b --no-pager | head -20
```

**Yang harus terjadi:** ketiganya ngeluarin log. `-k` khusus cuma log kernel.

**Kalau gagal:** `journalctl: command not found` → berarti bukan systemd. Atau `No journal files were found` → `sudo mkdir -p /var/log/journal && sudo systemd-tmpfiles --create --prefix /var/log/journal && sudo systemctl restart systemd-journald`.

**Catat di notes.md:** jabaran 5 flag `journalctl` yang paling lo pakai (dari [journalctl(1)](https://man7.org/linux/man-pages/man1/journalctl.1.html)).

---

## Hari 2 · Praktik (2 jam)

### Langkah 2.1 — Baseline kondisi sehat

Sebelum bisa nebak apa yang aneh, lo harus tahu yang normal kayak gimana.

```bash
uptime
free -h
df -h && df -i
vmstat 1 3
cat /proc/pressure/cpu /proc/pressure/memory /proc/pressure/io
```

**Yang harus terjadi:** `load average` di bawah jumlah core, `available` besar, `wa` di `vmstat` mendekati 0, dan PSI `avg300` di bawah 1.00.

**Kalau gagal:** kalau PSI `avg300` sudah di atas 5.00 di mesin nganggur, berarti mesin lo emang lagi ada beban lain. Catat sebagai baseline dan lanjut — nanti lo bandingkan dengan kondisi eksperimen.

**Catat di notes.md:** salin semua output ke `notes.md` di bawah judul `## Baseline (kondisi sehat)`.

### Langkah 2.2 — Build dan jalankan aplikasi lab

```bash
cd /workspace/production-reliability-lab      # sesuaikan path repo lo
go mod tidy                                   # pertama kali aja; bikin go.sum
go build -o /tmp/sre-api ./app
```

Jalanin dengan env dependency palsu. Ini aman karena `sql.Open()`dan `redis.NewClient()` di Go itu **lazy** — mereka nggak nyambung sampai dipakai.

```bash
DATABASE_URL='postgres://lab:lab@127.0.0.1:5432/lab?sslmode=disable' \
REDIS_URL='redis://127.0.0.1:6379' \
LOG_LEVEL=info \
/tmp/sre-api &
API_PID=$!
sleep 1
echo "PID aplikasi: $API_PID"
```

**Yang harus terjadi:** aplikasi nyala tanpa error. Log `{"level":"INFO","msg":"server listening","addr":":8080"}`.

**Kalau gagal:**
- `DATABASE_URL kosong` → env-nya belum kebaca. Pastikan nggak ada spasi sebelum `DATABASE_URL`.
- `go: cannot find main module` → lo belum `cd` ke root repo.
- Port 8080 kepakai → lihat Langkah 2.5.

**Catat di notes.md:** PID aplikasi dan waktu start.

### Langkah 2.3 — Identifikasi proses pakai `/proc`

```bash
pgrep -af sre-api
PID=$(pgrep -f '/tmp/sre-api' | head -1)
cat /proc/$PID/status | grep -E 'Name|Threads|VmRSS|VmSize'
cat /proc/$PID/limits | grep -E 'open files|Max processes'
ls /proc/$PID/fd | wc -l
cat /proc/$PID/cmdline | tr '\0' ' '; echo
cat /proc/$PID/oom_score_adj
```

**Yang harus terjadi:** `Threads` lebih dari 1 (runtime Go bikin beberapa thread OS), `open files` biasanya 1024 atau 1048576, jumlah fd belasan sampai puluhan.

**Kalau gagal:** `pgrep` kosong → aplikasi mati. Cek lagi log di terminal tempat lo jalanin.

**Catat di notes.md:** nilai `Threads`, `VmRSS`, jumlah fd, dan `oom_score_adj`.

### Langkah 2.4 — Buktikan siapa yang listen dan apa yang dipegang

```bash
ss -lntp | grep 8080
lsof -i :8080
lsof -p $PID | head -30
```

**Yang harus terjadi:**

```
LISTEN 0  4096  *:8080  *:*  users:(("sre-api",pid=12345,fd=3))
```

dan `lsof -p $PID` nunjukin socket LISTEN plus file binary-nya.

**Kalau gagal:** `ss` kosong padahal aplikasi jalan → kemungkinan aplikasi bind ke `127.0.0.1` (bukan `*`). Cek `ss -lntp` tanpa grep, lihat kolom ke-4.

**Catat di notes.md:** output `ss -lntp` — **ini bukti wajib** buat acceptance criteria.

### Langkah 2.5 — Bedakan healthz dan readyz

```bash
curl -sS -o /dev/null -w 'healthz=%{http_code}\n' --max-time 3 http://127.0.0.1:8080/healthz
curl -sS -o /dev/null -w 'readyz=%{http_code}\n'  --max-time 3 http://127.0.0.1:8080/readyz
echo "exit healthz: $?"
```

**Yang harus terjadi:** `healthz=200` dan `readyz=503`. Kenapa? Karena Postgres dan Redis nggak beneran ada — dan ini **justru contoh sempurna** dari "proses hidup tapi belum siap".

**Kalau gagal:** kalau dua-duanya `000`, prosesnya nggak listen — balik ke Langkah 2.2.

**Catat di notes.md:** tulis kenapa `healthz` 200 tapi `readyz` 503, dan kapan bedanya ini penting (petunjuk: minggu 8, readiness probe).

### Langkah 2.6 — Kirim SIGTERM, lihat graceful shutdown

```bash
kill -TERM $PID
sleep 1
pgrep -af sre-api || echo "proses sudah keluar"
```

**Yang harus terjadi:** di log muncul `shutdown signal received` lalu `shutdown selesai`. Proses keluar **dengan sendirinya** dalam 1 detik (karena `Shutdown()` langsung balas saat nggak ada request in-flight).

**Kalau gagal:** kalau proses nggak keluar juga, cek apakah `$PID` salah. Jangan langsung `kill -9`.

**Catat di notes.md:** dua baris log shutdown. Ini bukti acceptance criteria nomor 4.

---

## Hari 3 · Eksperimen kegagalan (2 jam)

> **Aturan Hari 3:** setiap langkah WAJIB punya cara balikin ke kondisi normal. Jangan tutup sesi sebelum semua cleanup jalan.

### Langkah 3.1 — CPU tinggi: temukan thread nakalnya dalam 5 menit

```bash
yes > /dev/null &
BAD=$!
sleep 2
top -H -b -n 1 -p $BAD | head -20
pidstat -urd -p $BAD 1 5
ps -eLo pid,tid,pcpu,comm --sort=-pcpu | head -5
```

**Yang harus terjadi:** `%CPU` mendekati 100 untuk thread `yes`. `pidstat` baris ke-2 dst nunjukin `%usr` tinggi.

**Kalau gagal:** `top -H` nggak nunjukin thread → pastikan flag `-H` ada dan jangan pakai `-b` kalau mau interaktif.

**Cara mengembalikan:**
```bash
kill -TERM $BAD
sleep 1
pgrep -af '^yes' || echo "bersih"
# verifikasi load turun:
uptime
```

### Langkah 3.2 — Bikin tekanan memori terukur, lalu amati PSI

Jangan bikin host OOM beneran. Bikin tekanan dengan proses yang alokasi memori perlahan:

```bash
# Alokasi ~256MB dan tahan (butuh python3; kalau nggak ada, pakai `stress`)
python3 -c "
import time
b = bytearray(256 * 1024 * 1024)
time.sleep(60)
" &
HOG=$!
sleep 3
cat /proc/pressure/memory
cat /sys/fs/cgroup/memory.current
free -h
```

**Yang harus terjadi:** `some avg10` di memory PSI naik dari 0, `memory.current` naik, `available` di `free -h` turun.

**Kalau gagal:** kalau `avg10` tetap `0.00`, berarti alokasinya kecepetan selesai atau mesin lo kegedean. Naikin jadi `1024 * 1024 * 1024` atau tambah durasi sleep.

**Cara mengembalikan:**
```bash
kill -TERM $HOG
sleep 2
cat /proc/pressure/memory     # avg300 akan pelan-pelan turun, avg10 turun lebih cepat
free -h
```

### Langkah 3.3 — OOM cgroup yang terkendali (pakai systemd-run)

Ini cara paling aman ngebuktiin OOM tanpa ngerusak host. `systemd-run` bikin cgroup sementara dengan batas memori.

```bash
sudo systemd-run --unit=sre-oom-lab --slice=sre-oom-lab.slice \
  -p MemoryMax=64M -p MemorySwapMax=0 \
  /usr/bin/python3 -c "
import time
b = bytearray(256 * 1024 * 1024)
time.sleep(120)
"
```

Langsung cek buktinya:

```bash
sleep 5
systemctl status sre-oom-lab.service --no-pager
sudo journalctl -u sre-oom-lab.service -n 20 --no-pager
sudo journalctl -k --no-pager | grep -i -E 'out of memory|killed process'
```

Baca counter cgroup-nya (path-nya beda tergantung versi systemd):

```bash
CG=$(find /sys/fs/cgroup -type d -name 'sre-oom-lab*' | head -1)
echo "cgroup: $CG"
cat "$CG/memory.events"
cat "$CG/memory.max"
cat "$CG/memory.current"
```

**Yang harus terjadi:** service berstatus `failed`, dan salah satu dari ini muncul:
- Baris `Memory cgroup out of memory: Killed process ...` di `journalctl -k`, **atau**
- `oom_kill` di `memory.events` bernilai `1` atau lebih.

**Kalau gagal:**
- `systemd-run` nggak nemu `MemoryMax` → systemd terlalu tua, atau cgroup v1. Cek Langkah Persiapan.
- Proses nggak mati → `bytearray` mungkin nggak kena touch semua halaman. Tambahin loop `for i in range(0, len(b), 4096): b[i] = 1`.

**Cara mengembalikan (WAJIB):**
```bash
sudo systemctl stop sre-oom-lab.service 2>/dev/null || true
sudo systemctl reset-failed sre-oom-lab.service 2>/dev/null || true
sudo systemctl stop sre-oom-lab.slice 2>/dev/null || true
systemctl status sre-oom-lab.service --no-pager   # harusnya "not-found" atau "inactive"
```

### Langkah 3.4 — systemd unit yang sengaja rusak

```bash
sudo tee /etc/systemd/system/sre-lab.service > /dev/null <<'EOF'
[Unit]
Description=SRE Lab (sengaja rusak)
After=network.target

[Service]
Type=simple
ExecStart=/tmp/binary-yang-tidak-ada
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl start sre-lab.service
sleep 5
systemctl status sre-lab.service --no-pager
systemctl --failed --no-pager
sudo journalctl -u sre-lab.service -n 30 --no-pager
```

**Yang harus terjadi:** status `failed` dengan `status=203/EXEC` (file nggak ada), dan `systemctl --failed` nyantumin unit ini. Karena `Restart=always`, lo bakal lihat beberapa kali percobaan gagal di journal — ini nunjukin kenapa service bisa "restart terus".

**Kalau gagal:** `systemd is-system-running: offline` → bukan systemd (WSL2 tanpa systemd). Lewati langkah ini, baca outputnya dari teman atau VPS.

**Cara mengembalikan (WAJIB):**
```bash
sudo systemctl stop sre-lab.service
sudo systemctl disable sre-lab.service
sudo rm -f /etc/systemd/system/sre-lab.service
sudo systemctl daemon-reload
sudo systemctl reset-failed
systemctl --failed --no-pager     # harusnya kosong
```

### Langkah 3.5 — Tiga kondisi yang sering ketuker

Bikin tabel perbandingan. Ini artefak yang bakal lo pakai sepanjang karier.

```bash
# Kondisi A: proses mati
kill -TERM $PID 2>/dev/null; sleep 1
pgrep -af '/tmp/sre-api' || echo "A: pgrep kosong"
ss -lntp | grep 8080 || echo "A: ss kosong"
curl -sS -o /dev/null --max-time 3 http://127.0.0.1:8080/healthz; echo "A: exit=$?"

# Kondisi B: proses hidup, port nggak listen (jalanin binary dengan PORT lain lalu cek port 8080)
DATABASE_URL='postgres://x@127.0.0.1:5432/x?sslmode=disable' REDIS_URL='redis://127.0.0.1:6379' \
  APP_PORT=9999 /tmp/sre-api &
B_PID=$!
sleep 1
pgrep -af '/tmp/sre-api' | head -2     # B: ada PID
ss -lntp | grep 8080 || echo "B: port 8080 nggak ada yang listen"
curl -sS -o /dev/null --max-time 3 http://127.0.0.1:8080/healthz; echo "B: exit=$?"
kill -TERM $B_PID

# Kondisi C: listen tapi nggak responsif (endpoint /slow)
DATABASE_URL='postgres://x@127.0.0.1:5432/x?sslmode=disable' REDIS_URL='redis://127.0.0.1:6379' \
  /tmp/sre-api &
C_PID=$!
sleep 1
curl -sS -o /dev/null --max-time 2 "http://127.0.0.1:8080/slow?ms=30000"; echo "C: exit=$?"
kill -TERM $C_PID
```

**Yang harus terjadi:**

| Kondisi | `pgrep` | `ss -lntp` | exit code curl |
| --- | --- | --- | --- |
| A: proses mati | kosong | kosong | `7` (refused) |
| B: hidup, nggak listen di port itu | ada | kosong | `7` (refused) |
| C: listen, nggak responsif | ada | ada | `28` (timeout) |

**Kalau gagal:** exit `7` di kondisi C berarti aplikasi mati sebelum request nyampe. Ulangi, tambah `sleep 2` setelah start.

**Cara mengembalikan:**
```bash
pkill -TERM -f '/tmp/sre-api' 2>/dev/null || true
sleep 1
pgrep -af '/tmp/sre-api' || echo "semua bersih"
```

---

## Hari 4 · Review & dokumentasi (2 jam)

### Langkah 4.1 — Simpan semua bukti ke `evidence/`

```bash
mkdir -p labs/week-01-linux/evidence
{
  echo "=== $(date -u +%FT%TZ) ==="
  echo "--- uname ---"; uname -a
  echo "--- uptime ---"; uptime
  echo "--- free ---"; free -h
  echo "--- PSI ---"; cat /proc/pressure/cpu /proc/pressure/memory /proc/pressure/io
  echo "--- ss -lntp ---"; ss -lntp
  echo "--- memory.events ---"; cat /sys/fs/cgroup/memory.events
} > labs/week-01-linux/evidence/01-baseline.txt
```

**Yang harus terjadi:** file `01-baseline.txt` kebentuk dan isinya lengkap.

**Kalau gagal:** permission denied → pastikan lo `cd` di root repo dan punya akses tulis.

**Catat di notes.md:** path tiap file bukti.

### Langkah 4.2 — Tulis 3 runbook

Bikin tiga file di `labs/week-01-linux/runbooks/`. Masing-masing wajib punya: Gejala, Cara deteksi, Diagnosis (decision tree), Mitigasi, Pencegahan.

```bash
$EDITOR labs/week-01-linux/runbooks/high-cpu.md       # load tinggi, request lambat
$EDITOR labs/week-01-linux/runbooks/oom-symptoms.md   # container restart mendadak
$EDITOR labs/week-01-linux/runbooks/service-down.md   # endpoint nggak bisa diakses
```

Isi minimum:

| File | Yang harus ada |
| --- | --- |
| `high-cpu.md` | Cara nemu thread nakal < 5 menit; keputusan kapan scale vs kapan perbaiki kode |
| `oom-symptoms.md` | Beda kernel `oom-kill` vs `systemd-oomd`; beda OOM host vs OOM cgroup |
| `service-down.md` | Decision tree: proses hidup? port listen? aplikasi responsif? dependency sehat? |

**Yang harus terjadi:** tiga file ada, masing-masing minimal punya 5 bagian wajib.

**Kalau gagal:** kalau lo nggak bisa nulis "cara deteksi" tanpa buka catatan, berarti Hari 2–3 belum nempel. Ulangi Langkah 3.5.

### Langkah 4.3 — Lengkapin `commands.md`

Isi `labs/week-01-linux/commands.md` dengan semua perintah yang lo jalanin Hari 2 dan Hari 3, lengkap dengan output tersanitasi dan interpretasinya. Minimal 15 perintah.

**Yang harus terjadi:** `commands.md` punya dua bagian (`Hari 2` dan `Hari 3`), tiap entri ada perintah + output + interpretasi.

**Kalau gagal:** copy polanya dari `ringkasan.md` bagian "Perintah kunci".

### Langkah 4.4 — Update `README.md`

```bash
$EDITOR labs/week-01-linux/README.md
```

Ganti `> Status: TODO` jadi `> Status: DONE`, dan centang semua kotak di bagian "Yang dikerjakan".

**Yang harus terjadi:** status `DONE` dan 4 kotak ke-centang.

**Kalau gagal:** jangan centang kalau acceptance criteria belum kebukti. Minggu ini belum selesai.

### Langkah 4.5 — Commit

```bash
git add labs/week-01-linux/
git status --short          # pastikan nggak ada file .env atau data sensitif ikut
git commit -m "week-01: linux observability — evidence, runbooks, commands"
```

**Yang harus terjadi:** commit kebentuk, dan `git status --short` sebelum commit **tidak** ngandung file berisi password/IP internal.

**Kalau gagal:** kalau ada file yang kebawa padahal harusnya nggak, `git reset` dulu, tambahin ke `.gitignore`, baru commit ulang.

---

## Daftar periksa "minggu ini selesai"

- [ ] Bisa nemu thread penyebab CPU tinggi dalam < 5 menit (bukti: `evidence/`)
- [ ] Pernah lihat sendiri OOM cgroup: `memory.events` `oom_kill > 0` **atau** baris journal `systemd-oomd`
- [ ] Punya tabel tiga kondisi (mati / nggak listen / nggak responsif) lengkap dengan exit code
- [ ] Bisa jelasin beda `kill -TERM` dan `kill -KILL`, dengan bukti log shutdown
- [ ] Bisa bedain load average karena CPU vs karena IO (`vmstat` + PSI)
- [ ] `notes.md` terisi (bukan cuma template)
- [ ] `commands.md` minimal 15 perintah + output
- [ ] 3 runbook di `runbooks/`
- [ ] `README.md` status `DONE`, bagian `## Referensi` terisi link nyata
- [ ] Semua eksperimen Hari 3 sudah dibersihin (`systemctl --failed` kosong, nggak ada `yes`/`python3` nyisa)

---

## Cara membuktikan ke reviewer

Jalanin blok ini dari root repo. Outputnya harus konsisten dan bisa lo ulang kapan aja.

```bash
# 1. Semua artefak ada
ls -1 labs/week-01-linux/{bahan-bacaan.md,ringkasan.md,lab.md,commands.md,notes.md,README.md}
ls -1 labs/week-01-linux/runbooks/
ls -1 labs/week-01-linux/evidence/

# 2. Bukti OOM (salah satu harus nonzero)
cat /sys/fs/cgroup/memory.events
sudo journalctl -k --no-pager | grep -ci 'out of memory'

# 3. Bukti bisa bedain tiga kondisi
grep -E '^\| (A|B|C):' labs/week-01-linux/notes.md

# 4. Kondisi host sudah bersih setelah eksperimen
systemctl --failed --no-pager
pgrep -af 'yes|sre-oom-lab' || echo "tidak ada sisa eksperimen"

# 5. Commit terakhir minggu ini
git log --oneline -3 -- labs/week-01-linux/
```

> Kalau reviewer nanya "kenapa lo yakin ini OOM, bukan crash biasa?", jawabannya harus: **"karena `oom_kill` di `memory.events` naik dari 0 jadi 1"** — bukan "kayaknya".
