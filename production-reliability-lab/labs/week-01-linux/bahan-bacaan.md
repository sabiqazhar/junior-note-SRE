# Minggu 01 · Linux observability dan troubleshooting — Bahan Bacaan

> Dokumen ini berdiri sendiri. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham konsepnya. Baru setelah itu kerjakan lab.md.

---

## 1. Kenapa minggu ini penting

Pernah nggak lo dapet pesan begini jam 2 pagi: *"Aplikasinya lemot banget, tolong dicek."* Terus lo buka dashboard, lihat grafik CPU naik, lalu... bingung mau ngapain. Restart? Tunggu? Naikin instance? Semua itu tebakan. Dan tebakan yang salah di jam 2 pagi harganya mahal: lo bangunin orang yang salah, lo nambah kapasitas buat masalah yang sebenernya cuma satu query jelek, atau lo restart dan bukti hilang — besok pagi lo nggak bisa jelasin apa pun.

Minggu ini ngajarin hal yang lebih mendasar dari Prometheus, Grafana, atau tracing: **gimana caranya nanya ke kernel Linux langsung**. Karena pada akhirnya, semua lapisan di atasnya (container, orchestrator, bahkan bahasa pemrograman lo) cuma ngasih *tampilan* dari apa yang sebenarnya terjadi di kernel. Kalau lo bisa baca kondisi kernel, lo nggak perlu nebak.

Analoginya gini. Bayangin lo montir. Ada yang bilang "mobil gue bunyi". Lo bisa: (a) ganti oli dan berharap bunyinya hilang, atau (b) buka kap, dengerin bunyinya dari mana, pegang bagian yang getar, baru mutusin. SRE itu montir tipe (b). Minggu ini lo belajar "buka kap"-nya Linux: `ps`, `ss`, `lsof`, `/proc`, `journalctl`, dan `cgroup`.

Alasan lain kenapa ini minggu nomor satu: hampir semua insiden produksi ujungnya cuma tiga pertanyaan. **Siapa yang makan CPU? Siapa yang makan memori atau kena limit? Prosesnya masih hidup nggak, dan kalau hidup, kenapa nggak bisa diajak ngomong?** Kalau tiga ini nggak bisa lo jawab dalam lima menit, observability stack secanggih apa pun cuma jadi dekorasi mahal.

---

## 2. Konsep inti

### 2.1 Proses vs thread

Satu proses itu wadah: punya memori sendiri, file descriptor sendiri, dan **satu atau lebih thread**. Thread adalah yang sebenarnya dieksekusi CPU. Ini penting karena `top` biasa nunjukin angka per proses — jadi satu thread nakal di dalam proses yang kelihatannya "normal" bisa nyamar.

Contoh nyata: aplikasi Go lo punya ratusan goroutine yang dipetakan ke belasan thread OS. Kalau satu goroutine looping tanpa henti, CPU satu core bisa 100% sementara prosesnya kelihatan cuma "lumayan sibuk".

```bash
# Lihat per THREAD, bukan cuma per proses
top -H -p <PID>

# Atau versi satu kali jalan, gampang di-copy ke notes
ps -eLo pid,tid,pcpu,comm --sort=-pcpu | head -15
```

### 2.2 Signal: SIGTERM vs SIGKILL

Signal itu cara kernel dan lo "nyolek" proses. Dua yang paling penting:

- **SIGTERM (15)** — "tolong berhenti". Proses **bisa** nangkep signal ini, beberes (nutup koneksi DB, nyelesaiin request yang lagi jalan), lalu keluar dengan rapi. Ini yang dipakai `systemctl stop`, `docker stop`, dan `kubectl delete pod`.
- **SIGKILL (9)** — "mati sekarang". Proses **tidak bisa** nangkep, tidak bisa abaikan, tidak bisa beberes. Kernel yang cabut langsung.

Kenapa ini penting buat SRE: `kill -9` yang jadi kebiasaan bakal nyiptain bug misterius. Koneksi database nggak ditutup, transaksi menggantung, file setengah jadi. Minggu 6 lo bakal lihat sendiri kenapa graceful shutdown itu nggak bisa dicapai kalau lo nembak prosesnya.

