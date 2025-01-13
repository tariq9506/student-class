-- Add `image` and `reschedule_allowed_days` columns to `subject` table
ALTER TABLE subject
ADD COLUMN image VARCHAR(200),
ADD COLUMN reschedule_allowed_days INT;

-- Add unique constraint on the `name` column
ALTER TABLE subject
ADD CONSTRAINT unique_subject UNIQUE (name);
