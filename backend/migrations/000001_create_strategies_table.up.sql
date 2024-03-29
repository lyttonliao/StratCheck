CREATE TABLE IF NOT EXISTS strategies (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL UNIQUE,
    fields text[] NOT NULL,
    criteria text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);