```bash
kill -TERM <PID>     # sopan
kill -KILL <PID>     # paksa — cuma kalau TERM nggak digubris setelah waktu wajar
```

### 2.3 systemd dan journalctl

systemd itu init system (PID 1) sekaligus manajer service di hampir semua distro modern. Dua hal yang lo pedulikan:

1. **Dia yang nyalain ulang service** — lewat `Restart=` di unit file. Jadi kalau service lo "misteriusnya hidup lagi", itu bukan keajaiban, itu systemd.
2. **`journalctl` adalah sumber kebenaran.** Log aplikasi bisa hilang, bisa ke-rotate, bisa salah level. Journal systemd nyimpen stdout/stderr service plus log kernel dalam satu tempat.

```bash
systemctl --failed                          # unit yang gagal
systemctl status sre-lab.service
journalctl -u sre-lab.service -n 100 --no-pager
journalctl -p err -b --no-pager             # error sejak boot ini
journalctl -k --no-pager                    # cuma log kernel (tempat OOM killer lapor)
journalctl --since "10 minutes ago" --no-pager
```

### 2.4 Memori: bedanya `free` dan `available`

Ini sumber salah diagnosis nomor satu. Linux itu pelit: memori yang nganggur dia pakai buat **page cache** (nyimpen isi file biar baca disk nggak perlu ulang). Cache ini bisa dibuang kapan aja kalau ada yang butuh.

Jadi pas lo lihat `free -h` dan kolom `free` tinggal sedikit, itu **bukan** berarti memori habis. Yang harus lo lihat:

| Kolom | Arti |
| --- | --- |
| `total` | total RAM fisik |
| `free` | memori yang benar-benar kosong dan nggak dipakai apa pun (selalu kecil, wajar) |
| `buff/cache` | page cache + buffer. Ini **bukan musuh lo** — ini memori yang bisa direklamasi |
| `available` | estimasi memori yang bisa dipakai aplikasi baru **tanpa swapping**. Ini angka yang lo pakai |

Kalau `available` masih besar tapi aplikasi lambat, masalahnya bukan memori. Kalau `available` kecil **dan** swap dipakai terus, baru itu masalah memori.

### 2.5 File descriptor

File descriptor (fd) itu "nomor tiket" yang dipegang proses buat tiap resource terbuka: file, socket, pipe. Setiap koneksi HTTP masuk, setiap koneksi database, setiap file yang dibaca — semuanya makan satu fd.

"Too many open files" bukan bug eksotis. Hampir selalu artinya: ada yang buka koneksi dan lupa nutup.

```bash
ulimit -n                      # batas per proses (shell lo)
ls /proc/<PID>/fd | wc -l      # berapa fd yang dipakai proses ini
cat /proc/<PID>/limits | grep 'open files'
lsof -p <PID> | head -30       # fd apa aja
```

### 2.6 cgroup v2

cgroup (control group) adalah mekanisme kernel buat **ngasih batas resource ke sekelompok proses**. cgroup v2 adalah versi baru dengan satu hierarki terpadu, dan ini yang dipakai Docker, containerd, systemd, dan Kubernetes hari ini.

Kenapa ini krusial: ini alasan container kena OOM **meski RAM host masih lega**. Container nggak ngeliat RAM host — dia cuma ngeliat batas cgroup-nya.

```bash
cat /sys/fs/cgroup/memory.max          # batas memori cgroup saat ini (max = tanpa batas)
cat /sys/fs/cgroup/memory.current
cat /sys/fs/cgroup/memory.events       # counter: low, high, max, oom, oom_kill
cat /sys/fs/cgroup/memory.stat
cat /sys/fs/cgroup/cpu.max             # format: "quota period", contoh "50000 100000" = 0.5 CPU
cat /sys/fs/cgroup/pids.current
```

Yang paling sering lo intip adalah `memory.events`. Kalau kolom `oom_kill` naik, berarti kernel pernah nembak proses di cgroup ini. Itu bukti, bukan perasaan.

### 2.7 PSI (Pressure Stall Information)

