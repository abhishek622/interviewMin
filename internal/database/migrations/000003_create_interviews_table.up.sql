CREATE TABLE IF NOT EXISTS interviews (
    interview_id    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    company_id      UUID NOT NULL REFERENCES companies(company_id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(user_id),
    source          VARCHAR(50) NOT NULL,   
    raw_input       TEXT NOT NULL,                  -- original input

    process_status  VARCHAR(20) NOT NULL DEFAULT 'queued', -- queued | processing | completed | failed
    process_error   TEXT,
    attempts        INT NOT NULL DEFAULT 0,

    position        VARCHAR(255),
    no_of_round     INT,
    metadata        JSONB DEFAULT NULL, -- ai output
    location        VARCHAR(255),
    search_tsv      TSVECTOR NOT NULL DEFAULT ''::tsvector,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- indexes
CREATE INDEX IF NOT EXISTS idx_interviews_company ON interviews(company_id);
CREATE INDEX IF NOT EXISTS idx_interviews_user ON interviews(user_id);

-- trigger function to populate search_tsv from relevant columns
CREATE OR REPLACE FUNCTION interviews_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_tsv :=
    to_tsvector('english',
      coalesce(NEW.position,'') || ' ' ||
      coalesce(NEW.metadata->>'title','')
    );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER tsv_update
BEFORE INSERT OR UPDATE ON interviews
FOR EACH ROW EXECUTE FUNCTION interviews_tsv_trigger();

-- GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_interviews_tsv ON interviews USING GIN(search_tsv);

CREATE TRIGGER trigger_update_interviews
BEFORE UPDATE ON interviews
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
