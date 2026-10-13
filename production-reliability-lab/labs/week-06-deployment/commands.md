# Perintah — Week 06

Format: perintah → output → interpretasi.

## Hari 2 (kondisi normal)

### 1. Validasi compose & health

```bash
docker compose -f docker/compose.yaml config >/dev/null && echo "compose valid"
curl -fsS http://localhost/health && echo " health ok"
curl -fsS http://localhost/readyz && echo " readyz ok"
```

- **Output:** `compose valid`, `ok health`, `ok readyz`.
- **Interpretasi:** Sistem sehat sebagai baseline. Ini angka pembanding buat eksperimen Hari 3.

### 2. Generate age key

```bash
age-keygen -o ~/.config/sops/age/keys.txt
grep -o 'age1[^ ]*' ~/.config/sops/age/keys.txt
```

- **Output:** `Public key: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`.
- **Interpretasi:** Public key dipakai buat enkripsi (SOPS); private key di `keys.txt` jangan pernah di-commit.

### 3. Enkripsi .env dengan SOPS + age

```bash
sops --encrypt --age "$AGE_PUBLIC" .env.example > .env.enc
file .env.enc
```

- **Output:** `.env.enc: ASCII text` (isi terenkripsi, bukan plaintext).
- **Interpretasi:** File ini aman di-commit. Tanpa `AGE_SECRET_KEY`, nggak ada yang bisa baca isinya.

### 4. Dekripsi round-trip

```bash
sops --decrypt .env.enc > /tmp/.env.dec
diff <(grep -v '^#' .env.example) <(grep -v '^#' /tmp/.env.dec) && echo "isi sama"
```

- **Output:** `isi sama`.
- **Interpretasi:** Enkripsi bisa dibalik dengan benar — wajib sebelum percaya ke SOPS.

### 5. Cek dua endpoint

```bash
curl -s -o /dev/null -w "healthz=%{http_code}\n" http://localhost/health
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
```

- **Output:** `healthz=200`, `readyz=200`.
- **Interpretasi:** Liveness dan readiness terpisah tapi sama-sama hijau saat normal.

### 6. Cek exec form di Dockerfile

```bash
grep -n 'ENTRYPOINT\|CMD' docker/Dockerfile
```

- **Output:** `ENTRYPOINT ["/app/sre-api"]`.
- **Interpretasi:** Exec form → sinyal `SIGTERM` nyampe ke binary, bukan ke shell.

## Hari 3 (eksperimen kegagalan)

### 7. Bombardir lalu SIGTERM (graceful shutdown)

```bash
hey -z 30s -c 50 http://localhost/orders &
docker compose stop api
docker compose logs api 2>&1 | grep -c 'shutdown selesai'
```

- **Output:** angka > 0 (log `shutdown selesai`), nyaris nol 5xx.
- **Interpretasi:** Proses selesaiin request in-flight lalu exit bersih. Inilah graceful shutdown.

### 8. Redis mati → readyz 503, healthz 200

```bash
docker compose stop redis
sleep 3
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
curl -s -o /dev/null -w "healthz=%{http_code}\n" http://localhost/health
```

- **Output:** `readyz=503`, `healthz=200`.
- **Interpretasi:** Liveness tetap hidup, readiness cabut dari LB. Ini yang bikin DB flaky nggak memicu restart massal.

### 9. Kembalikan Redis (normal)

```bash
docker compose start redis
sleep 3
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
```

- **Output:** `readyz=200`.
- **Interpretasi:** Sistem pulih sendiri tanpa restart aplikasi.

### 10. Password DB salah → startup gagal

```bash
sed -i.bak 's/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=salah123/' .env
docker compose up -d api
sleep 5
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
docker compose ps api
```

- **Output:** `readyz=503`, status `unhealthy`.
- **Interpretasi:** Readiness nggak bohong — container nggak pernah dianggap sehat.

### 11. Kembalikan password (normal)

```bash
mv .env.bak .env
docker compose up -d api
sleep 5
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
```

- **Output:** `readyz=200`.
- **Interpretasi:** Setelah config benar, aplikasi sehat lagi.

### 12. Lihat digest image jalan

```bash
docker inspect --format='{{index .RepoDigests 0}}' production-reliability-lab-api
```

- **Output:** `ghcr.io/.../sre-api@sha256:9f2c...`.
- **Interpretasi:** Inilah digest yang dipakai buat rollback — bukan tag `latest`.

### 13. Simulasi rollback by digest

```bash
docker pull ghcr.io/<owner>/<repo>/sre-api@sha256:<digest-lama>
docker compose up -d --wait
curl -s -o /dev/null -w "readyz=%{http_code}\n" http://localhost/readyz
```

- **Output:** `readyz=200` (kalau digest lama ada).
- **Interpretasi:** Rollback berhasil ditandai dengan digest, bukan tebakan tag.

### 14. Kalau digest lama hilang (temuan)

```bash
docker pull ghcr.io/<owner>/<repo>/sre-api@sha256:<digest-lama> 2>&1 | head -1
```

- **Output:** `manifest unknown`.
- **Interpretasi:** Retensi image belum diatur → rollback mustahil. Catat sebagai temuan di Hari 4.

### 15. Verifikasi batas Compose

```bash
docker compose up -d --wait
docker compose ps
```

- **Output:** container diganti; ada jeda singkat tanpa response.
- **Interpretasi:** Compose v2 nggak punya `rollout` seperti k8s — zero-downtime di sini terbatas, catat jujur.

### 16. Cleanup akhir

```bash
make down && make up
curl -fsS http://localhost/health && curl -fsS http://localhost/readyz
```

- **Output:** health & readyz 200.
- **Interpretasi:** Sistem kembali ke kondisi sehat, siap untuk commit.
