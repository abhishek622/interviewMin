ALTER TABLE interviews ADD COLUMN IF NOT EXISTS input_hash TEXT;
CREATE INDEX IF NOT EXISTS idx_interviews_input_hash ON interviews(input_hash);