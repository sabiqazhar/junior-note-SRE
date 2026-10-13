// k6 — skenario load test lab SRE (minggu 16).
// Jalankan: k6 run loadtest/loadtest.js
// Catatan versi: k6 v1.0 rilis Mei 2025; sekarang sudah ada v2.x. PIN versi di CI.
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const orderLatency = new Trend('order_latency_ms', true);
const cacheMiss    = new Counter('cache_miss');

export const options = {
  scenarios: {
    ramp: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50 },   // naik pelan
        { duration: '3m', target: 50 },   // tahan: ini yang dipakai buat baseline
        { duration: '1m', target: 0 },    // turun
      ],
      gracefulRampDown: '30s',
    },
  },
  // Threshold ini yang bikin load test bisa NGEGAGALKAN CI.
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99)<800'],
    http_req_failed:   ['rate<0.01'],
  },
};

const BASE = __ENV.BASE_URL || 'http://localhost';

export default function () {
  // timeout 2s = timeout budget. Lebih baik gagal cepat daripada numpuk.
  const res = http.get(`${BASE}/orders`, { timeout: '2s', tags: { endpoint: 'orders' } });

  check(res, {
    'status 200':       (r) => r.status === 200,
    'latensi < 300ms':  (r) => r.timings.duration < 300,
  });

  orderLatency.add(res.timings.duration);
  if (res.headers['X-Cache'] === 'MISS') cacheMiss.add(1);

  sleep(1);
}

export function handleSummary(data) {
  return {
    'loadtest/results.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data),
  };
}

function textSummary(d) {
  const m = d.metrics;
  return [
    '',
    '=== Ringkasan load test ===',
    `request total : ${m.http_reqs.values.count}`,
    `error rate    : ${(m.http_req_failed.values.rate * 100).toFixed(2)}%`,
    `p50 latency   : ${m.http_req_duration.values['p(50)'].toFixed(1)} ms`,
    `p95 latency   : ${m.http_req_duration.values['p(95)'].toFixed(1)} ms`,
    `p99 latency   : ${m.http_req_duration.values['p(99)'].toFixed(1)} ms`,
    `throughput    : ${m.http_reqs.values.rate.toFixed(1)} req/s`,
    '',
    'Isi tabel ini tiap kali nambah beban:',
    '| VU | req/s | p50 | p95 | p99 | error% | CPU | DB conn |',
    '',
  ].join('\n');
}
