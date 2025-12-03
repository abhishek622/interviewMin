ALTER TABLE experiences DROP COLUMN source;
ALTER TABLE experiences ADD COLUMN input_type VARCHAR(10) NOT NULL DEFAULT 'url';
ALTER TABLE experiences ADD COLUMN extracted_title VARCHAR(255);
ALTER TABLE experiences ADD COLUMN extracted_content TEXT;