# Minggu 15 · Kubernetes dan Orchestration — Bahan Bacaan

> *"Kubernetes itu bukan 'Docker tapi keren', tapi sistem yang menjaga keadaan yang lo mau tetap terjadi, walau node-nya mati."* — prinsip minggu ini.

---

## 1. Kenapa penting

Pas lo cuma punya satu service di satu VPS, `docker run` cukup. Tapi begitu lo punya belasan service, butuh auto-scaling, dan node sesekali mati tengah malam, lo butuh sesuatu yang **menjaga keadaan yang diinginkan (desired state)** secara otomatis. Kalau lo bilang "mau 3 replica", dan satu pod mati, Kubernetes bikin penggantinya tanpa lo bangun. Itulah *self-healing*.

Tapi Kubernetes juga **mahal secara kognitif**. Lo harus paham 15+ objek (Pod, Deployment, ReplicaSet, Service, ConfigMap, Secret, HPA, PDB, dll). Buat tim 2 orang dengan 1 service sepi, Compose + 1 VPS jauh lebih murah dan lebih sedikit insiden. Minggu ini gue ajarin K8s supaya lo paham **kapan pakai, kapan nggak** — bukan sekadar bisa ngetik `kubectl apply`.

Analognya: Deployment itu **restoran**, ReplicaSet itu **manajer shift** yang jaga jumlah pelayan tetap 3, Pod itu **pelayan itu sendiri**. Kalau satu pelayan pulang (crash), manajer langsung panggil pengganti. `kubectl delete pod` itu lo suruh satu pelayan pulang — manajer nggak peduli, dia langsung ganti. Rollout itu **ganti menu secara bertahap** sambil pelayan lama masih jalan, supaya pelanggan nggak lihat restoran tutup.

---

## 2. Konsep inti

### 2.1 Pod, Deployment, ReplicaSet, Service

- **Pod**: satuan terkecil, 1+ container yang share network/storage.
- **ReplicaSet**: jaga N Pod tetap jalan (desired state).
- **Deployment**: pembungkus ReplicaSet yang tahu cara **rolling update & rollback**.
- **Service**: nama DNS stabil (`svc/sre-api`) yang nemuin Pod yang lagi jalan (walau Pod-nya ganti).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata: { name: sre-api }
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate: { maxSurge: 1, maxUnavailable: 0 }   # nol downtime
  selector: { matchLabels: { app: sre-api } }
```

### 2.2 Probes: startup / readiness / liveness

Tiga probe punya tugas beda — jangan campur:

- **startupProbe**: kasih waktu booting lama. Tanpa ini, app lawas sering CrashLoopBackOff palsu.
- **readinessProbe**: boleh cek dependency. Gagal = **dikeluarkan dari Service** (nggak terima trafik) tapi nggak di-restart.
- **livenessProbe**: cuma cek proses hidup. **JANGAN cek dependency** — kalau DB down, semua pod restart massal = cascading failure.

```yaml
startupProbe:  { httpGet: { path: /healthz, port: 8080 }, failureThreshold: 30, periodSeconds: 2 }
readinessProbe:{ httpGet: { path: /readyz,  port: 8080 }, periodSeconds: 5, failureThreshold: 2 }
livenessProbe: { httpGet: { path: /healthz, port: 8080 }, periodSeconds: 10, failureThreshold: 3 }
```

### 2.3 requests / limits & QoS

- **requests**: jatah yang dijamin node sediakan → dipakai scheduler buat nempatin Pod.
- **limits**: batas atas → kalau CPU dilewati, di-throttle; kalau memory dilewati, **OOMKilled**.
- QoS: `Guaranteed` (req=lim), `Burstable` (req≠lim), `BestEffort` (tanpa keduanya). Saat node kehabisan memori, **BestEffort mati duluan**, lalu Burstable.

```yaml
resources:
  requests: { cpu: 100m, memory: 128Mi }
  limits:   { cpu: 500m, memory: 256Mi }
