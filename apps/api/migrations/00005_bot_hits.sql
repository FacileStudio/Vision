-- +goose Up
-- Requests classified as bots, crawlers or automated browsers, counted per site, UTC day,
-- category, bot name and path. They never become pageviews or sessions. No address and no
-- visitor id is kept. interacted marks a session whose visitor made a real gesture.

CREATE TABLE bot_hits (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    site_id bigint NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    day timestamp with time zone NOT NULL,
    category text NOT NULL,
    name text NOT NULL,
    path text NOT NULL,
    hits bigint NOT NULL DEFAULT 0,
    CONSTRAINT bot_hits_unique UNIQUE (site_id, day, category, name, path)
);

CREATE INDEX idx_bot_hits_site_day ON bot_hits (site_id, day DESC);

ALTER TABLE visitor_sessions ADD COLUMN interacted boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE visitor_sessions DROP COLUMN interacted;
DROP TABLE bot_hits;
