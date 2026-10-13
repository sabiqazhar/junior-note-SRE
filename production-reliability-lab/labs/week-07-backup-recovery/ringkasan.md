# Ringkasan — Minggu 7 · Backup dan Disaster Recovery

> Satu halaman buat dibaca ulang sebelum wawancara/ujian. Bukan pengganti `bahan-bacaan.md`.

## Inti minggu ini dalam 5 kalimat
1. **Backup yang belum pernah di-restore = harapan, bukan backup.**
2. `pg_dump` = logical backup (bisa restore ke titik backup, **nggak bisa PITR**).
3. PITR butuh **base backup + WAL archiving** (pgBackRest / Barman / WAL-G).
4. RPO = data berapa yang boleh hilang; RTO = berapa lama boleh mati. Dua-duanya harus diukur.
5. Verifikasi = cek jumlah row + constraint + checksum, bukan cuma "exit code 0".

## Peta konsep
- Backup
  - Logical (`pg_dump`) vs Physical (base backup)
  - PITR = base backup + WAL archiving
  - Tool 2025: pgBackRest (v2.57), Barman (3.14), WAL-G (3.0)
- RPO / RTO / Retention / Off-site
- Restore drill (di database DISPOSABLE, bukan produksi)

## Perintah kunci + yang harus lo lihat
| Perintah | Yang dicari | Artinya kalau beda |
| --- | --- | --- |
| `pg_dump -U app -d sre_lab -Fc > b.dump` | file `.dump` kecil & padat | DB kosong / kredensial salah |
| `pgbackrest --stanza=app info` | daftar backup + ukuran + waktu | repo belum diinit |
| `pgbackrest --stanza=app verify` | `status: ok` | integritas rusak |
| `SELECT count(*) FROM orders;` | jumlah row == sumber | restore nggak lengkap |
| `pgbackrest ... --type=time --target=... restore` | database balik ke titik waktu | WAL nggak lengkap → gagal |

## Failure matrix / decision tree
| Gejala | Kemungkinan penyebab | Cara buktikan | Mitigasi |
| --- | --- | --- | --- |
| PITR gagal | `archive_mode=off` | `show archive_mode` | nyalakan + restart |
| WAL hilang | retensi WAL < retensi full | `pgbackrest info` | panjain retensi WAL |
| Restore "sukses" tapi row kurang | nggak verifikasi | `count(*)` vs sumber | checksum + count |
| Backup + WAL kena ransomware | satu bucket tanpa lock | audit bucket | pisah + object lock |

## Jebakan umum
- `pg_dump` nggak bisa PITR.
- Lupa `archive_mode = on` → WAL nggak kearsip → PITR gagal total.
- Verifikasi cuma "exit 0" → bukan bukti data benar.
- Backup & WAL di bucket sama tanpa object lock → kena ransomware, hilang dua-duanya.

## Acceptance criteria
1. Restore benar-benar berhasil ke instance/database terpisah.
2. Data hasil restore lolos verifikasi (row + constraint + checksum).
3. Waktu recovery diukur & dibandingkan RTO.
4. Backup terenkripsi & lokasinya terpisah dari volume DB.

## Satu kalimat yang harus lo ingat
> Backup yang belum pernah di-restore itu cuma harapan — restore drill di database disposable bukan produksi.
