CREATE TABLE IF NOT EXISTS experiences (
    exp_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,

    -- user-provided input
    input_type      TEXT NOT NULL DEFAULT 'url',    -- url | text
    raw_input       TEXT NOT NULL,                  -- original input
    input_hash      TEXT NOT NULL,                  -- sha256(raw_input)

    -- async processing
    process_status  TEXT NOT NULL DEFAULT 'queued', -- queued | processing | completed | failed
    process_error   TEXT,
    attempts        INT NOT NULL DEFAULT 0,

    -- extracted before AI
    extracted_title    TEXT,
    extracted_content  TEXT,

    -- final user-facing fields
    company        TEXT,
    position       TEXT,
    source         TEXT NOT NULL,
    no_of_round    INT,
    metadata       JSONB DEFAULT NULL, -- ai output
    location       TEXT,

    search_tsv     TSVECTOR NOT NULL DEFAULT ''::tsvector,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- indexes
CREATE INDEX IF NOT EXISTS idx_experiences_user ON experiences(user_id);
CREATE INDEX IF NOT EXISTS idx_experiences_company ON experiences(company);
CREATE INDEX IF NOT EXISTS idx_experiences_position ON experiences(position);
CREATE INDEX IF NOT EXISTS idx_experiences_input_hash ON experiences(input_hash);

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
