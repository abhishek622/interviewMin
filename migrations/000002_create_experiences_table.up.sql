CREATE TABLE IF NOT EXISTS experiences (
    exp_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,

    company       TEXT NOT NULL,                    -- e.g., Google, Meta
    position      TEXT NOT NULL,                    -- e.g., SDE, Backend Engineer
    source        TEXT NOT NULL,                    -- e.g., Leetcode, Reddit, GFG
    no_of_round   INT NOT NULL,                     -- e.g., 1
    metadata      JSONB NOT NULL,                   -- full pasted experience
    source_link   TEXT NOT NULL,
    location      TEXT NOT NULL,
    search_tsv    TSVECTOR NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_experiences_user ON experiences(user_id);
CREATE INDEX idx_experiences_company ON experiences(company);
CREATE INDEX idx_experiences_position ON experiences(position);

CREATE FUNCTION experiences_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_tsv :=
    to_tsvector('english',
      coalesce(NEW.company,'') || ' ' ||
      coalesce(NEW.position,'') || ' ' ||
      coalesce(NEW.metadata->>'title','') || ' ' ||
          );
  RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER tsv_update BEFORE INSERT OR UPDATE
ON experiences FOR EACH ROW EXECUTE FUNCTION experiences_tsv_trigger();

CREATE INDEX idx_experiences_tsv ON experiences USING GIN(search_tsv);

CREATE OR REPLACE TRIGGER trigger_update_experiences
    BEFORE UPDATE ON experiences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
