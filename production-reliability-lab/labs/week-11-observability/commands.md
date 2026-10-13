# Perintah — Week 11 · Logs dan Distributed Tracing

Format: perintah → output → interpretasi.
Fokus: structured log (slog), LogQL, Alloy→Loki, OTel→Tempo, tail sampling.

## Persiapan

### 1. Cek stack observability jalan
```
docker compose ps
```
```
NAME             IMAGE                  STATUS
sre-lab-api-1    sre-lab-api            Up
sre-lab-loki-1   grafana/loki:latest    Up
sre-lab-tempo-1  grafana/tempo:latest   Up
sre-lab-alloy-1  grafana/alloy:latest   Up
```
Interpretasi: `loki`, `tempo`, `alloy` Up. Alloy = pipeline (bukan Promtail).

### 2. Cek app log keluar sebagai JSON
```
docker compose logs api --since 1m | head -2
```
```
{"time":"...","level":"info","msg":"request_done","route":"/orders","status":200,"duration_ms":12,"trace_id":"4b2c1a..."}
```
Interpretasi: `slog` JSON handler jalan. Ada `trace_id` → bisa dikorelasikan.

## Hari 2 (structured log & trace)

### 3. LogQL: lihat 5 log terstruktur
```logql
{job="api"} | json | limit 5
```
```
{level="info", route="/orders", status="200", duration_ms="12", trace_id="4b2c..."}
```
Interpretasi: Alloy parse JSON, field bisa difilter. Level jadi label.

### 4. LogQL: filter error saja
```logql
{job="api"} | json | level="error"
```
```
{"level":"error","msg":"query orders gagal","err":"..."}
```
Interpretasi: cuma log error. Berguna pas insiden.

### 5. LogQL: hitung log per route
```logql
{job="api"} | json | count by (route) ()
```
```
{route="/orders"} 1842
{route="/slow"}   37
```
Interpretasi: volume log per route, tanpa bikin metrik tambahan.

### 6. Verifikasi Alloy terima log (bukan Promtail)
```
docker compose logs alloy --since 1m | grep -i "loki.write\|docker"
```
```
level=info msg="sent batch" ...
```
Interpretasi: Alloy forward ke Loki. Promtail nggak ada di stack.

### 7. Cek trace masuk Tempo
```
curl -sS http://localhost/orders >/dev/null
```
Lalu di Grafana Explore → Tempo → cari trace terbaru.
Interpretasi: trace ada dgn span `sre-api` + child span DB.

## Hari 3 (telusuri request lambat)

### 8. Ambil trace_id dari log
```logql
{job="api"} | json | route="/orders" | limit 1
```
```
{"trace_id":"4b2c1a9f...","route":"/orders","duration_ms":1200}
```
Interpretasi: request ini 1,2s. Ambil `trace_id` buat langkah berikutnya.

### 9. Cari semua log untuk 1 request by trace_id
```logql
{job="api"} | json | trace_id="4b2c1a9f..."
```
```
{"msg":"request_done","route":"/orders","duration_ms":1200}
{"msg":"query orders gagal"...}  (kalau error)
```
Interpretasi: seluruh log request itu kekumpul. Inilah korelasi.

### 10. Buka trace di Tempo by trace_id
```
# Grafana Explore -> Tempo -> Search -> trace_id=4b2c1a9f...
```
```
Span sre-api (HTTP)        1200ms
  └─ Span db.query         1100ms   <-- 90% waktu di sini
```
Interpretasi: "oh, ternyata" — 90% waktu habis di span database, bukan app.

### 11. Verifikasi context diteruskan (cegah trace putus)
```bash
grep -n "QueryRowContext\|QueryRow(" app/main.go
```
```
194: rows, err := db.QueryContext(r.Context(), `SELECT ...`)
```
Interpretasi: pakai `QueryContext` → span DB nyambung ke span HTTP. Benar.

### 12. Verifikasi tail sampling di Alloy
```
grep -n "status_code\|latency\|probabilistic" observability/otel/config.alloy
```
```
policy { status_code { status_codes = ["ERROR"] } }
policy { latency { threshold_ms = 500 } }
policy { probabilistic { sampling_percentage = 1 } }
```
Interpretasi: error+lambat 100% disimpan, baseline cuma 1%. Nggak buang trace penting.

## Hari 4 (verifikasi pipeline)

### 13. Cek receiver OTLP Alloy aktif
```
docker compose logs alloy --since 1m | grep -i "otlp\|4317"
```
```
level=info msg="OTLP receiver started" endpoint=0.0.0.0:4317
```
Interpretasi: app bisa ekspor trace lewat OTLP gRPC ke Alloy.

### 14. Cek Loki nerima (label level)
```logql
{job="api"} | json | count by (level) ()
```
```
{level="info"} 1900
{level="error"} 11
```
Interpretasi: `level` jadi label (sesuai `stage.labels` di Alloy). `trace_id`
tetap field, bukan label (kardinalitas aman).

### 15. Cek Tempo writable dari Alloy
```
docker compose logs alloy --since 2m | grep -i "tempo\|export"
```
```
level=info msg="exported spans" exporter=otlp/tempo count=12
```
Interpretasi: 12 span diekspor ke Tempo. Pipeline trace jalan end-to-end.

### 16. Pastikan Promtail TIDAK ada
```
docker compose ps | grep -i promtail || echo "promtail: tidak ada"
```
```
promtail: tidak ada
```
Interpretasi: sesuai kebijakan (Promtail EOL 2 Mar 2026), pakai Alloy.

### 17. LogQL: cari request lambat (>500ms) by duration
```logql
{job="api"} | json | duration_ms >= 500
```
```
{"route":"/slow","duration_ms":"2000"}
```
Interpretasi: temukan request lambat langsung dari log terstruktur.

## Cara buktikan acceptance minggu 11
- Log JSON & filter by trace_id → perintah 3, 9.
- Trace span app+DB nyambung → perintah 7, 10, 11.
- Pipeline Alloy (bukan Promtail) → perintah 6, 16.
- Tail sampling benar → perintah 12.
