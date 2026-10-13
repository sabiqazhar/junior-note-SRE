# Perintah — Minggu 4 · DNS, TLS, dan Reverse Proxy

Format: perintah → output yang diharapkan (contoh, disanitasi) → interpretasi.

## Hari 2 (kondisi normal)

```bash
docker compose up -d
```
→ `Network ... Created`, `Container ... Started`.
→ Stack nyala. Cek `docker compose ps`.

```bash
docker compose exec nginx nginx -t
```
→ `nginx: the configuration file /etc/nginx/nginx.conf syntax is ok`
→ `nginx: configuration test is successful`
→ Wajib jalan SEBELUM reload. Kalau gagal, proxy nggak mau restart.

```bash
curl -v http://localhost/health
```
→ `* Connected to localhost (127.0.0.1) port 80`, `< HTTP/1.1 200 OK`
→ Request lewat Nginx ke `api:8080` berhasil.

```bash
ss -lntp | grep -E '5432|6379'
```
→ (kosong)
→ Port Postgres/Redis nggak terekspos ke host. Ini acceptance criteria #2.

```bash
nmap -p 1-1024 localhost
```
→ `80/tcp open http`, `5432/6379 closed`
→ Hanya proxy yang kebuka.

## Hari 3 (eksperimen kegagalan)

```bash
docker compose stop api
curl -sS -o /dev/null -w '%{http_code}\n' http://localhost/health
```
→ `502`
→ Upstream nolak koneksi (backend mati).

```bash
docker compose exec nginx tail -3 /var/log/nginx/error.log
```
→ `connect() failed (111: Connection refused) while connecting to upstream`
→ Bukti 502 ada di error.log.

```bash
docker compose start api
curl -sS -o /dev/null -w '%{http_code}\n' http://localhost/health
```
→ `200`
→ Balik normal.

```bash
# endpoint lambat (tambahin di app dulu): GET /slow?ms=60000
curl -sS -o /dev/null -w '%{http_code}\n' 'http://localhost/slow?ms=60000'
```
→ (setelah 30 detik) `504`
→ `proxy_read_timeout 30s` habis.

```bash
curl --max-time 2 -sS -o /dev/null -w '%{http_code}\n' 'http://localhost/slow?ms=60000'
```
→ curl berhenti di 2 detik, status di access.log Nginx = `499`
→ Client nutup duluan karena timeout client < proxy_read_timeout.

## Hari 4 (TLS lokal)

```bash
mkcert -install
```
→ `Created a new local CA` + `The local CA is now installed in the system trust store`
→ CA lokal masuk trust store mesin lo.

```bash
mkcert app.local 127.0.0.1 ::1
```
→ `Created app.local+1.pem` + `Created app.local+1-key.pem`
→ Sertifikat lokal tepercaya di mesin lo. JANGAN dipakai domain publik.

```bash
openssl s_client -connect localhost:443 </dev/null 2>/dev/null | grep 'Verify return code'
```
→ `Verify return code: 0 (ok)`
→ Sertifikat valid & tepercaya.
