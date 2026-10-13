# Lab — Minggu 05 · CI/CD dan release management

> Ikuti BERURUTAN, jangan lompat. Setiap langkah formatnya: perintah → Yang harus terjadi → Kalau gagal.

## Persiapan (10–15 menit)

Ceklist environment:

- [ ] `git --version` → v2.30+
- [ ] Docker dengan Compose v2: `docker compose version` → `Docker Compose version v2.x` (kalau muncul perintah `docker-compose` dengan tanda hubung, berarti lo masih pakai v1 yang sudah dihapus — install plugin Compose v2).
- [ ] `go version` → 1.22+ (repo ini build dengan Go 1.25 di `docker/Dockerfile`).
- [ ] GitHub CLI: `gh --version` → 2.x, dan sudah login: `gh auth status`.
- [ ] Repo ini sudah ada remote GitHub: `git remote -v`.
- [ ] Punya akses bikin package: `gh api /user` → keluar username lo.

Verifikasi sebelum mulai:

```bash
gh auth status
docker compose -f docker/compose.yaml config >/dev/null && echo "compose valid"
cd app && go vet ./... && echo "vet ok"
```

**Yang harus terjadi:** tiga perintah di atas keluar tanpa error.
**Kalau gagal:** `gh auth login` dulu; kalau `go vet` error, berarti kode di repo udah rusak — catat di `notes.md`, jangan lanjut.

> Catatan: sepanjang lab ini, ganti `<owner>/<repo>` dengan repo lo sendiri, mis. `budi/production-reliability-lab`.

## Hari 1 · Teori (2 jam)

### Langkah 1.1 — Baca bahan bacaan utuh

Buka `bahan-bacaan.md` di folder ini dan baca dari atas ke bawah tanpa buka tab lain.

**Yang harus terjadi:** lo bisa jawab 7 pertanyaan di bagian *Self-check* tanpa Google.
**Kalau gagal:** ulangi bagian 2 (konsep inti) dan bagian 3 (mekanisme), jangan lanjut ke praktik.
**Catat di notes.md:** 3 konsep yang paling nggak lo duga sebelumnya.

### Langkah 1.2 — Bedah `ci.yml` yang sudah ada

```bash
cat .github/workflows/ci.yml
```

Jawab tertulis di `notes.md`:

1. Job apa aja yang ada, dan mana yang jalan paralel?
2. Kenapa `build-push` punya `needs:` dan `if:`?
3. Permission apa yang diminta tiap job, dan apa alasannya?
4. Apa yang terjadi kalau `go test` exit code-nya 1?

**Yang harus terjadi:** jawaban nomor 4 = "job `build-push` tidak pernah jalan".
**Kalau gagal:** baca lagi bagian 3 di `bahan-bacaan.md`.
**Catat di notes.md:** keempat jawaban di atas.

### Langkah 1.3 — Baca sumber primer (2 sumber wajib)

Buka [Workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax) bagian `permissions` dan `concurrency`, lalu [Artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations).

**Yang harus terjadi:** lo tahu beda `contents: read` vs `packages: write`, dan tahu attestation mengklaim apa.
**Kalau gagal:** cari lagi di halaman yang sama, jangan pindah ke blog acak.
**Catat di notes.md:** satu kalimat definisi attestation dengan bahasa lo sendiri.

### Langkah 1.4 — Cek postur keamanan repo lo

```bash
gh api repos/<owner>/<repo>/actions/permissions | grep -E "default_workflow_permissions|can_approve"
```

**Yang harus terjadi:** keluar `default_workflow_permissions` (nilainya `read` atau `write`).
**Kalau gagal:** `gh: Not Found` berarti repo belum punya Actions aktif — aktifkan di *Settings → Actions → General*.
**Catat di notes.md:** nilainya. Kalau `write`, ini temuan yang wajib lo tulis di laporan Hari 4.

### Langkah 1.5 — Tulis target angka pipeline

Isi di `notes.md`: berapa target durasi pipeline lo (maksimal 10 menit), dan langkah mana yang paling lama sekarang.

**Yang harus terjadi:** ada angka tertulis, bukan "yang penting jalan".
**Kalau gagal:** ukur dulu durasi run terakhir pakai `gh run list`.
**Catat di notes.md:** tabel kecil: job | durasi sekarang | target.

