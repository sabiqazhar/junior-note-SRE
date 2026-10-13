# SLO Lab

> Angka di dokumen ini adalah **target latihan**, bukan klaim performa aplikasi.
> Tujuannya supaya kebiasaan mendefinisikan SLI/SLO terbentuk, bukan supaya angkanya terlihat bagus.

## SLI 1 — Availability

| Item | Isi |
| --- | --- |
| Definisi | Proporsi request **valid** yang tidak menghasilkan status 5xx |
| Sumber data | `http_server_requests_total`, diukur **di reverse proxy**, bukan dari internal aplikasi |
| Window evaluasi | 30 hari rolling |
| Eksklusi | `/health`, `/healthz`, `/readyz`, `/metrics`; traffic dari load generator internal |
| Alasan eksklusi |健康检查 bukan traffic pengguna; mengikutkannya akan membuat angka terlihat lebih baik dari kenyataan |

**Kenapa diukur di proxy?** Kalau aplikasi crash sebelum sempat mencatat metrik, request itu tidak pernah terhitung sebagai error — padahal itu kegagalan terburuk. Mengukur dari sisi klien menutup celah ini.

## SLO 1 — Availability

| Item | Isi |
| --- | --- |
| Target | **99,5%** selama 30 hari rolling |
| Error budget | 0,5% → sekitar 5.000 request gagal per 1 juta request |
| Alerting | Multiwindow multi-burn-rate: 14.4 (page), 6 (page), 1 (ticket) |

## SLI 2 — Latency

| Item | Isi |
| --- | --- |
| Definisi | Proporsi request valid yang selesai **< 300 ms** |
| Kondisi pengukuran | Beban ≤ 50 VU, diukur dengan k6, skenario `ramp` |
| Window evaluasi | 30 hari rolling |
| Eksklusi | Request ke `/slow` (memang sengaja lambat) |

## SLO 2 — Latency

| Item | Isi |
| --- | --- |
| Target | **95%** request valid selesai < 300 ms pada beban yang didefinisikan |

## Error budget policy

Kebijakan ini yang mengubah perdebatan "boleh rilis atau tidak" dari opini menjadi angka.

| Kondisi budget | Konsekuensi |
| --- | --- |
| > 50% tersisa | Rilis bebas, eksperimen boleh agresif |
| 25–50% tersisa | Rilis normal, tapi perubahan berisiko butuh review ekstra |
| < 25% tersisa | Rilis fitur dihentikan; hanya perbaikan reliabilitas |
| Habis | **Freeze rilis.** Wajib 1 postmortem + minimal 1 item reliabilitas masuk sprint berikutnya |

## Cara menghitung burn rate

```
burn rate = (error ratio terukur) / (1 - SLO)
```

Contoh: SLO 99,5% → budget 0,5%. Kalau error ratio 1 jam = 3,6%, maka burn rate = 3,6 / 0,5 = **7,2×**. Dengan burn rate segini, 5% budget 30 hari habis dalam 6 jam.

## Yang sengaja belum ditangani

Ditulis jujur supaya dokumen ini tidak jadi klaim palsu:

- SLI belum diukur dari sisi pengguna sungguhan (masih dari proxy lab).
- Traffic lab rendah, jadi persentase mudah berubah drastis oleh sedikit request. Perlu aturan khusus low-traffic.
- Belum ada SLO untuk durability data backup.
