module github.com/<user>/production-reliability-lab

go 1.25

require (
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.22.0
	github.com/redis/go-redis/v9 v9.7.3
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0
	go.opentelemetry.io/otel v1.47.0
)
