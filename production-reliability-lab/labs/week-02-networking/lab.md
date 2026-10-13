# Lab — Minggu 02 · Networking untuk backend engineer

> Target: lo bisa memecah satu request jadi fase DNS→TCP→TLS→HTTP→body, dan mereproduksi 4 kegagalan klasik dengan bukti.

## Persiapan (10–15 menit)

1. **Pastiin stack lab jalan** → perintah: `docker compose -f docker/compose.yaml up -d --build --wait` → Yang harus terjadi: `docker compose ps` nunjukin `api`, `postgres`, `redis` sehat. → Kalau gagal: baca `docker compose logs api`. → Catat di notes.md.
2. **Cek tool ada** → perintah: `command -v curl dig ss openssl` → Yang harus terjadi: keempatnya ada di PATH. → Kalau gagal: `sudo apt install -y dnsutils iproute2 openssl`. → Catat di notes.md.
3. **Catat IP & port** → perintah: `ip addr | grep 'inet '` dan `ss -lntp | grep 8080` → Yang harus terjadi: lo tau interface dan port 8080 listen. → Kalau gagal: API belum jalan, balik ke langkah 1. → Catat di notes.md.

## Hari 1 · Teori (2 jam)

1. **Baca bahan bacaan** → perintah: `cat labs/week-02-networking/bahan-bacaan.md` → Yang harus terjadi: lo paham 5 fase latency & beda exit code. → Kalau gagal: baca ulang bagian #2. → Catat di notes.md.
2. **Pahami model OSI vs TCP/IP** → perintah: buka [Cloudflare Learning Center](https://www.cloudflare.com/learning/network-layer/) → Yang harus terjadi: lo bisa sebutin tiap fase request ada di layer mana. → Kalau gagal: fokus ke DNS & TCP dulu. → Catat di notes.md.
3. **Kenali tool curl -w** → perintah: `curl --help | grep -A2 write-out` → Yang harus terjadi: lo lihat flag `-w` ada. → Kalau gagal: update curl (`sudo apt install -y curl`). → Catat di notes.md.
4. **Catat versi OpenSSL** → perintah: `openssl version` → Yang harus terjadi: keluar `OpenSSL 3.x`. Lo tahu 3.5 LTS rilis 8 Apr 2025. → Kalau gagal: nggak kritis, cuma catat versinya. → Catat di notes.md.
5. **Sketsa request flow** → perintah: gambar client→DNS→TCP→TLS→nginx→API→PG/Redis di `notes.md` → Yang harus terjadi: lo punya peta mental sebelum praktik. → Kalau gagal: lihat ringkasan.md bagian Peta konsep. → Catat di notes.md.

## Hari 2 · Praktik (2 jam)

1. **Interface & route** → perintah: `ip addr` lalu `ip route` → Yang harus terjadi: lo lihat IP lo & default gateway. → Kalau gagal: cek koneksi VM. → Catat di notes.md.
2. **Port yang listen** → perintah: `ss -lntp` → Yang harus terjadi: `:8080` milik `sre-api`, `:80` milik `nginx` (kalau sudah di-setup minggu 4). → Kalau gagal: API belum jalan. → Catat di notes.md.
3. **Resolusi DNS eksternal** → perintah: `dig +short example.com` dan `dig example.com +trace` → Yang harus terjadi: dapet IP; `+trace` nunjukin root→TLD→authoritative. → Kalau gagal: cek `/etc/resolv.conf`. → Catat di notes.md.
4. **Resolusi ala aplikasi** → perintah: `docker compose exec api getent hosts redis` → Yang harus terjadi: dapet IP internal `172.x.x.x redis`. → Kalau gagal: container belum di jaringan sama. → Catat di notes.md.
5. **Ukur 5 fase latency (publik)** → perintah:
   ```bash
   curl -sS -o /dev/null -w 'dns=%{time_namelookup} tcp=%{time_connect} tls=%{time_appconnect} ttfb=%{time_starttransfer} total=%{time_total} http=%{http_code} ver=%{http_version}\n' https://example.com
   ```
   → Yang harus terjadi: `dns` kecil-kecil, `ttfb` > `appconnect`, `ver` = `2` atau `3`. → Kalau gagal: cek koneksi internet. → Catat di notes.md.
6. **Ukur ke service lokal** → perintah:
   ```bash
   curl -sS -o /dev/null -w 'dns=%{time_namelookup} tcp=%{time_connect} ttfb=%{time_starttransfer} total=%{time_total}\n' http://localhost:8080/health
   ```
   → Yang harus terjadi: `dns` ~0 (localhost di /etc/hosts), `total` <50ms. → Kalau gagal: API mati, balik persiapan. → Catat di notes.md (jalankan 5x, bandingkan).

## Hari 3 · Eksperimen kegagalan (2 jam) — WAJIB ada cara kembali ke normal

1. **Kondisi A: port salah → refused** → perintah: `curl --connect-timeout 3 http://localhost:9999; echo "exit=$?"` → Yang harus terjadi: `Connection refused`, `exit=7`. → Kalau gagal: port 9999 kebetulan ada yang listen; ganti port acak. → Catat di notes.md. **Kembali normal:** nggak perlu (request gagal sendiri), lanjut langkah 2.
2. **Kondisi B: service mati** → perintah: `docker compose stop api && curl --connect-timeout 3 http://localhost:8080/health; echo "exit=$?"` → Yang harus terjadi: `Connection refused`, `exit=7`. → Kalau gagal: nginx masih reply (seharusnya nggak kalau api mati & proxy ke api). → Catat di notes.md. **Kembali normal (WAJIB):** `docker compose start api && docker compose ps` → tunggu `api` sehat sebelum lanjut.
3. **Kondisi C: DNS gagal** → perintah: `docker compose exec api curl -v --connect-timeout 3 http://nope-service:8080; echo "exit=$?"` → Yang harus terjadi: `Could not resolve host`, `exit=6`. → Kalau gagal: kebetulan ada service itu; ganti nama absurd. → Catat di notes.md. **Kembali normal:** nggak ada perubahan state, lanjut.
4. **Kondisi D: timeout (DROP)** → perintah: `iptables -A INPUT -p tcp --dport 8080 -j DROP && curl --connect-timeout 3 http://localhost:8080/health; echo "exit=$?"` → Yang harus terjadi: hang lalu `exit=28` (Timeout). → Kalau gagal: pastiin `iptables` ada & lo punya akses root; kalau nggak, simulasi dengan endpoint `/slow` di app. → Catat di notes.md. **Kembali normal (WAJIB):** `iptables -D INPUT -p tcp --dport 8080 -j DROP` lalu verifikasi `curl` balik normal.
5. **Bukti versi TLS + ALPN** → perintah: `openssl s_client -connect example.com:443 -servername example.com -tls1_3 -alpn h2 -brief` → Yang harus terjadi: `Verify return code: 0` + `TLSv1.3`. → Kalau gagal: host nggak support TLS 1.3; coba host lain. → Catat di notes.md.
6. **Isi tabel failure** → perintah: salin hasil A–D ke `commands.md` (bagian Hari 3) → Yang harus terjadi: tabel exit code lengkap. → Kalau gagal: ulang langkah 1–4. → Catat di notes.md.

## Hari 4 · Review & dokumentasi (2 jam)

1. **Diagram request flow** → perintah: gambar di `notes.md` (ASCII/Mermaid): client→DNS→TCP→TLS→nginx→API→PG/Redis. → Yang harus terjadi: diagram ada & lo bisa jelasin. → Kalau gagal: lihat ringkasan Peta konsep. → Catat di notes.md.
2. **Decision tree** → perintah: tulis decision tree (exit 6/7/28/0+5xx) di `notes.md`. → Yang harus terjadi: tree mencakup 4 cabang. → Kalau gagal: lihat ringkasan Failure matrix. → Catat di notes.md.
3. **Lengkapi commands.md** → perintah: isi bagian Hari 2 & Hari 3 dengan output asli. → Yang harus terjadi: ≥15 perintah terdokumentasi. → Kalau gagal: tambah percobaan. → Catat di notes.md.
4. **Update README** → perintah: isi `## Referensi` di README.md. → Yang harus terjadi: 5–8 link primer nyata. → Kalau gagal: lihat bahan bacaan #5. → Catat di notes.md.
5. **Commit** → perintah: `git add labs/week-02-networking && git commit -m "week 02 networking lab"` → Yang harus terjadi: commit masuk. → Kalau gagal: cek `.gitignore` jangan ke-commit secret. → Catat di notes.md.

## Daftar periksa "minggu ini selesai"

- [ ] Lo bisa jelasin 5 fase latency dan tunjuk angkanya di `curl -w`.
- [ ] Lo bisa reproduksi exit 6 (DNS), 7 (refused), 28 (timeout).
- [ ] Lo paham beda `dig` vs `getent hosts`.
- [ ] Lo bisa buktiin versi TLS & ALPN pakai `openssl s_client`.
- [ ] `commands.md` punya ≥15 perintah + output + `notes.md` punya diagram & decision tree.
- [ ] `iptables DROP` SUDAH DIHAPUS (state bersih).

## Cara membuktikan ke reviewer

```bash
# 1. Bukti fase latency (jalankan, tunjukkan angka)
curl -sS -o /dev/null -w 'dns=%{time_namelookup} tcp=%{time_connect} tls=%{time_appconnect} ttfb=%{time_starttransfer} total=%{time_total}\n' https://example.com

# 2. Bukti exit code (3 kondisi)
curl --connect-timeout 3 http://localhost:9999; echo "refused exit=$?"   # 7
docker compose stop api; curl --connect-timeout 3 http://localhost:8080/health; echo "exit=$?"; docker compose start api  # 7 lalu kembalikan
docker compose exec api curl --connect-timeout 3 http://nope-service:8080; echo "dns-exit=$?"  # 6

# 3. Bukti TLS
openssl s_client -connect example.com:443 -servername example.com -tls1_3 -alpn h2 -brief

# 4. State bersih
sudo iptables -C INPUT -p tcp --dport 8080 -j DROP 2>/dev/null && echo "MASIH DROP!" || echo "bersih"
docker compose ps
```
