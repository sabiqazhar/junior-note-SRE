# Minggu 05 · CI/CD dan release management — Bahan Bacaan

> Dokumen ini berdiri sendiri. Baca dari atas ke bawah tanpa buka apa pun lain. Setelah selesai, lo harusnya paham konsepnya. Baru setelah itu kerjakan lab.md.

## 1. Kenapa minggu ini penting

Bayangin lo mau rilis. Langkahnya ada 14: pull branch terbaru, jalanin test, build image, kasih tag, push ke registry, update compose, restart, cek health, cek log, hapus image lama, kasih tahu tim, update changelog... Sekarang bayangin lo ngerjain itu jam 17.50 hari Jumat, sambil mikirin weekend. Yang kelewat biasanya bukan langkah yang sulit — justru langkah yang kelihatannya sepele: lupa `go vet`, lupa tag, lupa cek health sebelum bilang "udah deploy". Otak manusia memang nggak didesain buat ngikutin 14 langkah berurutan dengan konsisten, apalagi kalau lagi buru-buru. Ini bukan soal ceroboh; ini soal desain sistem yang ngandelin ingatan manusia buat hal yang seharusnya jadi kebiasaan mesin.

Analogi yang paling pas: **checklist pilot**. Pilot senior pun nggak terbang cuma pakai ingatan. Kenapa? Karena checklist mindahin beban "inget" dari otak ke kertas. CI/CD itu checklist yang **jalan sendiri** — bukan cuma ngingetin, tapi beneran ngejalanin langkahnya dan **berhenti** kalau ada yang salah. Bedanya dengan checklist kertas: pipeline nggak bisa dibujuk, nggak capek, dan nggak pernah bilang "ah ini mah udah pasti aman".

Masalah kedua yang diselesaikan minggu ini: **lo harus bisa jawab "persisnya apa yang jalan di produksi sekarang?"** Kalau jawabannya `latest`, berarti lo nggak tahu. `latest` itu mutable — isinya bisa berubah kapan aja tanpa lo sadar. Kalau suatu saat produksi rusak dan lo mau balik ke versi kemarin, lo cuma bisa balik kalau lo tahu *persis* artefak mana yang kemarin jalan, dan artefak itu masih ada. Minggu ini lo bakal bikin supaya pertanyaan itu punya jawaban yang bisa dibuktikan: satu commit → satu image → satu digest → bisa diverifikasi asal-usulnya.

## 2. Konsep inti

### 2.1 CI vs CD

**CI (Continuous Integration)**: tiap perubahan kode digabung ke `main` sesering mungkin, dan tiap kali itu terjadi, mesin ngejalanin verifikasi otomatis (build, test, lint). Tujuannya: konflik dan bug ketemu **dalam hitungan menit**, bukan seminggu kemudian pas mau rilis.

**CD**: dua arti yang sering ketuker —
- *Continuous Delivery*: artefak selalu dalam kondisi siap rilis; rilisnya tetap keputusan manusia (tekan tombol).
- *Continuous Deployment*: tiap commit yang lolos pipeline langsung ke produksi tanpa persetujuan manual.

Buat SRE, CI itu wajib dan murah. CD-nya terserah matangnya organisasi — tapi **dua-duanya butuh pipeline yang bisa dipercaya**. Pipeline yang kadang hijau padahal kodenya rusak ("flaky") lebih berbahaya daripada nggak ada pipeline, karena orang berhenti percaya dan mulai nge-skip.

Contoh nyata: test yang pakai `time.Sleep(2 * time.Second)` buat nungguin Redis. Di laptop lulus, di runner yang lagi penuh gagal. Begitu orang mulai bilang "coba re-run", nilai CI-nya hilang.

### 2.2 Immutable artifact

Satu build menghasilkan satu artefak yang **nggak pernah diubah lagi**. Kalau lo build ulang image di tiap environment (staging build sekali, produksi build sekali), lo sebenarnya nggak pernah nguji barang yang sama dengan yang jalan di produksi. Bisa aja staging dapet dependency versi A, produksi dapet versi B yang baru rilis jam itu juga.

Aturannya: **build sekali, promosikan artefak yang sama** ke staging → produksi. Yang berubah antar environment cuma **konfigurasi**, bukan binernya.

```yaml
# Salah: build ulang per environment
- run: docker build -t api:staging .   # staging
- run: docker build -t api:prod .      # produksi -> artefak beda!

# Benar: build sekali, tag berbasis SHA, dipakai di semua environment
- run: docker build -t ghcr.io/owner/repo/api:sha-${{ steps.sha.outputs.value }} .
```

### 2.3 Versioning & traceability: SHA tag + digest

Dua level identitas image:

