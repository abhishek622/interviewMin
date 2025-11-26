DROP INDEX IF EXISTS idx_experiences_tsv;
DROP TRIGGER IF EXISTS tsv_update;
DROP FUNCTION IF EXISTS experiences_tsv_trigger;
DROP INDEX IF EXISTS idx_experiences_company;
DROP INDEX IF EXISTS idx_experiences_position;
DROP INDEX IF EXISTS idx_experiences_user;
DROP TABLE IF EXISTS experiences;