PSI menjawab pertanyaan yang paling bikin frustrasi: **"CPU-nya nggak 100%, tapi kok lambat?"**

PSI ngukur **berapa persen waktu proses menunggu** (stalled) karena CPU/memori/IO, bukan berapa persen resource terpakai. Beda fundamental. Utilisasi 80% bisa berarti sehat; stall 40% berarti separuh waktu proses lo cuma ngantri.

```bash
cat /proc/pressure/cpu
cat /proc/pressure/memory
cat /proc/pressure/io
```

Formatnya:

```
some avg10=0.00 avg60=0.00 avg300=2.31 total=1234567
full avg10=0.00 avg60=0.00 avg300=0.00 total=0
```

- **`some`** = persentase waktu di mana *minimal satu* task ngantri resource itu.
- **`full`** = persentase waktu di mana *semua* task ngantri bersamaan. Ini parah: CPU cycle benar-benar kebuang. Kalau `full` tinggi dan lama, sistem lo sedang **thrashing**.
- `avg10/60/300` = tren 10 detik, 1 menit, 5 menit. `total` = akumulasi mikrodetik, berguna buat deteksi spike pendek yang nggak kelihatan di rata-rata.

Aturan praktis buat lab: `avg10` di atas 10% itu sudah berasa oleh user. `avg300` tinggi berarti masalah struktural, bukan spike sesaat.

### 2.8 Load average

Load average (angka tiga di `uptime`) itu **bukan persentase CPU**. Dia adalah jumlah task yang sedang **runnable** (R) atau **tidak bisa di-interupsi** (D — biasanya nunggu disk/IO), dirata-rata eksponensial selama 1, 5, dan 15 menit.

Makanya load tinggi bisa muncul padahal CPU idle: banyak proses ngantri di state `D`. Cek kolom `wa` (wait) di `vmstat` dan kolom `b` (blocked).

```bash
uptime
vmstat 1 5        # kolom r = runnable, b = blocked, wa = % waktu nunggu IO
```

---

## 3. Mekanisme di balik layar: dari memori habis sampai OOM killer menembak

Ini urutan yang sebenarnya terjadi. Lo bakal butuh ini buat ngejelasin insiden dengan percaya diri.

**Langkah 1 — Ada yang minta memori.**
Aplikasi panggil `malloc()`. Di level kernel, ini biasanya cuma reservasi alamat virtual dulu. Memori fisik baru benar-benar dialokasi saat halaman itu disentuh (page fault).

**Langkah 2 — Kernel ngitung apakah ini boleh.**
Tergantung `vm.overcommit_memory` (lihat file `/proc/sys/vm/overcommit_memory`):
- `0` (default) — heuristik. Kernel nebak-nebak, boleh sedikit overcommit.
- `1` — selalu iya, jangan pernah nolak.
- `2` — strict. Tolak kalau sudah lewat `RAM * overcommit_ratio + swap`.

Jadi di mode default, kernel bisa bilang "oke" dulu walau memori sebenarnya nggak cukup. Ini yang bikin OOM muncul *tiba-tiba* jauh kemudian.

**Langkah 3 — Batas cgroup dicek lebih dulu.**
Kalau prosesnya masuk cgroup dengan `memory.max`, kernel nge-charge pemakaian ke cgroup itu. Kalau lewat batas:
- Kernel coba reklamasi page cache di dalam cgroup itu.
- Kalau masih lewat → **OOM cgroup**. Kernel cetak `Memory cgroup out of memory: Killed process ...`.
- Counter `oom` dan `oom_kill` di `memory.events` naik.

**Langkah 4 — Kalau bukan cgroup yang kena, host yang kena.**
Kernel panggil `out_of_memory()` → jalanin `select_bad_process()` → tiap proses dikasih skor `oom_badness()`. Skor itu kira-kira:

```
skor = (persentase memori yang dipakai proses) * 1000 + oom_score_adj
```