| Level | Bentuk | Mutable? | Dipakai buat |
| --- | --- | --- | --- |
| Tag | `api:sha-abc123def456` | secara teknis bisa digeser, tapi kebiasaannya nggak | identifikasi manusia |
| Digest | `api@sha256:9f2c...` | **nggak pernah bisa** | identitas kriptografis, untuk deploy & rollback |

Kenapa harus commit SHA? Karena SHA nyambung ke kode persis. `v1.2.3` juga boleh (dan bagus buat rilis), tapi SHA yang ngejamin "image ini dari commit ini". Di repo ini, `ci.yml` udah pakai `git rev-parse --short=12 HEAD` buat bikin tag `sha-...`.

Rollback yang benar: **nunjuk digest versi sebelumnya**, bukan `latest`, bukan "tag terakhir".

```bash
docker pull ghcr.io/<owner>/<repo>/sre-api@sha256:<digest-versi-sebelumnya>
```

### 2.4 Concurrency group

Dua push beruntun ke branch yang sama bisa bikin dua workflow jalan bersamaan. Akibatnya: dua deploy nabrak, artefak ketimpa, atau rollback yang lo jalanin malah ketimpa deploy baru yang masih panas.

```yaml
concurrency:
  group: ci-${{ github.ref }}     # satu grup per branch/PR
  cancel-in-progress: true        # run lama dibatalkan kalau ada yang baru
```

Hati-hati: buat workflow **deploy ke produksi**, `cancel-in-progress: true` bisa membatalkan deploy di tengah jalan. Kalau deploy lo nggak idempotent, lebih aman pakai `cancel-in-progress: false` supaya deploy antre, bukan dibunuh di tengah.

### 2.5 `permissions:` eksplisit (least privilege)

`GITHUB_TOKEN` itu token otomatis yang dikasih GitHub ke tiap workflow. Di repo lama, defaultnya bisa `write` ke banyak hal — artinya workflow lo secara gratis punya hak yang nggak lo butuhin. Kalau ada satu action di pipeline yang disusupi, hak itulah yang dipakai.

Tulis eksplisit di level workflow (minimal) dan di level job (lebih ketat):

```yaml
permissions:          # level workflow: default paling ketat
  contents: read

jobs:
  build-push:
    permissions:      # job ini butuh lebih, kasih seperlunya
      contents: read
      packages: write   # push ke GHCR
      id-token: write   # minta OIDC token
      attestations: write
```

Aturan praktis: mulai dari `contents: read`, tambahin yang lain satu-satu **dengan alasan**.

### 2.6 Secrets di CI — dan jebakan `pull_request_target`

Secret di GitHub Actions itu **terenkripsi saat disimpan** dan **di-mask di log** (keluar sebagai `***`). Tapi masking itu cuma kosmetik: kalau lo sendiri yang nyetak `echo ${{ secrets.TOKEN }} | base64`, hasilnya nggak ke-mask.

Bahaya terbesar di CI bukan secret bocor satu-satu, tapi **workflow yang jalan dengan akses penuh sambil mengeksekusi kode dari orang asing**. Contoh klasik:

```yaml
# JANGAN: jalanin kode dari fork dengan token yang punya hak write
on: pull_request_target
steps:
  - uses: actions/checkout@v5
    with:
      ref: ${{ github.event.pull_request.head.sha }}   # <- RCE
```

`pull_request_target` jalan dalam konteks repo tujuan (punya akses ke secret dan `GITHUB_TOKEN` berhak write), tapi `ref` di atas ngambil kode dari PR si pengirim. Kombinasi ini = siapa aja bisa buka PR yang "curi" secret repo lo. Pakai `pull_request` biasa (tanpa akses secret) atau pisahkan: workflow yang butuh secret jalan terpisah setelah kode direview.

### 2.7 OIDC, bukan secret statis

Cara lama: bikin AWS access key, taruh di secret, key itu umurnya 5 tahun dan nggak pernah dirotasi. Kalau bocor, lo nggak tahu (dan penyerangnya senang).

Cara sekarang: GitHub ngasih **OIDC token** berumur pendek (JWT) ke workflow. Cloud provider-nya dikonfigurasi buat **percaya** pada token dari repo + branch tertentu. Workflow "login" tanpa menyimpan kredensial permanen.

```yaml
permissions:
  id-token: write      # WAJIB kalau mau minta OIDC token
steps:
  - uses: aws-actions/configure-aws-credentials@v4
    with:
      role-to-assume: arn:aws:iam::123456789012:role/gha-deploy
      aws-region: ap-southeast-1
# Nggak ada AWS_ACCESS_KEY_ID di repo. Token-nya habis dalam hitungan menit.
```

