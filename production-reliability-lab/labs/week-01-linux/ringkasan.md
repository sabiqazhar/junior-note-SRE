# Ringkasan — Minggu 01 · Linux observability dan troubleshooting

> Satu halaman untuk dibaca ulang sebelum ujian/wawancara. Bukan pengganti `bahan-bacaan.md`.

---

## Inti minggu ini dalam 5 kalimat

1. Semua insiden produksi ujungnya cuma tiga pertanyaan: siapa makan CPU, siapa makan memori atau kena limit, dan prosesnya masih bisa diajak ngomong atau nggak.
2. `free -h` yang harus lo baca itu kolom `available`, bukan `free` — Linux sengaja pakai sisa RAM buat page cache.
3. Load average itu bukan persentase CPU: dia ngitung task yang runnable **dan** yang nunggu IO, makanya bisa tinggi saat CPU idle.
4. Container kena OOM karena batas **cgroup v2**-nya, bukan karena RAM host — buktinya ada di `memory.events` dan journal `systemd-oomd`, seringkali bukan di `dmesg`.
5. PSI (`/proc/pressure/*`) adalah satu-satunya jawaban buat pertanyaan "CPU nggak penuh, tapi kok lambat?": dia ngukur waktu *nunggu*, bukan waktu *terpakai*.

---

## Peta konsep (daftar bertingkat)

```
Linux observability
├─ Identitas proses
│  ├─ proses vs thread            → top -H, ps -eLo
│  ├─ PID & command line          → pgrep -af
│  └─ status internal             → /proc/<PID>/status, /proc/<PID>/limits
├─ CPU
│  ├─ utilisasi per proses        → pidstat -u, top
│  ├─ utilisasi per thread        → top -H -p <PID>
│  ├─ load average                → uptime, vmstat (kolom r, b, wa)
│  └─ tekanan CPU (stall)         → /proc/pressure/cpu
├─ Memori
│  ├─ free vs available           → free -h
│  ├─ page cache (bukan musuh)    → kolom buff/cache
│  ├─ swap & thrashing            → vmstat si/so, psi full
│  ├─ batas cgroup                → memory.max, memory.current, memory.events
│  └─ OOM
│     ├─ kernel OOM killer        → journalctl -k, oom_score, oom_score_adj
│     └─ systemd-oomd (userspace) → journalctl -u systemd-oomd
├─ IO & disk
│  ├─ disk penuh vs inode habis   → df -h, df -i
│  ├─ tekanan IO                  → /proc/pressure/io
│  └─ per proses                  → pidstat -d, iotop
├─ Jaringan & fd
│  ├─ siapa listen di port X      → ss -lntp
│  ├─ fd yang dibuka proses       → ls /proc/<PID>/fd, lsof -p
│  └─ koneksi database            → lsof -i, ss -tnp
├─ Signal & lifecycle
│  ├─ SIGTERM (15) = minta berhenti, bisa ditangkap
│  ├─ SIGKILL (9)  = paksa, nggak bisa ditangkap
│  └─ SIGINT (2)   = Ctrl-C, diperlakukan sama kayak SIGTERM di app Go kita
└─ Log terpusat
   ├─ log unit tertentu           → journalctl -u
   ├─ log kernel                  → journalctl -k
   └─ filter waktu & level        → --since, -p err, -b
```

---

## Perintah kunci + yang harus lo lihat di output

