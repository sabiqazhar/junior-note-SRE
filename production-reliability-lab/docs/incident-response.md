# Incident Response

## Severity

| Level | Kriteria | Respons | Siapa yang diberi tahu |
| --- | --- | --- | --- |
| **SEV1** | Layanan tidak tersedia untuk semua pengguna, atau data hilang/rusak | Langsung, 24/7 | Semua on-call + comms lead |
| **SEV2** | Fungsi utama terdampak signifikan; error budget terkuras cepat | < 15 menit | On-call utama |
| **SEV3** | Terdampak sebagian, ada workaround | Jam kerja | Ticket |

**Aturan penting:** tetapkan severity di **5 menit pertama**. Severity yang telat ditentukan adalah penyebab respons kacau paling umum. Lebih baik naik severity lalu turunkan, daripada sebaliknya.

## Peran

| Peran | Tanggung jawab |
| --- | --- |
| **Incident Commander (IC)** | Mengarahkan, memutuskan, mendelegasikan. **Tidak sambil debugging.** |
| **Ops lead** | Menjalankan diagnosis dan mitigasi |
| **Comms lead** | Kabar ke pengguna/internal. Cuma dia yang ngomong keluar |
| **Scribe** | Mencatat timeline |

Untuk lab, semua peran ini lo sendiri — tapi **tetap tuliskan timeline-nya**. Kebiasaan ini yang kepakai saat insiden beneran.

## Alur respons

```
1. Deteksi        <- alert nyala
2. Triage         <- tentukan severity dalam 5 menit
3. Mitigasi       <- PULIHKAN DULU. Jangan kejar root cause saat user sedang rugi
4. Diagnosis      <- cari akar masalah
5. Pemulihan      <- verifikasi dengan angka, bukan perasaan
6. Postmortem     <- blameless, dalam 48 jam
```

**Urutan 3 sebelum 4 itu sengaja.** Mitigasi dulu (rollback, failover, scale), analisis belakangan. Insiden yang "menang" itu yang cepat pulih, bukan yang cepat tahu penyebabnya.

## Pertanyaan pertama saat insiden

1. Apa yang berubah? (deploy, konfigurasi, traffic, dependency)
2. Sejak kapan?
3. Siapa yang terdampak dan seberapa banyak?
4. Apakah error budget terkuras? Berapa cepat?

Jawab dengan **angka**, bukan kata sifat. "Banyak yang error" tidak bisa ditindaklanjuti. "12% request gagal sejak 10:31" bisa.

## Template komunikasi

```
[SEV2] API error rate naik sejak 10:31 UTC
Dampak   : ~12% request gagal, 4.200 request terdampak
Status   : Investigating
Mitigasi : rollback ke sha-9f8e21 sedang dijalankan
Update   : berikutnya dalam 15 menit
```

## Yang harus selalu dicatat

- Waktu: mulai, deteksi, respons, mitigasi, pulih
- Angka: error rate, request terdampak, error budget yang terpakai
- Keputusan: apa yang dicoba, apa yang berhasil, apa yang tidak
- Yang bikin lambat: apa yang bikin diagnosis melambat (dashboard kurang, log hilang, akses tidak ada)

Bagian terakhir itu yang paling sering menghasilkan perbaikan nyata.
