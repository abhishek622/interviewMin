DROP INDEX IF EXISTS idx_experiences_tsv;
DROP INDEX IF EXISTS idx_experiences_company;
DROP INDEX IF EXISTS idx_experiences_position;
DROP INDEX IF EXISTS idx_experiences_input_hash;
DROP INDEX IF EXISTS idx_experiences_user;
DROP TRIGGER IF EXISTS tsv_update ON experiences;
DROP TRIGGER IF EXISTS trigger_update_experiences ON experiences;
DROP FUNCTION IF EXISTS experiences_tsv_trigger();
DROP TABLE IF EXISTS experiences;