| Perintah | Yang dicari di output | Artinya kalau beda |
| --- | --- | --- |
| `uname -a` | Kernel versi 6.x, arsitektur `x86_64` atau `aarch64` | Kernel < 5.13: `full` di `/proc/pressure/cpu` nggak terlapor |
| `uptime` | `load average` 1/5/15 menit | 1 > 5 > 15 = beban baru naik. 15 > 1 = beban sudah turun |
| `free -h` | Baris `Mem:` kolom **available** | `available` kecil + swap terpakai = masalah memori nyata |
| `df -h` | Kolom `Use%` tiap mount | 100% = disk penuh → aplikasi bisa gagal nulis log/DB |
| `df -i` | Kolom `IUse%` | Penuh padahal `df -h` lega = kehabisan inode (banyak file kecil) |
| `ps aux --sort=-%cpu \| head -10` | Baris teratas = proses paling rakus CPU | Yang muncul malah `top`/`ps` sendiri = nggak ada yang aneh |
| `ps -eLo pid,tid,pcpu,comm --sort=-pcpu \| head` | TID spesifik dengan `%cpu` tinggi | Proses wajar tapi thread tertentu 100% = cari thread itu, bukan prosesnya |
| `top -H -p <PID>` | Daftar thread dengan `%CPU` | Cuma satu thread tinggi → masalahnya di satu goroutine/worker |
| `pidstat -urd -p <PID> 1 5` | **Baris ke-2 dst**, kolom `%usr`, `%system`, `RSS`, `kB_rd/s` | Baris pertama = rata-rata sejak boot, jangan dipakai |
| `vmstat 1 5` | Kolom `r` (runnable), `b` (blocked), `wa` (wait IO) | `b` dan `wa` tinggi = masalah IO, bukan CPU |
| `cat /proc/pressure/cpu` | `some avg10`, `avg60`, `avg300` | `avg10` > 10% = user sudah ngerasa lambat |
| `cat /proc/pressure/memory` | `some` dan `full` | `full` > 0 terus-menerus = thrashing, kritis |
| `cat /proc/pressure/io` | `some avg10` | Tinggi = disk atau network storage jadi bottleneck |
| `ss -lntp` | Baris `LISTEN` dengan `users:(("sre-api",pid=..))` | Port nggak muncul = proses nggak listen (lihat decision tree) |
| `ss -lntp \| grep 8080` | Satu baris LISTEN di 8080 | Kosong = aplikasi belum bind; dobel = bentrok port |
| `lsof -i :8080` | Proses pemilik port | Muncul `COMMAND` yang bukan aplikasi lo = ada yang rebut port |
| `lsof -p <PID>` | Koneksi ke `postgres`, `redis`, file log | Nggak ada koneksi DB = aplikasi belum konek |
| `ls /proc/<PID>/fd \| wc -l` | Jumlah fd | Mendekati limit di `/proc/<PID>/limits` = risiko EMFILE |
| `cat /proc/<PID>/status \| grep -E 'Threads\|VmRSS'` | `Threads: N`, `VmRSS: N kB` | Threads naik terus = leak goroutine/thread |
| `systemctl --failed` | Daftar unit gagal (bisa kosong) | Ada isi = baca `systemctl status <unit>` |
| `journalctl -p err -b --no-pager` | Error sejak boot | Error `sre-api` = aplikasi; error kernel = infrastruktur |
| `journalctl -k \| grep -i 'out of memory'` | Baris `Killed process <PID> (<nama>)` | Ada = OAM kernel; kosong = cek `systemd-oomd` |
| `cat /sys/fs/cgroup/memory.events` | Kolom `oom` dan `oom_kill` | Angka > 0 = cgroup ini pernah OOM. Ini bukti paling kuat |
| `cat /sys/fs/cgroup/memory.max` | Angka byte atau `max` | Angka = ada batas; `max` = nggak dibatasi (bahaya di shared host) |
| `cat /proc/<PID>/oom_score_adj` | `-1000` s.d. `1000` | `-1000` = kebal OOM; `1000` = korban pertama |
| `curl -sS -o /dev/null -w '%{http_code}\n' --max-time 3 http://127.0.0.1:8080/healthz` | `200` | `000` + exit 7 = refused; `000` + exit 28 = timeout |

---

## Failure matrix / decision tree

| Gejala | Kemungkinan penyebab | Cara buktikan | Mitigasi |
| --- | --- | --- | --- |
| Request lambat, CPU < 50% | Proses ngantri IO atau memori (bukan CPU) | `cat /proc/pressure/io`, `vmstat` kolom `wa`, `pidstat -d` | Kurangi beban IO, tambah cache, atau pindahkan storage |
| Request lambat, CPU 100% | Satu thread looping / query berat | `top -H -p <PID>`, `ps -eLo ... --sort=-pcpu` | Identifikasi thread → perbaiki kode atau scale out |
| Load average tinggi, CPU idle | Banyak task di state D (nunggu disk/NFS) | `vmstat` kolom `b` tinggi, `wa` tinggi, PSI `io` tinggi | Cek storage: `iostat -x 1`, cek NFS/volume Docker |
| Container restart mendadak | OOM cgroup atau `systemd-oomd` | `memory.events` (`oom_kill`), `journalctl -u systemd-oomd` | Naikin `memory.max` **atau** turunkan working set; jangan asal naikin |
| `Too many open files` | Koneksi DB/HTTP nggak ditutup | `ls /proc/<PID>/fd \| wc -l`, bandingkan `/proc/<PID>/limits` | Perbaiki leak di kode; naikin `LimitNOFILE` cuma buat nafas sementara |
| `Connection refused` | Nggak ada yang listen di port itu | `ss -lntp \| grep <port>`, `pgrep -af <nama>` | Nyalain service, atau perbaiki port/konfigurasi |
| Proses hidup tapi port nggak listen | Aplikasi gagal bind (port dipakai, atau crash saat init) | `pgrep` ada tapi `ss` kosong; baca `journalctl -u` | Baca log aplikasi — hampir selalu ada `bind: address already in use` atau error init |
| Port listen, tapi `curl` nggak pernah balas | Thread pool / koneksi DB habis, aplikasi hang | `curl --max-time 3` exit 28; `ss -tnp` nunjukin backlog | Cek pool koneksi DB, jumlah worker, dependency health |
| Disk `100%` | Log nggak di-rotate, atau volume penuh | `df -h`, `du -sh /var/lib/docker/*` | Rotasi log (`max-size`/`max-file` di compose), bersihkan image lama |
| Disk lega tapi "No space left on device" | Inode habis | `df -i` kolom `IUse%` = 100% | Hapus file kecil yang numpuk (biasanya log atau cache) |