```

### 2.4 Rolling update & rollback

`maxSurge:1, maxUnavailable:0` = nambah 1 pod baru dulu, baru matiin lama — nol downtime. Kalau probe gagal, rollout mandek (bisa di-undo). `kubectl rollout undo` balikin ke ReplicaSet sebelumnya.

### 2.5 Kapan Kubernetes overkill

Pakai K8s kalau: banyak service, butuh auto-scaling (HPA), multi-node self-healing, ada platform engineer. **Jangan** pakai buat: 1 service kecil, traffic rendah, tim 2 orang — Compose + VPS lebih murah & lebih sedikit insiden. Jawaban matang di interview: "K8s nyelesain X dengan biaya Y; di kasus ini Y belum sepadan."

### 2.6 Fakta versi & rekomendasi 2025–2026

- **Kubernetes 1.34 rilis 27 Agustus 2025**; `spec.podReplacementPolicy` stable, dukungan containerd 1.x berakhir di v1.35.
- **Ingress NGINX di-retire**: maintenance best-effort cuma sampai Maret 2026. Rekomendasi resmi: pindah ke **Gateway API**.
- **Pod Security Admission** (sejak 1.25) gantikan PodSecurityPolicy — terapkan `restricted`.

---

### 2.7 Pod Security Admission (PSA) restricted

Sejak Kubernetes 1.25, **PodSecurityPolicy** udah dihapus dan diganti **Pod Security Admission** yang built-in. PSA nerapin satu dari tiga level ke namespace: `privileged` (bebas), `baseline` (cegah yang jelas berbahaya), dan `restricted` (paling ketat). Di lab ini kita targetkan `restricted` — itu sebabnya `deployment.yaml` punya `runAsNonRoot: true`, `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`, dan `capabilities: { drop: ["ALL"] }`.

Kenapa `restricted` penting: container yang jalan sebagai root dengan filesystem tulis-able itu ladang subur buat escape ke node. `readOnlyRootFilesystem` juga maksa aplikasi nulis ke volume yang disengaja, bukan sembarang tmp — memudahkan deteksi anomali. Kalau `kubectl apply` lo ditolak PSA dengan pesan "must not set securityContext.capabilities", berarti manifest lo masih `privileged`/lewat `baseline`; perbaiki `securityContext` sampai lolos `restricted`. Ini bukan hiasan — ini lapisan pertahanan pas satu dependency atau image bocor.

## 3. Mekanisme di balik layar

Gimana `kubectl apply` jadi Pod jalan, step by step:

1. **`kubectl apply`** kirim manifest ke **API Server**.
2. **Deployment controller** bikin/update **ReplicaSet** sesuai strategi.
3. **ReplicaSet controller** bikin **Pod** sampai jumlah = `replicas`.
4. **Scheduler** nempatin Pod ke node yang cukup `requests`.
5. **kubelet** di node jalanin container, jalankan **startup → readiness → liveness** probe.
6. Pod `Ready` → **Endpoint controller** masukin ke **Service** (DNS `svc/sre-api`).
7. Saat `kubectl delete pod`, ReplicaSet langsung bikin pengganti (self-healing).
8. Saat `set image`, Deployment bikin ReplicaSet baru & rolling update bertahap.

---

## 4. Glosarium mini

| Istilah | Arti singkat |
| --- | --- |
| Pod | Satuan terkecil, 1+ container share network/storage |
| Deployment | Pengelola ReplicaSet + cara rollback/rollout |
| ReplicaSet | Jaga N Pod tetap jalan (desired state) |
| Service | DNS stabil yang nemuin Pod aktif |
| Probe | Cek kesehatan (startup/readiness/liveness) |
| requests | Jatah resource dijamin (buat scheduling) |
| limits | Batas atas resource (OOMKilled kalau melampaui) |
| QoS | Kelas Pod saat node kehabisan memori |
| HPA | Horizontal Pod Autoscaler (scale by metric) |
| PDB | Pod Disruption Budget (jaga jumlah tersedia saat evict) |
| PSA | Pod Security Admission (kebijakan keamanan Pod) |
| Gateway API | Pengganti Ingress, rekomendasi resmi 2025+ |

---

## 5. Bahan bacaan wajib

1. **Kubernetes Basics** — https://kubernetes.io/docs/tutorials/kubernetes-basics/
2. **Kubernetes v1.34 Release** — https://kubernetes.io/blog/2025/08/27/kubernetes-v1-34-release/
3. **Configure Probes** — https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
4. **Gateway API** — https://gateway-api.sigs.k8s.io/ (pengganti Ingress NGINX)
5. **Terraform/OpenTofu Docs** — https://developer.hashicorp.com/terraform/tutorials (konteks IaC lab)
6. **SRE Workbook** — https://sre.google/workbook/ (capacity & reliability konteks)

---

## 6. Self-check

1. Bedanya Pod, ReplicaSet, dan Deployment — siapa yang pegang desired state?
2. Kenapa livenessProbe nggak boleh cek database, tapi readinessProbe boleh?
3. Apa yang terjadi kalau Pod kehabisan memory melewati `limits`?
4. Bedanya `Guaranteed` vs `BestEffort` QoS saat node pressure?
5. `maxSurge:1, maxUnavailable:0` artinya apa buat availabilitas?
6. Kapan K8s itu overkill, dan kapan wajib?
7. Kenapa Ingress NGINX nggak disarankan buat cluster baru di 2025+?

---

## 7. Yang sering salah paham

1. **"Limits tanpa requests oke."** Pod jadi `BestEffort` → korban pertama saat node kehabisan memori. Selalu set keduanya.
2. **"Liveness = readiness."** Kalau DB down, semua pod restart massal = cascading failure. Pisah.
3. **"Nggak perlu startupProbe."** App booting lama → liveness timeout → CrashLoopBackOff palsu.
4. **"Image `:latest` praktis."** `rollout undo` nggak deterministik — lo nggak tahu persis versi kemarin.
5. **"Ingress NGINX masih aman buat cluster baru."** Udah di-retire, nggak dapat security fix setelah Maret 2026. Pakai Gateway API.
6. **"K8s selalu lebih baik dari Compose."** Untuk 1 service sepi, K8s = overhead + insiden lebih banyak. Tahu kapan nggak pakai.