## Hari 2 · Praktik (2 jam)

### Langkah 2.1 — Jalankan pipeline dan lihat hasilnya

```bash
git switch -c lab/week-05-pipeline
git commit --allow-empty -m "week 05: uji pipeline" && git push -u origin lab/week-05-pipeline
gh pr create --fill
gh run watch
```

**Yang harus terjadi:** workflow `ci` jalan; job `lint-test` dan `integration` hijau; job `build-push` **skip** (karena event-nya PR, bukan push ke main).
**Kalau gagal:** `gh run view --log-failed`; kalau error-nya soal hak akses, cek `permissions:` di job terkait.
**Catat di notes.md:** durasi tiap job.

### Langkah 2.2 — Build image lokal dengan tag SHA

```bash
SHORT=$(git rev-parse --short=12 HEAD)
docker buildx build -t sre-api:sha-$SHORT -f docker/Dockerfile --load .
docker images --digests sre-api
```

**Yang harus terjadi:** image muncul dengan digest `sha256:...`.
**Kalau gagal:** `failed to solve` → cek konteks build (`.` = root repo, bukan folder `docker/`) dan path Dockerfile.
**Catat di notes.md:** SHA + digest.

### Langkah 2.3 — Merge ke main supaya job build jalan

```bash
gh pr merge --squash --delete-branch
gh run watch --workflow ci.yml
```

**Yang harus terjadi:** run baru jalan; `build-push` **aktif** dan selesai; image ter-push ke GHCR.
**Kalau gagal:** lihat log step `build`; error paling umum = `packages: write` belum diset atau login GHCR gagal.
**Catat di notes.md:** URL run + nama tag image.

### Langkah 2.4 — Verifikasi attestation

```bash
gh attestation verify oci://ghcr.io/<owner>/<repo>/sre-api:sha-$SHORT \
  -R <owner>/<repo> \
  --signer-workflow <owner>/<repo>/.github/workflows/ci.yml
```

**Yang harus terjadi:** ada baris `Loaded digest sha256:...` dan `Verification succeeded`.
**Kalau gagal:** `no attestations found` → step `actions/attest@v4` nggak jalan (cek apakah `push-to-registry: true` dan `id-token: write` ada).
**Catat di notes.md:** output lengkap yang sudah disanitasi.

### Langkah 2.5 — Tarik image by digest (jalan rollback)

```bash
DIGEST=$(docker inspect --format='{{index .RepoDigests 0}}' sre-api:sha-$SHORT | cut -d@ -f2)
docker pull ghcr.io/<owner>/<repo>/sre-api@$DIGEST
docker run --rm ghcr.io/<owner>/<repo>/sre-api@$DIGEST -healthcheck; echo "exit=$?"
```

**Yang harus terjadi:** pull sukses; `-healthcheck` **exit 1** (karena nggak ada server yang nyala di container itu) — ini contoh penting: healthcheck mode butuh server-nya hidup.
**Kalau gagal:** `manifest unknown` → digest salah, ulangi dari `docker inspect`.
**Catat di notes.md:** beda output `docker inspect` (digest lokal) dengan `docker pull` by digest (manifest dari registry).

### Langkah 2.6 — Simpan evidence

```bash
gh run list --workflow ci.yml --limit 5 > evidence/05-run-list.txt
gh api /repos/<owner>/<repo>/packages/container/sre-api/versions \
  | grep -E '"name"|sha256' | head -20 > evidence/05-registry-versions.txt
```

**Yang harus terjadi:** dua file di `evidence/` berisi output yang bisa dibaca ulang.
**Kalau gagal:** folder `evidence/` belum ada → `mkdir -p evidence`.
**Catat di notes.md:** daftar file evidence yang lo kumpulkan.

## Hari 3 · Eksperimen kegagalan (2 jam)

> Aturan: **setiap eksperimen punya langkah pemulihan.** Jangan tinggalin repo/branch dalam kondisi rusak.

### Langkah 3.1 — Gagalkan test, buktikan pipeline menolak

```bash
git switch -c lab/week-05-test-gagal
printf 'func TestSengajaGagal(t *testing.T){ t.Fatal("sengaja") }\n' > app/fail_test.go
git add app/fail_test.go && git commit -m "week 05: test sengaja gagal" && git push -u origin lab/week-05-test-gagal
gh pr create --fill && gh run watch
```

