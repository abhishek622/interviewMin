DROP INDEX IF EXISTS idx_interviews_slug_unique;

ALTER TABLE interviews DROP COLUMN slug;

ALTER TABLE interviews ALTER COLUMN company DROP DEFAULT;
ALTER TABLE interviews ALTER COLUMN company DROP NOT NULL;
