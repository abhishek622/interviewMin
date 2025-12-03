ALTER TABLE experiences ADD COLUMN source VARCHAR(50);
ALTER TABLE experiences DROP COLUMN IF EXISTS input_type;
ALTER TABLE experiences DROP COLUMN IF EXISTS extracted_title;
ALTER TABLE experiences DROP COLUMN IF EXISTS extracted_content;