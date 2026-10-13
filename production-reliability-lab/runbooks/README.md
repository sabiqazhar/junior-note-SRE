# Runbook

Runbook menjawab: **kalau ini terjadi lagi, apa yang harus dilakukan?**

Bukan dokumentasi konsep. Bukan penjelasan arsitektur. Runbook dibaca saat panik, jadi formatnya harus konsisten dan bisa dipindai dalam 10 detik.

## Format wajib

```markdown
# <Nama skenario>

## Gejala
Apa yang terlihat: alert apa, apa yang dikeluhkan pengguna, apa yang tampil di dashboard.

## Dampak
Siapa yang terdampak dan seberapa parah. Ini yang menentukan severity.

## Deteksi
Perintah atau query pertama yang dijalankan.
- `...`
- `...`

## Diagnosis (decision tree)
Kalau A → kemungkinan X → cek dengan ...
Kalau B → kemungkinan Y → cek dengan ...

## Mitigasi
Langkah cepat buat mengurangi dampak. Urut, dan sebutkan risikonya.
1. ...
2. ...

## Verifikasi pemulihan
Angka yang harus kembali normal sebelum insiden ditutup.

## Pencegahan
Perbaikan permanen. Link ke action item postmortem kalau ada.

## Eskalasi
Kapan dan ke siapa.
```

## Daftar runbook

| File | Skenario | Dibuat minggu |
| --- | --- | --- |
| `high-cpu.md` | CPU tinggi, request melambat | 1 |
| `oom-symptoms.md` | Container restart mendadak, indikasi OOM | 1 |
| `service-down.md` | Endpoint tidak bisa diakses | 1 |
| `network-failure.md` | DNS failure / connection refused / timeout | 2 |
| `container-restart-loop.md` | Container restart berulang | 3 |
| `proxy-5xx.md` | 502 / 504 / 499 di reverse proxy | 4 |
| `api-down.md` | API tidak bisa di-scrape / tidak merespons | 10 |
| `high-error-rate.md` | Error budget terbakar | 10 & 12 |
| `high-latency.md` | p95 latency naik | 10 |
| `db-saturation.md` | Koneksi database mendekati batas | 10 |
| `backup-failed.md` | Job backup gagal | 7 |
| `restore.md` | Prosedur restore database | 7 |

## Aturan

1. Runbook yang belum pernah dijalankan **belum teruji**. Uji saat game day.
2. Setiap alert wajib nunjuk ke satu runbook lewat `runbook_url`.
3. Kalau runbook ternyata salah saat dipakai, **perbaiki hari itu juga** — besok lo lupa detailnya.
