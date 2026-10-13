# Ringkasan — Minggu 05 · CI/CD dan release management

> Satu halaman untuk dibaca ulang sebelum ujian/wawancara. Bukan pengganti bahan-bacaan.md.

## Inti minggu ini dalam 5 kalimat

1. Deployment manual itu insiden yang nunggu waktu: manusia nggak andal ngikutin 14 langkah berurutan, jadi pindahkan langkah itu ke mesin.
2. Build **sekali**, hasilkan satu artefak immutable, lalu promosikan artefak yang sama ke semua environment — yang beda cuma konfigurasinya.
3. Identitas image yang bisa dipertanggungjawabkan itu digest (`sha256:...`), bukan `:latest`; rollback = nunjuk digest versi sebelumnya.
4. Pipeline harus bisa **menolak**: kalau test gagal, job build nggak boleh jalan (`needs:` + exit code ≠ 0).
5. Keamanan supply chain minggu ini: `permissions:` eksplisit, OIDC bukan kredensial permanen, attestation **plus verifikasinya**.

## Peta konsep (daftar bertingkat)

- **CI/CD**
  - CI: build + test + lint otomatis tiap perubahan
  - CD: *Delivery* (siap rilis, tombol manual) vs *Deployment* (otomatis ke produksi)
  - Pipeline yang bisa dipercaya → nggak flaky, cukup cepat (<10 menit), bisa menolak
- **Artefak**
  - Immutable: dibuat sekali, nggak diubah
  - Identitas: tag SHA (manusia) → digest `sha256:` (mesin)
  - Promotion: build di CI → dipakai di staging → dipakai di produksi
  - Rollback: pull digest versi sebelumnya
- **Kontrol alur workflow**
  - `on:` (push / pull_request)
  - `concurrency.group` + `cancel-in-progress`
  - `needs:` (urutan & penolakan)
  - `if:` (mis. cuma jalan di `refs/heads/main`)
- **Keamanan pipeline**
  - `permissions:` minimal (`contents: read` sebagai dasar)
  - Secret: encrypted at rest, di-mask di log, tapi gampang bocor kalau di-echo
  - `pull_request_target` + checkout PR head = RCE
  - OIDC (`id-token: write`) → token pendek, tanpa kredensial permanen
- **Supply chain**
  - Attestation (provenance): siapa build, dari commit apa, workflow mana
  - Verifikasi: `gh attestation verify --signer-workflow ...`
  - Pelengkap: image scan (Trivy), secret scan (gitleaks)
- **Evidence wajib**
  - Log run hijau (PR) dan run merah (test sengaja digagalkan)
  - Daftar tag + digest di GHCR
  - Output `gitleaks detect` bersih

## Perintah kunci + yang harus lo lihat di output

| Perintah | Yang dicari di output | Artinya kalau beda |
| --- | --- | --- |
| `gh run list --workflow ci.yml --limit 5` | Kolom `status` = `completed`, `conclusion` = `success` | `failure` = ada step gagal; `cancelled` = kena concurrency group |
| `gh run view <run-id> --log-failed` | Baris error + nama step yang gagal | Kalau kosong padahal merah, kemungkinan job di-cancel |
| `gh run watch <run-id>` | Durasi tiap job | Job > 10 menit = calon masalah adopsi |
| `git rev-parse --short=12 HEAD` | 12 karakter hex, mis. `abc123def456` | Beda dengan SHA di UI = lo di branch/detached HEAD lain |
| `docker buildx build -t sre-api:sha-$(git rev-parse --short=12 HEAD) -f docker/Dockerfile .` | `=> exporting to image`, `naming to ...` | `failed to solve` = Dockerfile/konteks salah |
| `docker images --digests sre-api` | Kolom `DIGEST` berisi `sha256:...` | `none` = image dibuat lokal tanpa push ke registry |
| `gh attestation verify oci://ghcr.io/<owner>/<repo>/sre-api:sha-<sha> -R <owner>/<repo>` | `Loaded digest sha256:...` + `Verification succeeded` | `no attestations found` = step attest belum jalan/nggak di-push |
| `gitleaks detect --no-git -v` | `no leaks found` (exit 0) | Ada `Finding:` = secret nyangkut di working tree |
| `gitleaks detect --log-opts=--all -v` | Sama, tapi scan seluruh history | Ketemu di history = secret pernah di-commit, harus dirotasi |
| `gh api repos/<owner>/<repo>/actions/permissions` | `default_workflow_permissions` (biasanya `read` atau `write`) | `write` = repo lo ngasih hak berlebih ke semua workflow |
| `gh api /repos/<owner>/<repo>/packages/container/sre-api/versions` | Daftar versi + `name` (digest) + tag | Nggak ada tag SHA = workflow build nggak jalan di commit itu |
| `docker pull ghcr.io/<owner>/<repo>/sre-api@sha256:<digest>` | `Status: Downloaded newer image` / `Image is up to date` | `manifest unknown` = digest salah atau image udah dihapus |
| `docker run --rm sre-api:sha-<sha> -healthcheck` | Exit 0 | Exit 1 = binary jalan tapi `/healthz` nggak jawab (nggak ada server hidup) |

