# Perintah — Week 05

Format: perintah → output → interpretasi. Ganti `<owner>/<repo>` dengan repo lo sendiri.

## Hari 2 (kondisi normal)

### 1. Siapkan branch lab & push

```bash
git switch -c lab/week-05-pipeline
git commit --allow-empty -m "week 05: uji pipeline" && git push -u origin lab/week-05-pipeline
```

- **Output:** `branch 'lab/week-05-pipeline' set up to track 'origin/lab/week-05-pipeline'` lalu hash commit.
- **Interpretasi:** Branch baru terdaftar di remote, jadi PR bisa dibuat. Commit kosong cukup buat memicu workflow tanpa ubah kode.

### 2. Buka PR dan pantau run

```bash
gh pr create --fill
gh run watch
```

- **Output:** tabel run dengan kolom STATUS/CONCLUSION; job `lint-test` & `integration` = success, `build-push` = skipped.
- **Interpretasi:** Karena event-nya `pull_request`, job `build-push` (yang pakai `if: push && ref==main`) dilewati. Ini perilaku yang benar.

### 3. Build image lokal pakai tag SHA

```bash
SHORT=$(git rev-parse --short=12 HEAD)
docker buildx build -t sre-api:sha-$SHORT -f docker/Dockerfile --load .
```

- **Output:** `=> exporting to image` ... `naming to docker.io/library/sre-api:sha-abc123def456`.
- **Interpretasi:** `--load` wajib kalau mau pakai image lokal setelah buildx; tanpa itu image cuma ada di cache buildx.

### 4. Lihat digest image

```bash
docker images --digests sre-api
```

- **Output:** kolom DIGEST berisi `sha256:9f2c1a...`.
- **Interpretasi:** Digest ini sidik jari immutable image. Tag `sha-...` cuma label manusia; digest yang jadi identitas sejati buat rollback.

### 5. Merge ke main supaya job build jalan

```bash
gh pr merge --squash --delete-branch
gh run watch --workflow ci.yml
```

- **Output:** run baru; `build-push` = success; `Successfully pushed ghcr.io/<owner>/<repo>/sre-api:sha-abc123def456`.
- **Interpretasi:** Event `push` ke `main` mengaktifkan job build + push ke GHCR. Inilah satu-satunya jalur yang menghasilkan image di registry.

### 6. Verifikasi attestation

```bash
gh attestation verify oci://ghcr.io/<owner>/<repo>/sre-api:sha-$SHORT \
  -R <owner>/<repo> \
  --signer-workflow <owner>/<repo>/.github/workflows/ci.yml
```

- **Output:** `Loaded digest sha256:...` + `Verification succeeded`.
- **Interpretasi:** Image ini benar-benar dibangun oleh workflow yang lo sebut. Tanpa baris ini, attestation cuma klaim yang belum dibuktikan.

### 7. Tarik image by digest (simulasi rollback)

```bash
DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' sre-api:sha-$SHORT | cut -d@ -f2)
docker pull ghcr.io/<owner>/<repo>/sre-api@$DIGEST
```

- **Output:** `Status: Downloaded newer image for ...@sha256:...`.
- **Interpretasi:** Menarik by digest = menarik artefak persis, bukan "versi terbaru yang kebetulan ada". Ini inti rollback yang bisa dipertanggungjawabkan.

### 8. Simpan evidence run

```bash
gh run list --workflow ci.yml --limit 5 > evidence/05-run-list.txt
```

- **Output:** file berisi daftar run (status, conclusion, durasi).
- **Interpretasi:** Bukti repo bisa direproduksi oleh reviewer tanpa akses ke akun lo.

## Hari 3 (eksperimen kegagalan)

### 9. Sengaja gagalkan test

```bash
git switch -c lab/week-05-test-gagal
printf 'func TestSengajaGagal(t *testing.T){ t.Fatal("sengaja") }\n' > app/fail_test.go
git add app/fail_test.go && git commit -m "week 05: test sengaja gagal" && git push -u origin lab/week-05-test-gagal
gh pr create --fill && gh run watch
```

- **Output:** job `lint-test` = failure (merah); `build-push` = skipped.
- **Interpretasi:** Bukti acceptance #3: test merah → release ditolak, job build nggak pernah jalan karena `needs: [lint-test, integration]`.

### 10. Lihat log kegagalan

```bash
gh run view --log-failed
```

- **Output:** baris `--- FAIL: TestSengajaGagal` dan `t.Fatal("sengaja")`.
- **Interpretasi:** Log merah ini bukti lo. Simpan ke `evidence/` kalau perlu.

### 11. Uji concurrency: dua push beruntun

```bash
git commit --allow-empty -m "week 05: push A" && git push origin main
git commit --allow-empty -m "week 05: push B" && git push origin main
gh run list --workflow ci.yml --limit 3
```

- **Output:** salah satu run berstatus `cancelled`.
- **Interpretasi:** `concurrency.group: ci-${{ github.ref }}` + `cancel-in-progress: true` membatalkan run lama supaya nggak nabrak.

### 12. Deteksi secret di file palsu

```bash
printf 'AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\n' > /tmp/fake.env
gitleaks detect --no-git -s /tmp/fake.env -v
```

- **Output:** `Finding: AWS Access Key ID` dengan lokasi file.
- **Interpretasi:** Detektor jalan. Di lab ini kita pakai file palsu supaya nggak ada kredensial nyata yang ke-commit.

### 13. Deteksi secret di history repo

```bash
gitleaks detect --log-opts=--all -v
```

- **Output:** `no leaks found` (untuk repo lab ini).
- **Interpretasi:** History bersih. Kalau ketemu, itu temuan serius: rotasi dulu, jangan commit temuannya mentah.

### 14. Permissions terlalu ketat (harus gagal)

```bash
git switch -c lab/week-05-permissions
# hapus baris `packages: write` dari job build-push, commit, push
gh run watch
```

- **Output:** step push GHCR gagal: `resource not accessible by integration` / `denied: permission`.
- **Interpretasi:** Bukti `permissions:` beneran berlaku. Ini yang bikin least-privilege bukan cuma tulisan.

### 15. Rollback sungguhan pakai digest lama

```bash
OLD=$(git rev-parse --short=12 HEAD~2)
docker pull ghcr.io/<owner>/<repo>/sre-api:sha-$OLD 2>/dev/null || echo "tag lama nggak ada"
docker inspect --format='{{index .RepoDigests 0}}' ghcr.io/<owner>/<repo>/sre-api:sha-$OLD 2>/dev/null
```

- **Output:** digest image 2 commit lalu, atau pesan "tag lama nggak ada".
- **Interpretasi:** Kalau tag-nya nggak ada, itu sendiri temuan: retensi image belum diatur. Rollback cuma mungkin kalau artefak lama masih ada di registry.

### 16. Pemulihan eksperimen (bersihkan branch)

```bash
rm app/fail_test.go && git add -A && git commit -m "week 05: hapus test gagal" && git push
git checkout .github/workflows/ci.yml && git commit -am "week 05: kembalikan permissions" && git push
git switch main && git branch -D lab/week-05-permissions lab/week-05-test-gagal
```

- **Output:** `Deleted branch lab/week-05-permissions`.
- **Interpretasi:** Repo kembali ke kondisi sehat. Nggak ada branch eksperimen yang dibiarin tergeletak.
