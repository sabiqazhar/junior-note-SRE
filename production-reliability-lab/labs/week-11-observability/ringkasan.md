# Ringkasan — Week 11 · Logs dan Distributed Tracing

## Inti (5 kalimat)
1. Tiga alat jawab tiga pertanyaan: **metrics** "ada yang aneh?", **logs** "apa yang terjadi?", **traces** "di span mana waktunya habis?" — urutannya selalu metrics→logs→traces.
2. Log harus **terstruktur** (JSON via `log/slog` Go) biar bisa di-query LogQL; `trace_id` wajib di-inject ke tiap log biar nyambung ke trace.
3. Trace paling sering **putus** cuma karena lupa nerusin `context.Context` ke fungsi turunan (mis. `QueryRowContext`, bukan `QueryRow`).
4. **Tail sampling** (Alloy) simpan error+lambat 100%, baseline 1% — jangan simpan semua trace.
5. Pipeline log jalan lewat **Alloy → Loki**, trace lewat **OTel OTLP → Tempo**; Promtail sudah EOL 2 Mar 2026.

## Peta konsep
```
app (Go)
 ├─ slog JSON stdout ──> Alloy loki.source.docker ──parse──> Loki  (LogQL)
 └─ otelhttp span ─────> Alloy otelcol.receiver.otlp
                              │ tail_sampling (error/lambat 100%, 1% baseline)
                              ▼
                          Tempo  (trace by trace_id)
Grafana: metrics panel ─klik trace_id─> Loki + Tempo berdampingan
```

## Query / perintah kunci + output
| Perintah / Query | Output | Arti |
| --- | --- | --- |
| `slog.InfoContext(...,"trace_id",span.TraceID())` | log JSON ber-trace_id | korelasi log↔trace |
| `db.QueryRowContext(ctx, ...)` | span DB nyambung | cegah trace putus |
| LogQL: `{job="api"} \| json \| trace_id="abc"` | baris log tsb | cari log by trace |
| LogQL: `{job="api"} \| json \| level="error"` | error log | filter level |
| Alloy tail_sampling policy | error+lambat 100% | sampling benar |
| OTLP receiver `:4317/:4318` | trace masuk | app ekspor trace |
| Tempo trace view | pohon span + durasi | lihat span lambat |

## Failure matrix
| Gejala | Penyebab | Cek | Mitigasi |
| --- | --- | --- | --- |
| Log nggak masuk Loki | Alloy mati / socket salah | `docker compose ps alloy` | pastikan Alloy Up, docker.sock benar |
| `trace_id` nggak nemu log | lupa inject ke slog | cek log JSON ada field? | tambah `trace_id` di slog |
| Trace putus di DB | `QueryRow` tanpa ctx | cek handler pakai Context | pakai `QueryRowContext(ctx,...)` |
| Trace nggak ada di Tempo | sampling buang / OTLP gagal | cek tail_sampling, receiver | pastikan policy error/lambat |
| LogQL berantakan | level beda nama (`severity`) | `\| json` cek field | samakan key `level` |
| Masih pakai Promtail | EOL | `docker compose ps` | pakai Alloy |

## Jebakan
- Trace putus karena `context` nggak diteruskan ke fungsi turunan.
- Correlation ID ada di header tapi lupa di-inject ke `slog`.
- `trace_id` dijadikan label Loki → kardinalitas meledak.
- Tail sampling tanpa policy 100% untuk error → trace penting kebuang.
- Ngira semua semconv stabil — sebagian masih Mixed.
- Log JSON tapi `level` vs `severity` nggak konsisten → LogQL rusak.
- Masih pakai Promtail (EOL 2 Mar 2026).

## Acceptance
1. Log terstruktur JSON, bisa difilter by trace/correlation ID. ✅ LogQL #1
2. Trace nunjukin span app DAN dependency (Postgres/Redis). ✅ Tempo
3. Request lambat bisa ditelusuri tanpa nebak. ✅ lab Hari 3
4. `traceparent` diteruskan, trace nggak putus. ✅ `QueryRowContext`

## Satu kalimat
Metrics kasih tahu ada masalah, tapi cuma lewat logs (correlation ID) dan traces (context propagation) lo bisa telusuri request spesifik sampai ke span database-nya.