## Failure matrix / decision tree

| Gejala | Kemungkinan penyebab | Cara buktikan | Mitigasi |
| --- | --- | --- | --- |
| Job `integration` gagal di `go test` | Service container belum sehat / DSN salah | Lihat log step; cek `pg_isready` healthcheck | Tambah `health-cmd` + `health-retries`, pakai env `DATABASE_URL` yang sama dengan compose |
| Build jalan meski test merah | `needs:` nggak diisi, atau jobnya beda workflow | Cek `gh run view --log-failed` dan bagian `needs:` | Tambah `needs: [lint-test, integration]` |
| Dua deploy nabrak | Nggak ada `concurrency` | Dua run aktif di `gh run list` | Tambah `concurrency.group` per ref |
| `resource not accessible by integration` | `permissions:` kurang | Bandingkan hak yang dibutuhkan action (`packages: write` buat GHCR push) | Tambah permission **spesifik** di level job, jangan `write-all` |
| `no attestations found` | Step attest nggak jalan, atau `subject-digest` kosong | Cek output step build (`digest`) dan step attest | Pastikan `id: build` ada dan `attestations: write` + `id-token: write` diset |
| Attestasi ada tapi verifikasi gagal | `--signer-workflow` beda path/branch, atau image dari runner self-hosted | `gh attestation verify ... --signer-workflow <path>` | Cocokin path workflow; tambah `--deny-self-hosted-runners` kalau kebijakan lo gitu |
| Secret muncul di log | Di-echo/`env`/`base64` di step | `gh run view --log` lalu grep nama variabel | Jangan nyetak secret; pakai `::add-mask::` untuk nilai turunan |
| `latest` di produksi nggak sesuai ekspektasi | Tag mutable ketimpa build baru | `docker inspect` bandingkan digest | Pakai digest di compose/manifest |
| Test lulus di lokal, gagal di CI (acak) | Flaky: waktu, urutan, network | Re-run 3× tanpa ubah kode | Hilangkan `sleep`, pakai retry/healthcheck dan `-count=1` |
| Cache bikin hasil beda antar versi Go | Go build cache bentrok sama matrix | Matiin `cache: true`, jalanin ulang | Pisah `cache-dependency-path` + key per versi |

## Jebakan umum

- **`:latest` musuh rollback.** Mutable, dan lo nggak pernah tahu persis apa yang jalan.
- **`pull_request_target` + checkout PR head** = celah RCE. Jangan.
- **Lupa `permissions:`** di repo lama → workflow dapet akses `write` gratis.
- **Attestation tanpa verifikasi** = teater keamanan. Sertakan `gh attestation verify`.
- **Cache Go bentrok sama matrix build** → build sukses di satu versi Go, gagal di versi lain secara acak.
- **Test flaky** bikin orang mulai re-run tanpa baca; itu awal mula pipeline kehilangan fungsi.
- **Naruh secret di argumen `run:`** (`run: ./deploy.sh ${{ secrets.TOKEN }}`) → keliatan di `ps` dan log; pakai `env:`.
- **Nge-pin action ke `@main`** → action bisa berubah tanpa lo tahu; pin ke tag versi atau SHA commit.

## Acceptance criteria (salin dari README, boleh ditambah)

1. Test + lint jalan otomatis tiap PR dan selesai **< 10 menit**.
2. Setiap commit punya image dengan tag berbasis SHA yang bisa dilacak (+ digest tercatat).
3. Test sengaja digagalkan → pipeline menolak release (bukti: log run merah, job build nggak jalan).
4. Tidak ada secret tercetak di log/artefak; `permissions:` ditulis eksplisit.
5. *(tambahan)* `gh attestation verify` lulus untuk minimal satu image, dan perintahnya tercatat di `commands.md`.
6. *(tambahan)* Prosedur rollback tertulis dengan perintah yang nunjuk digest, bukan tag.

## Satu kalimat yang harus lo ingat

> Kalau lo nggak bisa sebutin **digest** image yang lagi jalan di produksi, lo belum punya release management — lo cuma punya kebiasaan deploy.
