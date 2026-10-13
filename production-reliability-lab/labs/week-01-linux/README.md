# Week 01 · Linux observability & troubleshooting

> Status: `TODO` → `IN_PROGRESS` → `DONE`
> Commit terakhir lab ini: `-`

## Tujuan minggu ini

Pahami dan kuasai linux observability & troubleshooting sampai bisa menjelaskannya dengan bukti, bukan tebakan.

## Konsep kunci yang harus lo kuasai

- uname/uptime/free/df
- ps/pgrep/lsof/ss
- journalctl/systemctl
- PSI & cgroup v2
- bpftrace/bcc (opsional)

## Yang dikerjakan (Hari 1–4)

- [ ] **Hari 1 · Teori (2 jam)** — baca referensi, tulis catatan konsep di `notes.md` dengan bahasa lo sendiri.
- [ ] **Hari 2 · Praktik (2 jam)** — implementasi lab, uji kondisi normal, simpan output di `evidence/`.
- [ ] **Hari 3 · Eksperimen (2 jam)** — sengaja rusak, ukur, troubleshoot, catat apa yang lo pelajari.
- [ ] **Hari 4 · Review (2 jam)** — acceptance test, README, runbook, commit.

## Acceptance criteria

1. Bisa menunjuk proses/thread penyebab CPU tinggi dalam <5 menit lengkap dengan bukti (`pidstat`, `top -H`, `ps -eLo`).
2. Bisa membuktikan OOM: entri kernel `oom-kill` ATAU `systemd-oomd`, plus cgroup yang kena (`memory.events`, `memory.pressure`).
3. Bisa membedakan service mati vs hidup tapi tidak menerima koneksi (`ss -lntp`, `curl`, `journalctl`).
4. Bisa menjelaskan beda load average tinggi karena CPU vs karena I/O (PSI, `vmstat`).

## Struktur folder

```
labs/week-01-linux/
├── README.md        <- file ini
├── notes.md         <- catatan konsep Hari 1
├── commands.md      <- semua perintah yang dijalankan + output yang disanitasi
├── runbooks/        <- cara memperbaiki kegagalan (Hari 4)
└── evidence/        <- bukti: screenshot, log, hasil pengukuran
```

## Aturan bukti

1. Setiap acceptance criteria di atas harus punya **bukti yang bisa direproduksi** (perintah + output).
2. Output yang mengandung IP internal, token, password, atau data pribadi **wajib disanitasi** sebelum di-commit.
3. Jangan tandai minggu ini selesai hanya karena tutorialnya sudah ditonton.

## Referensi

- [Control Group v2 — The Linux Kernel documentation](https://docs.kernel.org/admin-guide/cgroup-v2.html) — otoritas soal `memory.max`, `memory.current`, `memory.events` (counter `oom_kill`). Wajib buat paham kenapa container OOM padahal host lega.
- [PSI — Pressure Stall Information](https://docs.kernel.org/accounting/psi.html) — beda `some` vs `full`, dan kenapa stall > utilisasi buat ngecek "CPU nggak 100% tapi lemot".
- [signal(7) — Linux manual page](https://man7.org/linux/man-pages/man7/signal.7.html) — tabel sinyal: SIGTERM (15) vs SIGKILL (9), dan sinyal mana yang bisa di-*catch*.
- [journalctl(1) — Linux manual page](https://man7.org/linux/man-pages/man1/journalctl.1.html) — flag `-u`, `-b`, `-k`, `-p`, `--since` yang lo pakai tiap minggu sampai minggu 16.
- [Documentation for /proc/sys/vm/](https://docs.kernel.org/admin-guide/sysctl/vm.html) — `overcommit_memory`, `overcommit_ratio`, `swappiness` yang nentuin kapan OOM muncul.
- [systemd-oomd.service(8)](https://www.freedesktop.org/software/systemd/man/systemd-oomd.service.html) — daemon OOM userspace yang mantau PSI; jelasin kenapa container mati tanpa jejak di `dmesg`.

