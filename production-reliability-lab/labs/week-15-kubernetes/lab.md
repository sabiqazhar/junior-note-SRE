# Lab — Minggu 15 · Kubernetes dan Orchestration

> Ikuti BERURUTAN, jangan lompat. Setiap langkah formatnya: perintah → Yang harus terjadi → Kalau gagal → Catat notes.

## Persiapan

- [ ] Docker sudah jalan (kind butuh docker).
- [ ] `kubectl` dan `kind` terinstall.
- [ ] Image API lab sudah build (tag immutable `sha-abc123`, **bukan `:latest`**).
- Lab minggu ini lokal (kind) → **nggak ada biaya cloud**. Tapi cluster tetap perlu dibersihkan biar nggak makan RAM.

---

## Hari 1 · Teori (2 jam)

1. **Baca** `bahan-bacaan.md` section 1–2 → **Yang harus terjadi:** paham Pod/Deployment/ReplicaSet/Service. → **Kalau gagal:** ulang section 2.1. → **Catat notes:** beda ketiganya di `notes.md`.
2. **Beda probe** → **Yang harus terjadi:** jelasin kenapa liveness nggak boleh cek DB. → **Kalau gagal:** ulang 2.2. → **Catat notes:** tabel startup/readiness/liveness.
3. **QoS & kapan overkill** → **Yang harus terjadi:** bisa sebutkan kapan K8s nggak sepadan. → **Kalau gagal:** ulang 2.3 & 2.5. → **Catat notes:** contoh kasus tim 2 orang.
4. **Versi & rekomendasi** → **Yang harus terjadi:** tahu K8s 1.34, Ingress NGINX retire → Gateway API, PSA restricted. → **Kalau gagal:** ulang 2.6. → **Catat notes:** kenapa nggak pakai Ingress NGINX baru.

---

## Hari 2 · Praktik (2 jam)

1. **`kind create cluster --name sre --image kindest/node:v1.34.0`** → **Yang harus terjadi:** "Creating cluster sre ... done." → **Kalau gagal:** cek docker. → **Catat notes:** versi node 1.34.
2. **`kubectl cluster-info && kubectl get nodes`** → **Yang harus terjadi:** node Ready. → **Kalau gagal:** cek kubeconfig. → **Catat notes.**
3. **`kubectl apply -f kubernetes/`** → **Yang harus terjadi:** deployment + service created. → **Kalau gagal:** cek manifest (tag image). → **Catat notes:** perhatikan `replicas: 3`.
4. **`kubectl rollout status deployment/sre-api`** → **Yang harus terjadi:** "successfully rolled out". → **Kalau gagal:** `kubectl describe pod`. → **Catat notes.**
5. **`kubectl get pods -w`** (Ctrl-C) → **Yang harus terjadi:** 3 pod Running 3/3. → **Kalau gagal:** `kubectl logs`. → **Catat notes.**
6. **`kubectl port-forward svc/sre-api 8080:80`** + **`curl -fsS http://localhost:8080/readyz`** → **Yang harus terjadi:** `ok`. → **Kalau gagal:** cek Service selector. → **Catat notes.**
7. **Cek HPA:** **`kubectl get hpa`** → **Yang harus terjadi:** min 3, max 20, target 70%. → **Kalau gagal:** cek `autoscaling/v2`. → **Catat notes:** behaviour scaling.
8. **Cek PDB:** **`kubectl get pdb`** → **Yang harus terjadi:** `minAvailable: 2`. → **Kalau gagal:** cek manifest. → **Catat notes:** jaga availabilitas saat evict.

---

## Hari 3 · Eksperimen (2 jam) — sengaja rusak

1. **Self-healing:** **`kubectl delete pod <pod-name>`** lalu **`kubectl get pods`** → **Yang harus terjadi:** pod baru muncul otomatis, tetap 3. → **Kalau gagal:** cek ReplicaSet. → **Catat notes:** ini acceptance #2.
2. **Scale:** **`kubectl scale deployment sre-api --replicas=5`** → **Yang harus terjadi:** 5 pod. → **Kalau gagal:** cek resource node. → **Catat notes.**
3. **Rollout gagal (probe):** **`kubectl set image deployment/sre-api api=ghcr.io/<user>/sre-api:sha-bad123`** → **`kubectl rollout status deployment/sre-api`** → **Yang harus terjadi:** mandek/gagal karena probe. → **Kalau gagal:** `kubectl describe` cari Events. → **Catat notes:** ini cara rollback kebukti.
4. **Rollback:** **`kubectl rollout undo deployment/sre-api`** → **Yang harus terjadi:** balik ke ReplicaSet sehat. → **Kalau gagal:** `kubectl rollout history`. → **Catat notes.**
5. **Readiness eksperimen (paling berharga):** set `readinessProbe` ke path salah, deploy → **Yang harus terjadi:** rollout mandek TANPA service down (pod lama tetap serve). → **Kalau gagal:** cek Endpoint. → **Catat notes:** bukti probe = pelindung.
6. **Debug:** **`kubectl logs <pod> --previous`**, **`kubectl describe pod <pod>`**, **`kubectl top pod --sort-by=memory`** → **Yang harus terjadi:** lihat penyebab crash. → **Catat notes.**

