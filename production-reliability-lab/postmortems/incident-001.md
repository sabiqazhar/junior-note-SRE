# Incident 001 — Template

> Salin file ini tiap kali ada insiden. Nama: `incident-002.md`, dst.
> **Blameless:** fokus ke sistem yang memungkinkan kejadian ini, bukan ke orang.
> Kalau di bagian root cause tertulis nama orang, postmortem ini gagal.

```markdown
# Incident 00X — Judul singkat (gejala, bukan penyebab)

## Metadata
| Item | Isi |
| --- | --- |
| Severity | SEV2 |
| Durasi | 34 menit |
| Penulis | <nama> |
| Tanggal | YYYY-MM-DD |
| Status | Teratasi / Monitoring |

## Impact
- 12% request gagal selama 34 menit
- ~4.200 request terdampak (dari total X request)
- Error budget 30 hari terpakai 8%
- Tidak ada kehilangan data / tidak ada dampak finansial langsung

## Timeline (UTC — tulis apa adanya, jangan dibaguskan)
| Waktu | Kejadian |
| --- | --- |
| 10:14 | Deploy versi sha-abc123 ke lab |
| 10:31 | Alert HighErrorBudgetBurn firing |
| 10:33 | On-call mulai investigasi |
| 10:36 | Cek dashboard: error rate naik, latency normal |
| 10:41 | Hipotesis: connection pool habis |
| 10:44 | Verifikasi: koneksi DB mencapai batas maksimum |
| 10:47 | Rollback ke sha-9f8e21 |
| 10:48 | Error rate kembali normal |

## Deteksi
- Alert yang menyala: HighErrorBudgetBurn
- Waktu dari masalah mulai → alert: 17 menit
- Apakah ada sinyal yang lebih cepat seharusnya menyala? Ya: saturasi koneksi DB
  belum punya alert.

## Root cause
Satu kalimat, jujur, menyebut mekanisme bukan orang.
Contoh: "Deploy memperkenalkan query tanpa batas LIMIT sehingga koneksi database
dipegang lebih lama dari kapasitas pool yang tersedia."

## Contributing factors (5-whys)
1. Kenapa request gagal? Koneksi DB habis.
2. Kenapa koneksi habis? Query baru menahan koneksi lebih lama.
3. Kenapa query baru lolos? Tidak ada integration test yang mengukur durasi query.
4. Kenapa tidak ada test itu? Tidak pernah ada insiden serupa sebelumnya.
5. Kenapa tidak pernah terdeteksi? Tidak ada alert untuk saturasi koneksi.

## What went well
- Alert SLO menyala dalam 17 menit
- Rollback berjalan tanpa masalah — prosedur sudah pernah diuji

## What went wrong
- Tidak ada dashboard untuk saturasi koneksi database
- Deteksi bergantung pada error rate, bukan pada sinyal leading (saturasi)
- Butuh 12 menit untuk sampai ke hipotesis yang benar

## Luck
- Kejadian di jam kerja, bukan jam 3 pagi
- Traffic sedang rendah, jadi dampak terbatas

## Action items
| # | Tindakan | Tipe | Owner | Tenggat | Prioritas |
| --- | --- | --- | --- | --- | --- |
| 1 | Tambah panel + alert saturasi koneksi DB | Deteksi | | +1 minggu | P1 |
| 2 | Integration test yang mengukur durasi query | Pencegahan | | +2 minggu | P1 |
| 3 | Batasi ukuran hasil query di layer aplikasi | Mitigasi | | +2 minggu | P2 |

## Pelajaran
Satu atau dua kalimat yang bisa dipakai orang lain.
```

## Aturan penulisan

1. Timeline pakai waktu **sebenarnya**, termasuk yang membuat lo terlihat lambat.
2. Action item wajib punya **owner dan tenggat**. Tanpa itu, ia cuma harapan.
3. Bedakan **mitigasi** (mengurangi dampak sekarang) dan **pencegahan** (mencegah terulang).
4. Jangan tutup postmortem sebelum action item P1 selesai.
