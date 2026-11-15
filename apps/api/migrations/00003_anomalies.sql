-- +goose Up
-- One row per detected deviation: a site, a metric, a UTC day. goal_id is 0 except for
-- the goal_conversions metric. breakdown is JSON text, the top contributing path,
-- referrer, country and device values as counts. Rows go with their site.

CREATE TABLE anomalies (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    site_id bigint NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    metric text NOT NULL,
    goal_id bigint NOT NULL DEFAULT 0,
    day timestamp with time zone NOT NULL,
    direction text NOT NULL,
    value double precision NOT NULL,
    baseline double precision NOT NULL,
    mad double precision NOT NULL,
    score double precision NOT NULL,
    change_pct double precision,
    weeks integer NOT NULL,
    breakdown text NOT NULL DEFAULT '{}',
    detected_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT anomalies_unique UNIQUE (site_id, metric, goal_id, day)
);

CREATE INDEX idx_anomalies_site_day ON anomalies (site_id, day DESC);

-- +goose Down
DROP TABLE anomalies;