- `/proc/<PID>/oom_score` — skor saat ini (0–1000).
- `/proc/<PID>/oom_score_adj` — **bobot yang bisa lo atur**, rentang `-1000` sampai `+1000`.
  - `-1000` = "jangan pernah bunuh gue" (ini yang dipakai `sshd`, systemd, dan pod Kubernetes QoS `Guaranteed`).
  - `+1000` = "gue korban pertama yang wajar" (pod QoS `BestEffort`).

**Langkah 5 — Tembak.**
Proses dengan skor tertinggi dikirim **SIGKILL**. Ingat: SIGKILL nggak bisa ditangkap. Nggak ada graceful shutdown. Nggak ada log aplikasi. Yang ada cuma log kernel.

**Langkah 6 — Lo nemuin buktinya.**
```bash
journalctl -k --no-pager | grep -i -E 'out of memory|oom-kill|killed process'
# atau
dmesg -T | grep -i -E 'out of memory|killed process'
```

**Catatan penting buat 2025+:** di Ubuntu 22.04/24.04 ada **systemd-oomd**, daemon OOM di userspace yang jalan berdampingan dengan kernel OOM killer. Dia nggak nunggu memori benar-benar habis — dia mantau **PSI** dan nembak cgroup dengan tekanan tertinggi begitu kondisi terpenuhi (mis. swap terpakai dan stall memori tinggi). Bedanya: aksinya kelihatan di journal **systemd-oomd**, bukan di log kernel. Makanya "container restart mendadak tanpa jejak OOM di dmesg" itu hal yang nyata, dan jawabannya ada di dua tempat.

---

## 4. Glosarium mini

| Istilah | Arti dengan bahasa sederhana | Contoh konkret |
| --- | --- | --- |
| **Proses** | Wadah berisi memori + fd + minimal satu thread | Satu instance binary `sre-api` |
| **Thread** | Yang benar-benar dieksekusi CPU; bagian dari proses | `top -H` nunjukin 12 thread di PID 9001 |
| **SIGTERM** | "Tolong berhenti, beberes dulu" — bisa ditangkap aplikasi | `systemctl stop` kirim ini |
| **SIGKILL** | "Mati sekarang" — nggak bisa ditangkap, nggak ada beberes | `kill -9`, dan OOM killer pakai ini |
| **Page cache** | Memori yang dipakai nyimpen isi file biar baca disk cepat | Kolom `buff/cache` di `free -h` |
| **MemAvailable** | Estimasi memori yang bisa dipakai aplikasi baru tanpa swap | Angka yang harus lo lihat di `free -h` |
| **File descriptor** | Nomor tiket buat tiap file/socket yang dibuka proses | `ls /proc/9001/fd \| wc -l` → 43 |
| **cgroup v2** | Mekanisme kernel buat ngasih batas CPU/memori ke sekelompok proses | `memory.max` bikin container OOM walau host lega |
| **PSI** | Berapa persen waktu proses *nunggu* resource, bukan terpakai | `some avg10=2.31` di `/proc/pressure/memory` |
| **Load average** | Jumlah task yang runnable atau nunggu IO, rata-rata 1/5/15 menit | `load average: 8.2` di mesin 2 core = parah |
| **oom_score_adj** | Bobot `-1000..1000` yang nentuin siapa dibunuh duluan saat OOM | `-1000` = kebal, `+1000` = korban pertama |
| **systemd-oomd** | Daemon OOM userspace yang mantau PSI dan nembak cgroup | Container mati tanpa OOM di `dmesg` |

---

## 5. Bahan bacaan wajib

