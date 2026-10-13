# Minggu 11 · Logs dan Distributed Tracing — Bahan Bacaan

> "Metrics bilang ada api. Logs bilang di ruangan mana. Traces bilang api itu menjalar lewat kabel mana."

Dokumen ini berdiri sendiri. Baca dari atas ke bawah, kerjakan `lab.md` setelahnya.

---

## 1. Kenapa minggu ini penting

Minggu lalu lo sudah tahu *ada* masalah (metrics: error naik, p95 melonjak). Tapi metrics nggak bisa jawab **request mana** yang bermasalah, apalagi **di titik mana** waktunya habis. Bayangin gedung kebakaran: alarm (metrics) bilang "kebakaran!", tapi lo butuh CCTV (logs) buat tahu ruangan mana yang asapnya paling tebal, dan lo butuh denah jalur kabel (traces) buat tahu api itu menjalar dari panel listrik ke mana. Tiga alat ini jawab tiga pertanyaan beda, dan urutannya selalu sama: **metrics → logs → traces**.

Di lab ini aplikasi lo sudah pakai `log/slog` bawaan Go (sejak 1.21) buat log JSON terstruktur, dan `otelhttp` buat bikin trace otomatis. Tapi trace paling sering **putus** cuma karena lo lupa nerusin `context.Context` ke fungsi turunan — dan correlation ID paling sering **ada tapi nggak masuk log**. Dua kesalahan ini bikin "oh, ternyata" minggu ini nggak kejadian. Makanya lo bakal praktik: bikin request lambat, ambil `trace_id`, cari log-nya di Loki, buka trace di Tempo, dan jawab "90% waktunya habis di span database".

Satu lagi yang wajib lo tahu: **Promtail sudah EOL 2 Maret 2026.** Pipeline log lab ini jalan lewat **Grafana Alloy** (`observability/otel/config.alloy`). Jangan mulai lab baru pakai Promtail.

---

## 2. Konsep inti

### 2.1 Tiga alat, tiga pertanyaan
| Alat | Pertanyaan | Analogi |
| --- | --- | --- |
| Metrics | "Ada yang aneh nggak?" | Alarm kebakaran |
| Logs | "Persisnya apa yang terjadi?" | CCTV |
| Traces | "Di titik mana waktunya habis?" | Peta rute kurir |

**Kapan pakai mana:**
- Pakai **metrics** buat deteksi & SLO (murah, agregat, selalu nyala).
- Pakai **logs** kalau lo sudah tahu *kira-kira* request mana & mau baca detail (error stack, payload).
- Pakai **traces** kalau lo tahu ada latency tapi nggak tahu *di span mana* (app vs DB vs Redis).

### 2.2 Structured logging dengan `log/slog` (Go)
Jangan log string mentah. Pakai `slog` JSON handler biar tiap field jadi key:value yang gampang di-query LogQL.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
slog.SetDefault(logger)

slog.InfoContext(r.Context(), "request_done",
    "method", r.Method,
    "route",  r.Pattern,
    "status", rw.status,
    "duration_ms", time.Since(start).Milliseconds(),
    "trace_id", span.SpanContext().TraceID().String(),
)
```

### 2.3 Correlation ID & W3C Trace Context
Biara log dan trace nyambung, lo butuh ID yang sama. Standarnya **W3C Trace Context** (`traceparent`) — dipakai otomatis oleh OTel dan berlaku lintas service & vendor. `trace_id` dari span harus di-inject ke tiap log (`slog` di atas).

### 2.4 Trace / Span & context propagation
Satu **trace** = satu request lintas service. Di dalamnya ada **span** (tiap langkah: HTTP handler, query DB, call Redis). Kunci: **teruskan `context.Context`** ke tiap fungsi turunan, kalau nggak trace putus.

```go
// SALAH — trace putus
db.QueryRow("INSERT ...")                       // context nggak diteruskan

