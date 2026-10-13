-- Inisialisasi skema lab. Jalan otomatis saat volume Postgres pertama kali dibuat.
-- Isi data contoh supaya endpoint /orders bisa diuji dan load test punya bahan.

CREATE TABLE IF NOT EXISTS orders (
    id          BIGSERIAL PRIMARY KEY,
    customer_id BIGINT      NOT NULL,
    total       NUMERIC(12,2) NOT NULL DEFAULT 0,
    status      TEXT        NOT NULL DEFAULT 'created',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_customer    ON orders (customer_id);

-- Data contoh. Angka kecil saja: ini lab, bukan dataset produksi.
INSERT INTO orders (customer_id, total, status)
SELECT
    (random() * 500)::bigint + 1,
    (random() * 500000 + 10000)::numeric(12,2),
    (ARRAY['created','paid','shipped'])[1 + (random() * 2)::int]
FROM generate_series(1, 1000)
ON CONFLICT DO NOTHING;