**Yang harus terjadi:** job `lint-test` **merah**, dan `build-push` nggak jalan (skip/needs). Ini bukti acceptance criteria #3.
**Kalau gagal:** kalau build tetap jalan, berarti `needs:` nggak kepasang — perbaiki `ci.yml` dan ulangi.
**Catat di notes.md:** potongan log merah + URL run.

**Pemulihan:**

```bash
rm app/fail_test.go && git add -A && git commit -m "week 05: hapus test gagal" && git push
gh pr close lab/week-05-test-gagal --delete-branch
```

### Langkah 3.2 — Uji concurrency: dua push beruntun

```bash
git commit --allow-empty -m "week 05: push A" && git push origin main
git commit --allow-empty -m "week 05: push B" && git push origin main
gh run list --workflow ci.yml --limit 3
```

**Yang harus terjadi:** run yang lebih lama statusnya `cancelled` (kena `cancel-in-progress: true`).
**Kalau gagal:** dua-duanya tetap jalan → `concurrency.group` nggak kebaca (salah indent, atau `github.ref` beda).
**Catat di notes.md:** mana yang dibatalkan, dan berapa detik jaraknya.

**Pemulihan:** nggak ada yang rusak; dua commit kosong ini boleh di-revert: `git revert` dua kali, atau biarkan (kosong, aman).

### Langkah 3.3 — Uji kebocoran secret di history

```bash
# JANGAN bikin secret beneran. Ini cuma file palsu buat nguji detektor.
printf 'AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\n' > /tmp/fake.env
gitleaks detect --no-git -s /tmp/fake.env -v
gitleaks detect --log-opts=--all -v
```

**Yang harus terjadi:** deteksi pertama ketemu "finding" di `/tmp/fake.env` (bukti detektornya jalan); deteksi kedua **bersih** buat repo ini.
**Kalau gagal:** kalau deteksi repo ketemu sesuatu, berarti ada kredensial nyata — **rotasi dulu** baru lanjut, dan jangan commit temuannya mentah-mentah.
**Catat di notes.md:** output yang sudah disanitasi (potong nilainya).

**Pemulihan/cleanup:** `rm /tmp/fake.env`.

### Langkah 3.4 — Uji `permissions:` terlalu ketat (harusnya gagal, dan itu bagus)

```bash
git switch -c lab/week-05-permissions
# sementara: hapus baris `packages: write` dari job build-push, commit, push ke branch ini
gh run watch
```

**Yang harus terjadi:** step push GHCR gagal dengan pesan mirip `resource not accessible by integration` / `denied`. Ini ngebuktiin bahwa `permissions:` beneran berlaku.
**Kalau gagal (justru sukses):** berarti lo lupa nyimpan perubahan, atau repo masih punya `default_workflow_permissions: write`.
**Catat di notes.md:** pesan error persisnya.

**Pemulihan:**

```bash
git checkout .github/workflows/ci.yml && git commit -am "week 05: kembalikan permissions" && git push
git switch main && git branch -D lab/week-05-permissions
```

### Langkah 3.5 — Rollback sungguhan pakai digest

```bash
# Anggap versi sekarang bermasalah. Ambil digest image dari 2 commit lalu.
git log --oneline -3
OLD=$(git rev-parse --short=12 HEAD~2)
docker pull ghcr.io/<owner>/<repo>/sre-api:sha-$OLD 2>/dev/null || echo "tag lama nggak ada (belum pernah di-push)"
docker inspect --format='{{index .RepoDigests 0}}' ghcr.io/<owner>/<repo>/sre-api:sha-$OLD 2>/dev/null
```

**Yang harus terjadi:** lo dapet digest image lama yang bisa dipakai buat deploy ulang. Kalau tag-nya nggak ada, itu sendiri adalah temuan: **retensi image belum diatur**.
**Kalau gagal:** ulangi dengan SHA commit yang pernah ke-push (lihat `gh run list`).
**Catat di notes.md:** perintah rollback final yang akan lo tulis di `docs/release-process.md`.

## Hari 4 · Review & dokumentasi (2 jam)

### Langkah 4.1 — Isi `docs/release-process.md`

Bikin file baru `docs/release-process.md` yang menjawab 4 pertanyaan ini:

