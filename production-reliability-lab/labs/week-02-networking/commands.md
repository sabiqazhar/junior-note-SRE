# Perintah — Week 02 · Networking untuk backend engineer

Format: `perintah` → output yang diharap → interpretasi. Output disanitasi (IP/waktu diganti placeholder). Jalankan di Ubuntu 22.04/24.04 + Docker Compose v2 dengan stack lab jalan.

---

## Hari 2 (kondisi normal)

**1. Alamat IP & interface**
```bash
ip addr
```
→ `2: eth0: <BROADCAST,...> inet 192.168.1.20/24 ...`
→ Lo tau IP lo dan interface yang aktif.

**2. Tabel routing**
```bash
ip route
```
→ `default via 192.168.1.1 dev eth0` + subnet lokal.
→ Ke mana traffic dilempar kalau tujuan bukan lokal.

**3. Port yang listen + pemilik proses**
```bash
ss -lntp
```
→ `LISTEN 0 4096 *:8080 users:(("sre-api",pid=9001,fd=3))`
→ `pid=9001` = proses Go lo. Bukti port listen & siapa pemiliknya.

**4. Resolusi DNS eksternal (cuma jawaban)**
```bash
dig +short example.com
```
→ `93.184.216.34`
→ Jawaban DNS mentah dari resolver publik.

**5. Jejak resolusi DNS lengkap**
```bash
dig example.com +trace
```
→ baris `A` dari root (`.`) → TLD (`com.`) → authoritative.
→ Lihat persis di mana resolusi terjadi.

**6. Resolusi ala aplikasi (Docker internal)**
```bash
docker compose exec api getent hosts redis
```
→ `172.20.0.3 redis`
→ Ini yang dipakai aplikasi beneran, bukan `dig` di host.

**7. 5 fase latency ke publik**
```bash
curl -sS -o /dev/null -w 'dns=%{time_namelookup} tcp=%{time_connect} tls=%{time_appconnect} ttfb=%{time_starttransfer} total=%{time_total} ver=%{http_version}\n' https://example.com
```
→ `dns=0.04 tcp=0.06 tls=0.09 ttfb=0.25 total=0.26 ver=2`
→ `tls>tcp` (TLS butuh TCP), `ttfb-tls` = server mikir, `total-ttfb` = transfer body.

**8. Latency ke service lokal**
```bash
curl -sS -o /dev/null -w 'dns=%{time_namelookup} tcp=%{time_connect} ttfb=%{time_starttransfer} total=%{time_total}\n' http://localhost:8080/health
```
→ `dns=0.00 tcp=0.00 ttfb=0.002 total=0.003`
→ `dns~0` karena localhost ada di `/etc/hosts`.

**9. Protokol + kode via curl**
```bash
curl -sS -o /dev/null -w 'http=%{http_code} ver=%{http_version} exit=%{exitcode}\n' https://example.com
```
→ `http=200 ver=2 exit=0`
→ `ver=2` = HTTP/2. `exit=0` = sukses.

---

## Hari 3 (eksperimen kegagalan)

**10. Port salah → connection refused (exit 7)**
```bash
curl --connect-timeout 3 http://localhost:9999; echo "exit=$?"
```
→ `curl: (7) Failed to connect ... Connection refused` `exit=7`
→ Nggak ada yang listen di 9999.

**11. Service mati → refused (exit 7)**
```bash
docker compose stop api
curl --connect-timeout 3 http://localhost:8080/health; echo "exit=$?"
docker compose start api
```
→ `Connection refused` `exit=7`, lalu `docker compose start` kembalikan state.
→ Bukti: matiin container = refused.

**12. DNS gagal → could not resolve (exit 6)**
```bash
docker compose exec api curl -v --connect-timeout 3 http://nope-service:8080; echo "exit=$?"
```
→ `Could not resolve host: nope-service` `exit=6`
→ Host nggak ada di DNS internal.

**13. Timeout (DROP) → exit 28**
```bash
sudo iptables -A INPUT -p tcp --dport 8080 -j DROP
curl --connect-timeout 3 http://localhost:8080/health; echo "exit=$?"
sudo iptables -D INPUT -p tcp --dport 8080 -j DROP
```
→ hang lalu `exit=28` (Timeout), lalu `iptables -D` kembalikan normal.
→ DROP = diam → client nunggu sampai timeout.

**14. Versi TLS + ALPN (h2)**
```bash
openssl s_client -connect example.com:443 -servername example.com -tls1_3 -alpn h2 -brief
```
→ `Verify return code: 0 (ok)` + `TLSv1.3` + `ALPN protocol: h2`
→ Bukti server nerima TLS 1.3 & HTTP/2.

**15. SNI salah → sertifikat berbeda**
```bash
openssl s_client -connect example.com:443 -brief 2>/dev/null | grep -E 'subject|issuer'
```
→ `subject=CN=example.com` (tanpa `-servername` bisa dapet default vhost).
→ Bukti SNI ngaruh ke sertifikat yang dikembalikan.

**16. DNS lokal vs resolver**
```bash
cat /etc/resolv.conf
getent hosts localhost
```
→ `nameserver 127.0.0.53` (systemd-resolved stub) + `127.0.0.1 localhost`.
→ Jelasin kenapa `dig` dan aplikasi bisa beda hasil.

**17. Cek exit code setelah curl**
```bash
curl -sS -o /dev/null -w 'code=%{http_code} exit=%{exitcode}\n' http://localhost:8080/health
```
→ `code=200 exit=0` (sehat) / `code=000 exit=7` (refused).
→ `http_code=000` = curl gagal sebelum response HTTP.