**Decision tree utama minggu ini (hapal ini):**

```
"Endpoint nggak bisa diakses"
├─ pgrep -af <nama> kosong?
│   └─ YA  → proses mati. Baca journalctl -u <unit>, cari kenapa exit.
├─ ss -lntp | grep <port> kosong?
│   └─ YA  → proses hidup tapi nggak listen. Baca log aplikasi (gagal bind/init).
├─ curl --max-time 3 exit 28?
│   └─ YA  → listen tapi nggak responsif. Cek pool koneksi DB, worker, dependency.
└─ exit 0 tapi 5xx
    └─ aplikasi jalan dan jawab, tapi jawabannya jelek. Cek log aplikasi + dependency.
```

---

## Jebakan umum

- **Load average tinggi ≠ CPU tinggi.** Bisa jadi proses numpuk nunggu I/O. Cek `wa` di `vmstat` dan `b`.
- **"Memory penuh" palsu.** Linux pakai sisa RAM buat page cache. Lihat `MemAvailable`, dan `buff/cache` bukan musuh lo.
- **OOM di container nggak selalu muncul di `dmesg` host.** Yang kena limit itu cgroup — cek `memory.events` dan `systemd-oomd`.
- **`pidstat` iterasi pertama itu rata-rata sejak boot**, bukan kondisi sekarang. Tunggu baris ke-2.
- **`kill -9` sebagai solusi pertama.** Nggak ngasih aplikasi kesempatan nutup koneksi DB dengan rapi → bakal jadi bug misterius di minggu 6.
- **`top` biasa nyembunyiin thread.** Selalu lanjutkan dengan `top -H` sebelum nyimpulin "nggak ada yang aneh".
- **`df -h` lega bukan jaminan aman.** Bisa aja inode yang habis — selalu jalanin `df -h && df -i` berpasangan.
- **Log aplikasi hilang saat container restart.** Satu-satunya yang bertahan ya `journalctl` (kalau service-nya lewat systemd) atau log driver Docker.

---

## Acceptance criteria

Disalin dari `README.md`, ditambah cara buktinya:

1. Bisa menunjuk proses/thread penyebab CPU tinggi dalam **< 5 menit** lengkap dengan bukti (`pidstat`, `top -H`, `ps -eLo`).
   - Bukti: file `evidence/week-01-high-cpu.txt` berisi output `top -H` + `pidstat` + timestamp.
2. Bisa membuktikan OOM: entri kernel `oom-kill` **ATAU** `systemd-oomd`, plus cgroup yang kena (`memory.events`, `memory.pressure`).
   - Bukti: `evidence/week-01-oom.txt` berisi `memory.events` dengan `oom_kill > 0` dan baris journal yang relevan.
3. Bisa membedakan service mati vs hidup tapi tidak menerima koneksi (`ss -lntp`, `curl`, `journalctl`).
   - Bukti: `evidence/week-01-three-states.md` berisi tabel tiga kondisi + output per kondisi.
4. Bisa menjelaskan beda `kill -TERM` dan `kill -KILL`, serta kapan `-KILL` jadi keputusan buruk.
   - Bukti: catatan di `notes.md` + log aplikasi yang nunjukin `shutdown signal received` saat TERM.
5. *(tambahan)* Bisa menjelaskan beda load average tinggi karena CPU vs karena I/O.
   - Bukti: output `vmstat` dan `/proc/pressure/*` yang diambil saat dua eksperimen berbeda.

---

## Satu kalimat yang harus lo ingat

> **Jangan nebak — tanya kernelnya langsung: `ss` buat port, `/proc` buat proses, `memory.events` buat OOM, dan PSI buat "kok lambat padahal CPU nggak penuh".**