Intinya: **nggak ada lagi kredensial jangka panjang yang bisa dicuri**, dan aksesnya bisa dibatasi sampai level "cuma repo ini, cuma branch main".

### 2.8 Artifact attestation & provenance

Attestation itu **pernyataan yang ditandatangani** tentang sebuah artefak: "image dengan digest ini dibangun dari commit ini, oleh workflow ini, di repo ini, pada jam sekian". Ditandatangani pakai identitas workflow (Sigstore/Fulcio), jadi nggak bisa dipalsukan tanpa akses ke repo.

```yaml
- uses: actions/attest@v4
  with:
    subject-name: ghcr.io/${{ github.repository }}/sre-api
    subject-digest: ${{ steps.build.outputs.digest }}
    push-to-registry: true
```

Verifikasinya (ini yang paling sering kelupaan):

```bash
gh attestation verify oci://ghcr.io/<owner>/<repo>/sre-api:sha-abc123def456 \
  -R <owner>/<repo> --signer-workflow <owner>/<repo>/.github/workflows/ci.yml
```

Attestation **tanpa** verifikasi = teater keamanan. Yang bikin berguna justru langkah verifikasinya — dan verifikasi ini bisa dijalankan **sebelum deploy** produksi, jadi image yang nggak jelas asal-usulnya ditolak.

## 3. Mekanisme di balik layar

Dari `git push` sampai image terverifikasi, ini yang sebenarnya terjadi:

1. **Push ke `main`** → GitHub bikin event `push` dan nyocokin ke `on:` di tiap workflow. Yang cocok: `ci.yml`.
2. **Concurrency check** → GitHub ngitung `group: ci-${{ github.ref }}`. Kalau grup ini lagi jalan dan `cancel-in-progress: true`, run lama dibatalkan.
3. **Job di-schedule** → `lint-test` dan `integration` jalan paralel (nggak ada `needs`). `integration` minta service container Postgres 18 dan Redis; runner nunggu healthcheck-nya (`pg_isready`, `redis-cli ping`) sebelum step pertama jalan.
4. **Token dibuat** → tiap job dapet `GITHUB_TOKEN` dengan scope sesuai blok `permissions`. Job `lint-test` cuma dapet `contents: read`.
5. **Test & lint** → `go vet`, `go test -race`. Kalau exit code ≠ 0, job gagal, dan karena `build-push` punya `needs: [lint-test, integration]`, job build **nggak pernah jalan**. Ini yang bikin "test gagal = rilis ditolak".
6. **Build (hanya kalau push ke main)** → `if: github.event_name == 'push' && github.ref == 'refs/heads/main'`. Buildx nyusun image dari `docker/Dockerfile`, `cache-from: type=gha` narik layer cache.
7. **Push ke GHCR** → `docker/login-action` pakai `GITHUB_TOKEN` (butuh `packages: write`), lalu image di-push dengan tag `sha-<12 char>`. Registry ngasih balik **digest** (`sha256:...`) di output step.
8. **Attestation** → `actions/attest@v4` minta OIDC token (`id-token: write`), minta sertifikat singkat dari Sigstore/Fulcio atas nama workflow, tanda tangani metadata build (repo, SHA, workflow path), lalu unggah ke registry + ke GitHub (butuh `attestations: write`).
9. **Verifikasi (di luar pipeline)** → siapa pun (termasuk pipeline deploy) jalanin `gh attestation verify`, yang narik attestation, ngecek tanda tangan ke transparency log, lalu nyocokin digest image dengan yang diklaim.

Yang penting dicatat: **urutan 5 sebelum 6 adalah inti dari CI**. Kalau lo bisa bikin build jalan meski test gagal, lo nggak punya CI — lo cuma punya cron job yang rajin.

## 4. Glosarium mini

| Istilah | Arti dengan bahasa sederhana | Contoh konkret |
| --- | --- | --- |
| **Pipeline** | Rangkaian langkah otomatis yang jalan tiap ada perubahan kode | `ci.yml`: lint → test → build → push → attest |
| **Runner** | Mesin (VM) yang menjalankan job | `runs-on: ubuntu-latest` |
| **Job** | Sekumpulan step yang jalan di satu runner | `lint-test`, `integration`, `build-push` |
| **Step** | Satu perintah atau satu action di dalam job | `- run: go vet ./...` |
| **Artifact** | Hasil build yang bisa dipakai ulang | image `ghcr.io/.../sre-api:sha-abc123def456` |
| **Immutable** | Nggak bisa diubah setelah dibuat; kalau mau beda, bikin yang baru | digest `sha256:...` |
| **Digest** | Sidik jari kriptografis isi image | `sha256:9f2c1a...` |
| **Mutable tag** | Nama yang isinya bisa diganti kapan aja | `:latest`, `:stable` |
| **Concurrency group** | Penanda supaya dua run nggak nabrak di hal yang sama | `group: ci-${{ github.ref }}` |
| **OIDC** | Cara "login" pakai token identitas berumur pendek, tanpa kredensial permanen | `id-token: write` + IAM role |
| **Attestation** | Surat pernyataan yang ditandatangani: "artefak ini dibuat oleh siapa, dari kode apa" | `actions/attest@v4` + `gh attestation verify` |
| **Provenance** | Riwayat asal-usul artefak (bagian dari attestation) | commit SHA, workflow, waktu build |
| **Secret masking** | Secret ditampilkan sebagai `***` di log | `${{ secrets.X }}` |
| **Flaky test** | Test yang kadang lulus kadang gagal tanpa perubahan kode | test bergantung `sleep` atau waktu nyata |