| Pertanyaan | Jawaban yang harus ada |
| --- | --- |
| Gimana cara build? | Perintah + di mana jalannya (lokal vs CI) |
| Gimana cara publish image? | Registry + siapa yang punya akses + permission yang dipakai |
| Gimana deploy versi tertentu? | Perintah yang nunjuk **digest** |
| Gimana balikin versi sebelumnya? | Perintah rollback + estimasi waktu + apa yang diverifikasi |

**Yang harus terjadi:** 4 baris itu kejawab dengan perintah nyata dari lab ini, bukan kalimat umum.
**Kalau gagal:** ulangi langkah 2.4 dan 3.5.
**Catat di notes.md:** link ke file-nya.

### Langkah 4.2 — Lengkapi `commands.md`

Salin semua perintah yang lo jalanin Hari 2 dan Hari 3 ke `commands.md`, lengkap dengan output yang disanitasi dan interpretasinya.

**Yang harus terjadi:** minimal 15 perintah, terkelompok per hari.
**Kalau gagal:** lo kurang nyatet — ini alasan kenapa tiap langkah minta lo nulis di `notes.md`.

### Langkah 4.3 — Isi runbook kegagalan

Tulis minimal 3 skenario di `runbooks/README.md` (atau file baru di folder `runbooks/`):

1. Pipeline merah karena test gagal → apa yang dicek pertama.
2. Push ke GHCR ditolak → permission mana yang kurang.
3. Attestation nggak ketemu → langkah verifikasi dan kemungkinan sebab.

**Yang harus terjadi:** tiap skenario punya: gejala → cara buktikan → perintah pemulihan.
**Kalau gagal:** contohin formatnya dari `docs/deployment.md`.

### Langkah 4.4 — Acceptance test ulang dari nol

```bash
gh run list --workflow ci.yml --limit 3
gh attestation verify oci://ghcr.io/<owner>/<repo>/sre-api:sha-$(git rev-parse --short=12 HEAD) -R <owner>/<repo>
gitleaks detect --log-opts=--all -v
```

**Yang harus terjadi:** run hijau, verifikasi lulus, gitleaks bersih.
**Kalau gagal:** balik ke Hari 3, perbaiki, ulangi. Jangan centang minggu ini cuma karena tutorial udah ditonton.

### Langkah 4.5 — Commit

```bash
git add labs/week-05-cicd docs/release-process.md .github/workflows/ci.yml
git commit -m "week 05: CI/CD pipeline + attestation + prosedur rollback"
git push origin main
gh run watch --workflow ci.yml
```

**Yang harus terjadi:** commit masuk, pipeline hijau, dan SHA commit terakhir tercatat di README (`Commit terakhir lab ini:`).
**Kalau gagal:** cek `git status`; jangan `--force` push.

## Daftar periksa "minggu ini selesai"

- [ ] `bahan-bacaan.md` dibaca; 7 self-check bisa dijawab.
- [ ] Pipeline jalan otomatis tiap PR, durasi < 10 menit (ada angkanya).
- [ ] Image punya tag `sha-<12>` dan digest tercatat.
- [ ] Ada bukti run **merah** karena test sengaja digagalkan + job build nggak jalan.
- [ ] `gh attestation verify` lulus (output ada di `evidence/`).
- [ ] `gitleaks detect` bersih di history repo.
- [ ] `permissions:` eksplisit di semua job (top-level + job-level).
- [ ] `docs/release-process.md` menjawab 4 pertanyaan wajib.
- [ ] Semua branch eksperimen (`lab/week-05-*`) sudah dibersihkan.
- [ ] `README.md` bagian Referensi terisi (bukan lagi komentar kosong).

## Cara membuktikan ke reviewer

Tunjukkin 4 hal ini, berurutan:

1. `gh run list --workflow ci.yml --limit 5` — tunjukkin mana yang hijau dan mana yang sengaja merah.
2. `gh attestation verify ...` — tunjukkin output `Verification succeeded` + digest yang cocok dengan yang di registry.
3. `docs/release-process.md` — tunjukkin perintah rollback yang nunjuk digest.
4. `evidence/` — tunjukkin file log/versi registry yang lo simpan (sudah disanitasi).

Kalau reviewer nanya "kenapa nggak pakai `:latest`?", jawabannya ada di `ringkasan.md` → *Jebakan umum*.
