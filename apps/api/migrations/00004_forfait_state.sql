-- +goose Up
-- The plan entitlements Registre last served, one forfait export document in row 1, so a
-- restart enforces the last known plans instead of failing open again.

CREATE TABLE forfait_state (
    id bigint PRIMARY KEY,
    document text NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE forfait_state;
