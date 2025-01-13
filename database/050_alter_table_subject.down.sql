-- Remove `image` and `reschedule_allowed_days` columns from `subject` table
ALTER TABLE subject
DROP COLUMN image,
DROP COLUMN reschedule_allowed_days;

-- Remove unique constraint on the `name` column
ALTER TABLE subject
DROP CONSTRAINT unique_subject;