## 5. Bahan bacaan wajib

| # | Sumber (link) | Bagian mana yang dibaca & apa yang dicari | Estimasi |
| --- | --- | --- | --- |
| 1 | [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax) | Bagian `permissions`, `concurrency`, `jobs.<job_id>.needs`, `jobs.<job_id>.if`. Cari: nilai yang diterima `permissions` dan gimana `concurrency` nge-grup run. | 30 menit |
| 2 | [Use GITHUB_TOKEN for authentication in workflows](https://docs.github.com/en/actions/tutorials/authenticate-with-github_token) | Cara kerja token otomatis + contoh `permissions` paling ketat. Cari: kenapa harus least privilege. | 15 menit |
| 3 | [Artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations) | Apa yang diklaim attestation, dan dua jenisnya (provenance vs SBOM). Cari: bedanya "attest" sama "sign". | 20 menit |
| 4 | [`gh attestation verify` — GitHub CLI manual](https://cli.github.com/manual/gh_attestation_verify) | Semua flag: `--signer-workflow`, `--deny-self-hosted-runners`, `-R`. Cari: gimana verifikasi gagal itu kelihatannya. | 15 menit |
| 5 | [About security hardening with GitHub Actions](https://docs.github.com/en/actions/reference/security/secure-use) | Bahaya script injection via `${{ }}` dan `pull_request_target`. Cari: contoh RCE dari fork PR. | 25 menit |
| 6 | [gitleaks](https://github.com/gitleaks/gitleaks) | Cara jalanin `gitleaks detect` dan `gitleaks protect`. Cari: beda deteksi di history vs di staging area. | 15 menit |

## 6. Self-check (harus bisa dijawab tanpa buka Google)

1. Apa beda Continuous Delivery dan Continuous Deployment? Yang mana yang wajib dulu dikuasai?
2. Kenapa build ulang image di tiap environment itu masalah, meski kodenya sama?
3. Kalau produksi rusak dan lo mau rollback, identitas apa yang harus lo pakai — tag `latest`, tag SHA, atau digest? Kenapa?
4. Apa efek `cancel-in-progress: true` pada workflow deploy produksi yang nggak idempotent?
5. Kenapa `pull_request_target` + checkout `head.sha` itu celah RCE, sementara `pull_request` biasa nggak?
6. Apa yang dibuktikan oleh attestation, dan kenapa attestation tanpa verifikasi dianggap nggak ada gunanya?
7. Sebutin dua cara secret bisa bocor di CI meski GitHub udah nge-mask secret di log.

## 7. Yang sering salah paham

| Mitos | Kenyataannya |
| --- | --- |
| "CI hijau berarti kodenya benar." | CI hijau cuma berarti **pemeriksaan yang lo tulis** lulus. Kalau nggak ada test buat fitur itu, CI-nya tetep hijau. |
| "Tag `:latest` praktis, tinggal pull." | `latest` mutable dan nggak bisa dipakai buat rollback yang bisa dipertanggungjawabkan. Lo nggak tahu persis apa yang jalan. |
| "Secret di GitHub Actions udah aman, kan di-mask di log." | Masking cuma ngeganti tampilan. Kalau lo nyetak hasil `base64`/`env` ke log, isinya keluar. |
| "Attestation = image-nya aman dari vulnerability." | Attestation nyambungin artefak ke **asal-usul** (siapa/build darimana), bukan nyatakan bebas CVE. Tetep butuh scan (Trivy). |
| "Gratis nambahin `permissions: write-all` biar nggak error." | Itu ngasih hak penuh ke tiap action di pipeline, termasuk action pihak ketiga. Mulai dari `read`, tambahin seperlunya. |
| "Pipeline yang lamaan dikit nggak apa-apa, yang penting hijau." | Pipeline > 10–15 menit bikin orang malas nunggu dan mulai merge tanpa lihat hasil. Kecepatan pipeline itu fitur keselamatan. |
