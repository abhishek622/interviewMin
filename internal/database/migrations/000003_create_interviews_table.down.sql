DROP INDEX IF EXISTS idx_interviews_tsv;
DROP INDEX IF EXISTS idx_interviews_company;
DROP INDEX IF EXISTS idx_interviews_user;
DROP TRIGGER IF EXISTS tsv_update ON interviews;
DROP TRIGGER IF EXISTS trigger_update_interviews ON interviews;
DROP FUNCTION IF EXISTS interviews_tsv_trigger();
DROP TABLE IF EXISTS interviews;
