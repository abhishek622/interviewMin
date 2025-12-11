ALTER TABLE interviews ADD COLUMN slug VARCHAR(255);
UPDATE interviews SET company = 'unknown company' WHERE company IS NULL OR company = '';
UPDATE interviews SET slug = regexp_replace(lower(company), '[^a-z0-9]+', '-', 'g');
ALTER TABLE interviews ALTER COLUMN company SET DEFAULT 'unknown company';
ALTER TABLE interviews ALTER COLUMN slug SET DEFAULT 'unknown-company';
ALTER TABLE interviews ALTER COLUMN slug SET NOT NULL;
ALTER TABLE interviews ALTER COLUMN company SET NOT NULL;

CREATE UNIQUE INDEX idx_interviews_slug_unique ON interviews(slug);