// BENAR
db.QueryRowContext(r.Context(), "INSERT ...")   // span DB nyambung
```

### 2.5 Sampling — jangan simpan semua
Trace mahal. **Tail sampling** di Alloy memutuskan SETELAH trace lengkap: simpan error + lambat 100%, sisanya baseline 1%.

```river
otelcol.processor.tail_sampling "default" {
  policy { status_code { status_codes = ["ERROR"] } }   // SELALU simpan error
  policy { latency { threshold_ms = 500 } }             // simpan lambat
  policy { probabilistic { sampling_percentage = 1 } }  // baseline 1%
}
```

### 2.6 Pipeline log: Alloy (bukan Promtail)
`loki.source.docker` kumpulin log container → `loki.process.json` parse JSON
(level jadi label, `trace_id` tetap field) → `loki.write` kirim ke Loki.
`trace_id` **jangan** jadi label (kardinalitas meledak) — cukup field yang bisa
dicari.

### 2.7 Tempo sebagai trace backend
Alloy ekspor trace lewat OTLP ke **Tempo**. Di Grafana lo buka trace by
`trace_id`, lihat pohon span + durasi tiap span.

---

## 3. Mekanisme di balik layar

1. **App log**: `slog` tulis JSON ke stdout, berisi `trace_id`.
2. **Alloy scrape**: `loki.source.docker` baca stdout container, parse JSON, kirim ke Loki.
3. **App trace**: `otelhttp` bikin span tiap request, `QueryRowContext` bikin child span DB. Context diteruskan → span nyambung.
4. **Alloy terima OTLP**: `otelcol.receiver.otlp` terima dari app (port 4317/4318).
5. **Tail sampling**: keputusan simpan error/lambat/baseline.
6. **Tempo simpan**: `otelcol.exporter.otlp.tempo` kirim ke Tempo.
7. **Korelasi**: di Grafana, dari panel metrics → klik trace_id → Loki (log) + Tempo (trace) terbuka berdampingan.

---

## 4. Glosarium mini

| Istilah | Arti singkat |
| --- | --- |
| Structured log | Log JSON key:value, bisa di-query |
| `slog` | Logger bawaan Go (sejak 1.21) |
| Correlation ID | ID sama lintas log & trace |
| W3C Trace Context | Standar `traceparent` penyambung trace |
| Trace | Satu request lintas service |
| Span | Satu langkah dalam trace |
| Context propagation | Meneruskan context agar span nyambung |
| Tail sampling | Sampling setelah trace lengkap |
| Alloy | Pipeline log/trace (ganti Promtail) |
| Promtail | Collector log lama, EOL 2 Mar 2026 |
| Tempo | Backend penyimpan trace (Grafana) |
| LogQL | Bahasa query Loki |

---

## 5. Bahan bacaan wajib (primer)

1. OpenTelemetry — Docs (traces, logs, metrics): https://opentelemetry.io/docs/
2. OpenTelemetry — Trace Context (W3C): https://opentelemetry.io/docs/specs/otel/trace/trace-context/
3. OpenTelemetry — Sampling: https://opentelemetry.io/docs/concepts/sampling/
4. Grafana — Loki LogQL: https://grafana.com/docs/loki/latest/query/logql/
5. Grafana — Alloy (pengganti Promtail): https://grafana.com/docs/alloy/latest/
6. Grafana — Tempo: https://grafana.com/docs/tempo/latest/

---

## 6. Self-check (harus bisa dijawab tanpa Google)

1. Metrics, logs, traces jawab pertanyaan apa bedanya?
2. Kenapa `trace_id` jangan dijadikan label Loki?
3. Apa penyebab paling umum trace putus?
4. Apa bedanya head sampling dan tail sampling?
5. Kenapa Promtail nggak dipakai di lab ini?
6. Bagaimana `context.Context` membuat span DB nyambung ke span HTTP?

---

## 7. Yang sering salah paham

- **"Metrics cukup, log nggak perlu."** Metrics agregat — nggak bisa tunjukin request spesifik yang gagal.
- **"Trace otomatis selalu nyambung."** Salah. Lupa `QueryRowContext(ctx,...)` → span DB putus, lo nggak tahu waktu habis di DB.
- **"Correlation ID ada, berarti nyambung."** Sering ID ada di header tapi lupa di-inject ke `slog` → nggak bisa cari log by trace_id.
- **"Simpan semua trace."** Mahal & nggak perlu. Tail sampling: error+lambat 100%, baseline 1%.
- **"Semua semconv sudah stabil."** Sebagian masih Mixed — cek `OTEL_SEMCONV_STABILITY_OPT_IN`.
- **"Promtail masih oke."** EOL 2 Maret 2026 — pakai Alloy.
