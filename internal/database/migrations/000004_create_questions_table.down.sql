DROP INDEX IF EXISTS idx_questions_interview;
DROP INDEX IF EXISTS idx_questions_type;
DROP TRIGGER IF EXISTS trigger_update_questions ON questions;
DROP TABLE IF EXISTS questions;