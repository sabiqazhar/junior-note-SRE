// Package main — SRE lab API.
//
// Endpoint:
//
//	GET  /healthz   liveness  — cuma bilang "proses gue hidup".
//	                            TIDAK boleh ngecek dependency (lihat minggu 6).
//	GET  /readyz    readiness — cek dependency wajib (Postgres, Redis).
//	GET  /metrics   Prometheus (jangan diekspos publik).
//	GET  /orders    contoh endpoint yang kena instrumentasi.
//	GET  /slow?ms=N endpoint buat nguji timeout (minggu 4 & 10).
//
// Graceful shutdown: SIGTERM -> server berhenti nerima koneksi baru,
// request in-flight diselesaikan maksimum SHUTDOWN_TIMEOUT, lalu pool ditutup.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const shutdownTimeout = 20 * time.Second

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_server_requests_total",
			Help: "Total request HTTP",
		},
		// route = pola (/orders/{id}), BUKAN path mentah. Kardinalitas rendah.
		[]string{"method", "route", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_server_request_duration_seconds",
			Help:    "Durasi request dalam detik",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "route"},
	)
	inFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_server_requests_in_flight",
		Help: "Request yang sedang dilayani",
	})
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(os.Getenv("LOG_LEVEL")),
	}))
	slog.SetDefault(logger)

	// Mode healthcheck buat Docker HEALTHCHECK (distroless nggak ada curl/wget).
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		runHealthcheck()
	}

	prometheus.MustRegister(requestsTotal, requestDuration, inFlight)

	db := mustDB()
	defer db.Close()
	rdb := mustRedis()
	defer rdb.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz(db, rdb))
	mux.HandleFunc("GET /orders", handleOrders(db))
	mux.HandleFunc("GET /slow", handleSlow)
	mux.Handle("GET /metrics", promhttp.Handler())

	handler := instrument(otelhttp.NewHandler(mux, "sre-api"))

	srv := &http.Server{
		Addr:              ":" + envOr("APP_PORT", "8080"),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second, // cegah slowloris
	}

	// Graceful shutdown — kunci minggu 6.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	// Timeout ini HARUS lebih kecil dari terminationGracePeriodSeconds (K8s)
	// atau grace period orchestrator lain yang dipakai.
	c, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(c); err != nil {
		slog.Error("graceful shutdown gagal, paksa tutup", "err", err)
		_ = srv.Close()
	}
	slog.Info("shutdown selesai")
}

// ---------- middleware ----------

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inFlight.Inc()
		defer inFlight.Dec()

		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		// r.Pattern = "/orders/{id}" (kardinalitas rendah)
		route := r.Pattern
		if route == "" {
			route = "unmatched" // JANGAN pakai r.URL.Path mentah
		}
		status := strconv.Itoa(rw.status)

		requestDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
		requestsTotal.WithLabelValues(r.Method, route, status).Inc()

		slog.InfoContext(r.Context(), "request_done",
			"method", r.Method,
			"route", route,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// ---------- handler ----------

// handleHealthz cuma njawab "proses hidup". Jangan tambahin cek dependency di sini.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleReadyz cek dependency wajib. Kalau dependency mati, 503 — itu wajar dan
// lebih baik daripada ngasih 200 lalu request gagal di tengah jalan.
func handleReadyz(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(c); err != nil {
			slog.WarnContext(c, "readyz: database belum siap", "err", err)
			http.Error(w, "db not ready", http.StatusServiceUnavailable)
			return
		}
		if err := rdb.Ping(c).Err(); err != nil {
			slog.WarnContext(c, "readyz: redis belum siap", "err", err)
			http.Error(w, "redis not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	}
}

func handleOrders(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// context DITERUSKAN — ini yang bikin trace nggak putus (minggu 11).
		rows, err := db.QueryContext(r.Context(),
			`SELECT id, total FROM orders ORDER BY created_at DESC LIMIT 20`)
		if err != nil {
			slog.ErrorContext(r.Context(), "query orders gagal", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type order struct {
			ID    int64   `json:"id"`
			Total float64 `json:"total"`
		}
		out := []order{}
		for rows.Next() {
			var o order
			if err := rows.Scan(&o.ID, &o.Total); err != nil {
				slog.ErrorContext(r.Context(), "scan gagal", "err", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			out = append(out, o)
		}
		writeJSON(w, out)
	}
}

// handleSlow sengaja dibuat buat nguji timeout: /slow?ms=60000
func handleSlow(w http.ResponseWriter, r *http.Request) {
	ms, _ := strconv.Atoi(r.URL.Query().Get("ms"))
	if ms <= 0 || ms > 120000 {
		ms = 1000
	}
	select {
	case <-time.After(time.Duration(ms) * time.Millisecond):
		_, _ = w.Write([]byte("selesai lambat\n"))
	case <-r.Context().Done(): // client pergi, jangan terus ngerjain
	}
}

// ---------- helper ----------

// runHealthcheck dipanggil Docker Compose lewat:
//
//	test: ["CMD", "/app/sre-api", "-healthcheck"]
//
// Pakai /healthz (bukan /readyz) supaya container nggak di-restart gara-gara
// dependency yang sebentar bermasalah — lihat minggu 3 dan minggu 6.
func runHealthcheck() {
	c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(c, http.MethodGet,
		"http://127.0.0.1:"+envOr("APP_PORT", "8080")+"/healthz", nil)
	if err != nil {
		os.Exit(1)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	os.Exit(0)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func mustDB() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL kosong")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("sql.Open gagal", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db
}

func mustRedis() *redis.Client {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		slog.Error("REDIS_URL tidak valid", "err", err)
		os.Exit(1)
	}
	return redis.NewClient(opt)
}

func envOr(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
