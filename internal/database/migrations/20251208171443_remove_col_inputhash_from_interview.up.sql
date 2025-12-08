DROP INDEX IF EXISTS idx_interviews_input_hash;
ALTER TABLE interviews DROP COLUMN IF EXISTS input_hash;