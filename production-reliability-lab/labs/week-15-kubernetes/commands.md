# Perintah — Week 15 · Kubernetes

Format: perintah → output → interpretasi.

## Hari 2 (kondisi normal)

```bash
# 1. Buat cluster lokal (kind, node v1.34)
$ kind create cluster --name sre --image kindest/node:v1.34.0
# output: Creating cluster "sre" ... ✓
# interpretasi: cluster 1 node lokal siap buat lab.

# 2. Cek node
$ kubectl get nodes
# output: sre-control-plane   Ready   control-plane   1m
# interpretasi: node sehat.

# 3. Apply deployment + service
$ kubectl apply -f kubernetes/
# output: deployment.apps/sre-api created / service/sre-api created / hpa created
# interpretasi: objek masuk ke desired state.

# 4. Rollout status
$ kubectl rollout status deployment/sre-api
# output: deployment "sre-api" successfully rolled out
# interpretasi: 3 replica Ready, nol downtime.

# 5. Lihat pod (3 replica)
$ kubectl get pods
# output: sre-api-7d9-x1 Running 3/3 30s / ... (3 baris)
# interpretasi: desired 3 tercapai.

# 6. Port-forward & probe
$ kubectl port-forward svc/sre-api 8080:80 &
$ curl -fsS http://localhost:8080/readyz
# output: ok
# interpretasi: readiness lolos, trafik tembus.
```

## Hari 3 (eksperimen kegagalan)

```bash
# 7. Self-healing
$ kubectl delete pod sre-api-7d9-x1
$ kubectl get pods
# output: pod baru sre-api-7d9-y2 muncul, tetap 3
# interpretasi: ReplicaSet ganti pod yang dihapus (desired state).

# 8. Scale
$ kubectl scale deployment sre-api --replicas=5
$ kubectl get pods
# output: 5 pod Running
# interpretasi: jumlah replica naik.

# 9. Rollout gagal (image jelek)
$ kubectl set image deployment/sre-api api=ghcr.io/<user>/sre-api:sha-bad123
$ kubectl rollout status deployment/sre-api
# output: ...waiting... / Error: progress deadline exceeded
# interpretasi: probe image baru gagal, rollout mandek.

# 10. Rollback
$ kubectl rollout undo deployment/sre-api
$ kubectl rollout status deployment/sre-api
# output: successfully rolled out
# interpretasi: balik ke ReplicaSet sehat.

# 11. Lihat history
$ kubectl rollout history deployment/sre-api
# output: REVISION 1 (sehat) / 2 (gagal)
# interpretasi: tahu versi mana yang dipakai.

# 12. Log pod crash
$ kubectl logs sre-api-xxx --previous
# output: panic: ... / exited 1
# interpretasi: penyebab crash di container sebelumnya.

# 13. Describe (sumber kebenaran)
$ kubectl describe pod sre-api-xxx
# output: Events: ... Back-off restarting failed container
# interpretasi: lihat kenapa pod CrashLoopBackOff.

# 14. Resource usage
$ kubectl top pod --sort-by=memory
# output: sre-api-xxx  12m   140Mi
# interpretasi: cek apakah mendekati limits 256Mi.

# 15. Cek QoS & PSA
$ kubectl get pod sre-api-xxx -o jsonpath='{.status.qosClass}'
# output: Burstable
# interpretasi: karena req≠lim. Pastikan nggak BestEffort.

# 16. CLEANUP wajib
$ kubectl delete -f kubernetes/
$ kind delete cluster --name sre
# output: deleted / Deleted cluster sre
# interpretasi: lab lokal bersih, RAM bebas.
```
