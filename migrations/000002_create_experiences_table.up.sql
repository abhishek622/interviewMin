CREATE TABLE IF NOT EXISTS experiences (
    exp_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,

    company       TEXT NOT NULL,
    position      TEXT NOT NULL,
    source        TEXT NOT NULL,
    no_of_round   INT NOT NULL,
    metadata      JSONB NOT NULL,
    source_link   TEXT NOT NULL,
    location      TEXT NOT NULL,
    -- default to an empty tsvector; trigger will fill the real tsvector
    search_tsv    TSVECTOR NOT NULL DEFAULT ''::tsvector,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- indexes
CREATE INDEX IF NOT EXISTS idx_experiences_user ON experiences(user_id);
CREATE INDEX IF NOT EXISTS idx_experiences_company ON experiences(company);
CREATE INDEX IF NOT EXISTS idx_experiences_position ON experiences(position);

-- trigger function to populate search_tsv from relevant columns
CREATE OR REPLACE FUNCTION experiences_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_tsv :=
    to_tsvector('english',
      coalesce(NEW.company,'') || ' ' ||
      coalesce(NEW.position,'') || ' ' ||
      coalesce(NEW.metadata->>'title','')
    );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER tsv_update
BEFORE INSERT OR UPDATE ON experiences
FOR EACH ROW EXECUTE FUNCTION experiences_tsv_trigger();

-- GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_experiences_tsv ON experiences USING GIN(search_tsv);

CREATE TRIGGER trigger_update_experiences
BEFORE UPDATE ON experiences
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
