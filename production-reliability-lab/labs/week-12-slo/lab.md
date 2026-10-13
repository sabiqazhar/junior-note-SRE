# Lab — Week 12 · SLI, SLO, Error Budget, dan Incident Response

> Target: mikir kayak SRE, bukan cuma operator monitoring.
> Waktu: 4 hari × 2 jam. Stack: `api`, `prometheus` (rules/slo.yml), `k6` (load).

## Persiapan (wajib)
```
docker compose ps
```
Pastikan `sre-lab-api-1` & `sre-lab-prometheus-1` Up. Buka `rules/slo.yml`
sudah ter-mount (`rule_files: /etc/prometheus/rules/*.yml`). Di Prometheus UI →
Status → Rules, lihat `job:slo_errors_per_request:ratio_rate*` dan alert
`HighErrorBudgetBurn`/`SlowErrorBudgetBurn` terdaftar (mungkin `inactive`).

---

## Hari 1 · Teori (2 jam)
- Baca `bahan-bacaan.md` seksi 1–7 + `docs/slo.md` & `docs/incident-response.md`.
- Catat di `notes.md`: tabel SLI/SLO/SLA + rumus burn rate + policy budget.
- Hitung manual: SLO 99,5%, error ratio 1 jam = 3,6% → burn rate berapa?

**Yang harus terjadi:** lo bisa hitung burn rate & jelasin kenapa diukur dari proxy.

**Kalau gagal:** balik ke `bahan-bacaan.md` 2.1, 2.4.

**Catat notes:** tulis 1 contoh SLO lab (availability 99,5%) + eksklusinya.

---

## Hari 2 · Praktik (2 jam) — definisikan & uji rules
1. Validasi rule dengan promtool:
   ```
   promtool check rules observability/prometheus/rules/slo.yml
   ```
   **Yang harus terjadi:** `SUCCESS: N rules found`.
   **Kalau gagal:** cek indent YAML & nama metrik (`http_server_requests_total`).

2. (Opsional) buat unit test:
   ```
   promtool test rules observability/prometheus/rules/slo_test.yml
   ```
   **Yang harus terjadi:** semua test case `SUCCESS`.

3. Cek recording rule jalan di Prometheus:
   ```
   curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_errors_per_request:ratio_rate5m"
   ```
   **Yang harus terjadi:** nilai error ratio (mis. ~0).

**Catat notes:** simpan output `promtool check` ke `evidence/`.

---

## Hari 3 · Eksperimen (2 jam) — simulasi insiden + burn-rate
1. Traffic normal dulu (k6 ringan / curl loop):
   ```
   for i in $(seq 1 200); do curl -sS -o /dev/null http://localhost/orders; done
   ```
2. Rusak sesuatu — matiin Redis (dependency) atau bikin query lambat:
   ```
   docker compose stop redis
   ```
   **Yang harus terjadi:** error ratio naik, `HighErrorBudgetBurn`/`SlowErrorBudgetBurn`
   firing (lihat Prometheus → Alerts). Catat **waktu alert nyala**.
   **Kalau gagal:** cek recording rule aktif & burn-rate threshold (14.4×0.005).

3. Catat **timeline** (UTC): mulai rusak → alert firing → mitigasi → pulih.
   Hitung error budget terpakai:
   ```
   curl -fsS "http://localhost:9090/api/v1/query?query=job:slo_errors_per_request:ratio_rate1h"
   ```

4. **CARA KEMBALI NORMAL (wajib):**
   ```
   docker compose start redis        # kembalikan dependency yang di-stop
   # atau rollback deploy kalau rusaknya dari versi baru
   curl -fsS http://localhost:9090/api/v1/query?query=up{job=%22api%22}
   ```
   **Yang harus terjadi:** `up{job="api"}` = 1, error ratio turun ke baseline,
   alert kembali `inactive`. Bukan rollback image sembarangan — pulihkan dependency
   / rollback versi, lalu verifikasi dengan angka (bukan perasaan).

**Catat notes:** isi tabel timeline + berapa % budget terpakai.

---

## Hari 4 · Review (2 jam) — postmortem blameless
1. Salin `postmortems/incident-001.md` jadi `incident-002.md`, isi dengan data
   simulasi Hari 3: impact, timeline, root cause (1 kalimat mekanisme), 5-whys,
   what went well/wrong, action item **ber-owner & tenggat**.
2. Pastikan blameless: nggak ada nama orang sebagai penyebab.
3. Commit `incident-002.md` + `commands.md` + `evidence/`.

**Catat notes:** 1 pelajaran yang bisa dipakai orang lain.

---

## Daftar periksa (checklist)
- [ ] `promtool check rules` SUCCESS
- [ ] Recording rule `job:slo_*` aktif di Prometheus
- [ ] Alert burn-rate (14.4/6/1) terdaftar
- [ ] Simulasi: error ratio naik → alert firing
- [ ] Timeline insiden tercatat (UTC)
- [ ] Cara kembali normal sudah dicoba & diverifikasi angka
- [ ] Postmortem blameless lengkap (owner+tenggat)
- [ ] `docs/slo.md` & `docs/incident-response.md` sudah ada

## Cara buktikan (acceptance)
1. `promtool check` output → `evidence/`.
2. Screenshot Prometheus Alerts firing → `evidence/`.
3. Timeline + % budget → `incident-002.md`.
4. Postmortem action item ber-owner → `evidence/`.
