# Ringkasan — Minggu 15 · Kubernetes dan Orchestration

> Satu halaman untuk dibaca ulang sebelum ujian/wawancara. Bukan pengganti bahan-bacaan.md.

## Inti 5 kalimat

1. Kubernetes menjaga **desired state**: lo bilang "3 replica", dan ReplicaSet bikin pengganti tiap Pod yang mati (self-healing).
2. Tiga probe tugasnya beda — **liveness cuma cek proses**, readiness boleh cek dependency, startup kasih waktu booting.
3. **requests/limits** menentukan QoS: tanpa keduanya Pod jadi `BestEffort` dan mati duluan saat node kehabisan memori.
4. **Rolling update** (`maxSurge:1, maxUnavailable:0`) = nol downtime; `rollout undo` balikin ke versi sebelumnya.
5. K8s itu **mahal kognitif** — pakai buat banyak service/auto-scale/self-healing, jangan buat 1 service sepi (Compose + VPS lebih murah).

## Peta konsep

```
kubectl apply ──> API Server
   │
   ├─ Deployment ──> ReplicaSet ──> Pod ──> kubelet ──> container
   │                    │              │
   │                    │              └─ probes: startup→readiness→liveness
   │                    │              └─ resources: requests/limits → QoS
   │                    └─ rollout status / undo
   └─ Service (DNS svc/sre-api) ── Endpoint controller ──> Pod Ready
        HPA ── cpu util 70% ──> scale replica
        PDB ── minAvailable:2 ──> jaga saat evict
```

## Perintah kunci + output

```bash
# 1. Buat cluster lokal
$ kind create cluster --name sre --image kindest/node:v1.34.0
# output: Creating cluster "sre" ... done.
# interpretasi: cluster 1 node lokal siap.

# 2. Cek node
$ kubectl get nodes
# output: sre-control-plane   Ready   control-plane   1m
# interpretasi: node sehat.

# 3. Apply manifest
$ kubectl apply -f kubernetes/
# output: deployment.apps/sre-api created / service/sre-api created / hpa created
# interpretasi: objek masuk desired state.

# 4. Tunggu rollout
$ kubectl rollout status deployment/sre-api
# output: deployment "sre-api" successfully rolled out
# interpretasi: 3 replica Ready, nol downtime.

# 5. Lihat pod (3 replica)
$ kubectl get pods
# output: sre-api-xxx   Running   3/3   1m
# interpretasi: desired 3 tercapai.

# 6. Port-forward & cek
$ kubectl port-forward svc/sre-api 8080:80
$ curl -fsS http://localhost:8080/readyz
# output: ok
# interpretasi: readiness lolos, trafik tembus.

# 7. Self-healing
$ kubectl delete pod <pod-name>
$ kubectl get pods
# output: pod baru muncul otomatis (tetap 3)
# interpretasi: ReplicaSet ganti pod yang dihapus.

# 8. Scale
$ kubectl scale deployment sre-api --replicas=5
# output: deployment.apps/sre-api scaled
# interpretasi: jumlah replica naik jadi 5.

# 9. Rollout gagal (image jelek)
$ kubectl set image deployment/sre-api api=ghcr.io/<user>/sre-api:sha-bad123
$ kubectl rollout status deployment/sre-api
# output: Error: progress deadline exceeded
# interpretasi: probe image baru gagal, rollout mandek.

# 10. Rollback
$ kubectl rollout undo deployment/sre-api
# output: deployment.apps/sre-api rolled back
# interpretasi: balik ke ReplicaSet sehat.

# 11. Log crash
$ kubectl logs <pod> --previous
# output: panic: ... exited 1
# interpretasi: penyebab crash container sebelumnya.

# 12. Describe (sumber kebenaran)
$ kubectl describe pod <pod>
# output: Events: Back-off restarting failed container
# interpretasi: lihat kenapa CrashLoopBackOff.

# 13. Resource usage
$ kubectl top pod --sort-by=memory
# output: sre-api-xxx  12m   140Mi
# interpretasi: cek apakah mendekati limits 256Mi.

# 14. Cek QoS
$ kubectl get pod <pod> -o jsonpath='{.status.qosClass}'
# output: Burstable
# interpretasi: req≠lim. Pastikan nggak BestEffort.

# 15. CLEANUP wajib
$ kubectl delete -f kubernetes/ && kind delete cluster --name sre
# output: deleted / Deleted cluster sre
# interpretasi: lab lokal bersih, RAM bebas.
```

## Failure matrix

