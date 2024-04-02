CREATE INDEX IF NOT EXISTS strategies_names_idx ON strategies USING GIN (to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS strategies_fields_ ON strategies USING GIN (fields);