🧹 **HARI 3 WAJIB — CLEANUP (lab lokal):** cluster kind makan RAM; hapus setelah lab.

```bash
kubectl delete -f kubernetes/        # hapus manifest lab
kind delete cluster --name sre       # hapus cluster lokal
# Yang harus terjadi: "Deleted nodes ... / Deleted cluster sre".
# Pod yang dihapus di atas CUMA eksperimen — jangan dikira produksi.
```

---

## Hari 4 · Review (2 jam)

1. **Bandingkan** Docker Compose vs K8s (tabel di `bahan-bacaan.md` section 2.5) → **Yang harus terjadi:** lo bisa jelasin kapan overkill. → **Kalau gagal:** ulang. → **Catat notes.**
2. **Verifikasi** PSA `restricted`: non-root, `readOnlyRootFilesystem`, requests/limits ada. → **Yang harus terjadi:** `kubectl apply` lolos tanpa warning. → **Kalau gagal:** cek securityContext. → **Catat notes.**
3. **Simpan** hasil eksperimen ke `evidence/`. → **Yang harus terjadi:** ada output `kubectl get pods` sebelum/sesudah delete. → **Catat notes.**

---

## Cheat sheet per hari

| Hari | Fokus | Perintah inti |
| --- | --- | --- |
| 1 | Teori | bedah `deployment.yaml`, probe, QoS |
| 2 | Praktik | `kind create`, `kubectl apply`, `rollout status` |
| 3 | Rusak | `delete pod`, `scale`, `set image`, `rollout undo`, `destroy cluster` |
| 4 | Review | banding Compose vs K8s, cek PSA |

## Rubrik observasi per hari

| Hari | Yang lo amati | Tanda lulus |
| --- | --- | --- |
| 1 | Paham Pod/Deployment/Probe | Bisa jelasin bedanya tanpa catatan |
| 2 | App jalan di cluster | 3 pod Ready, `readyz` ok |
| 3 | Self-healing & rollback | delete pod → ganti; undo jalan |
| 4 | Review & PSA | Compose vs K8s, PSA restricted lolos |

## Troubleshooting cepat

| Error | Penyebab | Solusi |
| --- | --- | --- |
| CrashLoopBackOff | app boot lama | tambah startupProbe |
| pod nggak dapat trafik | readiness path salah | perbaiki `/readyz` |
| OOMKilled | memory > limits | naik limits/requests |
| rollout mandek | probe image gagal | `rollout undo` |
| restart massal | liveness cek DB | pisah liveness/readiness |
| PSA reject | root / `latest` | non-root + tag immutable |

## Alat bantu wajib

| Alat | Kegunaan | Cara cek |
| --- | --- | --- |
| `kind` | cluster lokal | `kind version` |
| `kubectl` | kontrol cluster | `kubectl version` |
| `kubectl` debug | ephemeral container | `kubectl debug --help` |
| `jq` | parse output JSON | `jq --version` |

## Langkah pemulihan darurat

- Pod stuck: `kubectl describe pod` → lihat Events; `kubectl logs --previous` untuk crash terakhir.
- Rollout gagal: `kubectl rollout undo deployment/sre-api` untuk balik ke ReplicaSet sehat.
- Node panic: `kind delete cluster --name sre` lalu buat ulang (lab lokal, aman).

## Verifikasi akhir

- [ ] `kubectl get pods` nunjukin 3 pod dari image lab.
- [ ] `kubectl delete pod` → jumlah balik 3 otomatis.
- [ ] `kind delete cluster` selesai, nggak ada cluster nyangkut.

## Daftar periksa

- [ ] Deployment + Service jalan dengan image milik lab (tag immutable).
- [ ] `kubectl delete pod` → pod diganti otomatis (3 tetap).
- [ ] Rollout, rollback, probe failure bisa diamati.
- [ ] Non-root + readOnlyRootFilesystem + requests/limits; PSA restricted aktif.
- [ ] `kind delete cluster` selesai (nggak ada cluster nyangkut).

## Cara buktikan

Lo lulus kalau: (1) `kubectl get pods` nunjukin 3 pod dari image lab, (2) setelah `kubectl delete pod` jumlah balik 3 dalam hitungan detik, (3) `rollout undo` mengembalikan versi sehat, dan (4) `kubectl describe pod` nggak ada warning PSA. Bukti di `evidence/` sudah disanitasi.