| Gejala | Penyebab | Cek | Penyelesaian |
| --- | --- | --- | --- |
| CrashLoopBackOff | app boot lama, liveness timeout | `kubectl describe pod` Events | tambah startupProbe |
| Pod nggak dapat trafik | readiness gagal (path salah) | `kubectl get ep` | perbaiki `/readyz` |
| OOMKilled | memory > limits | `kubectl top pod`, logs `--previous` | naik limits/requests |
| Rollout mandek | probe image baru gagal | `kubectl rollout status` | `rollout undo` |
| Restart massal | liveness cek DB | events | pisah liveness/readiness |
| PSA rejected | container root/`latest` | `kubectl apply` error | non-root + tag immutable |

## Jebakan

- `limits` tanpa `requests` (atau sebaliknya) → QoS Burstable/BestEffort.
- Nggak ada startupProbe buat app boot lama → CrashLoopBackOff palsu.
- Liveness = readiness → restart massal saat overload.
- Masih pakai Ingress NGINX buat cluster baru (nggak dapet security fix).
- Image `:latest` → rollback nggak deterministik.

## Cheat sheet cepat

| Mau ngapain | Perintah |
| --- | --- |
| Cluster lokal | `kind create cluster --name sre --image kindest/node:v1.34.0` |
| Deploy | `kubectl apply -f kubernetes/` |
| Tunggu rollout | `kubectl rollout status deployment/sre-api` |
| Self-heal test | `kubectl delete pod <pod>` |
| Scale | `kubectl scale deployment sre-api --replicas=5` |
| Rollback | `kubectl rollout undo deployment/sre-api` |
| Debug | `kubectl logs --previous` / `describe pod` |
| Bersih | `kind delete cluster --name sre` |

## Tanya jawab kilat

- **Q: Siapa pegang desired state?** A: ReplicaSet (diatur Deployment).
- **Q: Liveness boleh cek DB?** A: TIDAK — restart massal kalau DB down.
- **Q: Tanpa requests/limits?** A: BestEffort, mati duluan saat node pressure.
- **Q: Kapan K8s overkill?** A: 1 service sepi, tim kecil → Compose + VPS.
- **Q: Ingress NGINX 2025?** A: di-retire; pakai Gateway API.

## Acceptance

1. Deployment + Service jalan dengan image API lab (tag immutable, bukan `:latest`).
2. Pod yang dihapus diganti otomatis (desired state).
3. Rollout, rollback, probe failure bisa diamati & dijelaskan.
4. Non-root + `readOnlyRootFilesystem` + requests/limits; PSA `restricted` aktif.

## Studi kasus mini

- **Kasus A:** `livenessProbe` cek DB. Pas DB maintenance, semua pod restart massal → cascading failure. Pelajaran: liveness cuma cek proses.
- **Kasus B:** tanpa `startupProbe`, app boot 40s → liveness (10s×3) timeout → CrashLoopBackOff palsu. Pelajaran: kasih startupProbe.
- **Kasus C:** `limits` ada, `requests` nggak → Pod `BestEffort`, mati duluan saat node pressure. Pelajaran: set keduanya.

## Istilah kunci (flashcard)

- **Pod** = satuan terkecil, 1+ container.
- **ReplicaSet** = jaga N pod (desired state).
- **Deployment** = pengelola rollout/rollback.
- **Service** = DNS stabil ke pod aktif.
- **Probe** = startup / readiness / liveness.
- **requests** = jatah dijamin (scheduling).
- **limits** = batas atas (OOMKilled kalau lewat).
- **QoS** = kelas saat node pressure.
- **HPA** = auto-scale by metric.
- **PSA** = kebijakan keamanan pod (restricted).

## Pertanyaan wawancara

1. Siapa pegang desired state?
2. Kenapa liveness nggak cek DB?
3. Tanpa requests/limits gimana QoS-nya?
4. Kapan K8s overkill?
5. Kenapa Ingress NGINX di-retire?

## Pola yang sering kelewat

- `limits` tanpa `requests`.
- Nggak ada startupProbe.
- Liveness = readiness.
- Image `:latest`.

## Cek cepat sebelum deploy

| Cek | Perintah |
| --- | --- |
| Rollout | `kubectl rollout status` |
| Self-heal | `kubectl delete pod` |
| Rollback | `kubectl rollout undo` |
| PSA | `kubectl apply` (tanpa warning) |

## Satu kalimat

> Kubernetes itu manajer shift yang jaga jumlah pelayan tetap 3: Pod mati diganti, rollout bertahap, tapi cuma sepadan kalau kompleksitasnya emang lo butuh.