| # | Sumber (link) | Bagian mana yang dibaca & apa yang dicari | Estimasi |
| --- | --- | --- | --- |
| 1 | [Control Group v2 — The Linux Kernel documentation](https://docs.kernel.org/admin-guide/cgroup-v2.html) | Fokus ke bagian **Memory**: `memory.max`, `memory.current`, `memory.events`, `memory.stat`. Cari juga kenapa v2 cuma punya satu hierarki. Jangan baca semua — bagian memory dan cpu aja cukup. | 45 menit |
| 2 | [PSI — Pressure Stall Information](https://docs.kernel.org/accounting/psi.html) | Baca bagian **Pressure interface**. Pastiin lo paham beda `some` vs `full`, dan kenapa `full` itu indikasi thrashing. | 20 menit |
| 3 | [signal(7) — Linux manual page](https://man7.org/linux/man-pages/man7/signal.7.html) | Cari tabel Standard signals. Catat nomor dan arti SIGTERM (15) dan SIGKILL (9), lalu baca paragraf soal signal yang tidak bisa di-*catch*. | 20 menit |
| 4 | [Documentation for /proc/sys/vm/](https://docs.kernel.org/admin-guide/sysctl/vm.html) | Cari entri `overcommit_memory`, `overcommit_ratio`, `swappiness`, dan `oom_kill_allocating_task`. Ini yang nentuin kapan OOM bisa terjadi. | 25 menit |
| 5 | [systemd-oomd.service(8) — Ubuntu manpages](https://manpages.ubuntu.com/manpages/noble/en/man8/systemd-oomd.service.8.html) | Baca deskripsi dan syarat kapan dia bertindak. Ini ngejelasin kenapa container bisa mati tanpa jejak di `dmesg`. | 15 menit |
| 6 | [journalctl(1) — Linux manual page](https://man7.org/linux/man-pages/man1/journalctl.1.html) | Cari flag `-u`, `-b`, `-k`, `-p`, `--since`. Ini perintah yang bakal lo pakai tiap minggu sampai minggu 16. | 15 menit |

---

## 6. Self-check (harus bisa dijawab tanpa buka Google)

1. Apa bedanya `kill -TERM` dan `kill -KILL`? Sebutkan satu situasi di mana `-KILL` jadi keputusan yang buruk.
2. Di output `free -h`, kolom mana yang harus lo lihat buat mutusin "memori kita habis atau nggak"? Kenapa kolom `free` selalu kelihatan kecil?
3. Load average tinggi tapi CPU idle. Apa kemungkinan penyebabnya, dan satu perintah buat ngebuktiin?
4. Container lo mati mendadak. `dmesg` nggak nunjukin apa-apa. Ke mana lagi lo cari, dan apa nama daemon-nya?
5. Apa beda baris `some` dan `full` di `/proc/pressure/memory`? Kalau `full avg300` tinggi, apa yang sedang terjadi?
6. Proses hidup (`pgrep` nemu PID) tapi `curl` ke portnya `Connection refused`. Apa yang lo curigai dan apa perintah pembuktinya?
7. Sebutkan dua file di `/sys/fs/cgroup/` yang nunjukin apakah cgroup ini pernah kena OOM.

---

## 7. Yang sering salah paham

| Mitos | Kenyataannya |
| --- | --- |
| "Load average tinggi berarti CPU tinggi." | Load itu jumlah task runnable **plus** task yang nunggu IO (state D). Cek kolom `wa` di `vmstat` dan `b` sebelum nyimpulin CPU. |
| "`free` tinggal dikit, berarti memori habis." | Linux pakai sisa RAM buat page cache. Lihat `available`. `buff/cache` itu memori yang bisa direklamasi, bukan memori bocor. |
| "OOM di container selalu muncul di `dmesg` host." | Yang kena limit itu cgroup. Bisa jadi cuma kelihatan di `memory.events`, atau di journal `systemd-oomd`. Kadang malah cuma di `docker inspect`. |
| "`pidstat` baris pertama itu kondisi sekarang." | Baris pertama `pidstat` itu **rata-rata sejak boot**. Tunggu baris ke-2 baru itu kondisi nyata. |
| "`kill -9` itu solusi cepat yang aman." | Itu nggak ngasih aplikasi kesempatan nutup koneksi DB dengan rapi. Bakal jadi bug misterius di minggu 6 pas lo bahas graceful shutdown. |
| "Satu proses = satu thread." | Satu proses bisa punya puluhan thread. Makanya `top` biasa bisa nyembunyiin pelaku aslinya; pakai `top -H` atau `ps -eLo`. |
| "Kalau CPU nggak 100%, berarti sistem sehat." | Bisa aja semua proses ngantri memori atau IO. Itu yang diukur PSI, bukan utilisasi. |
