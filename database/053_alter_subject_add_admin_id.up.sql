-- Add `admin_id` column to `subject` table
ALTER TABLE subject
ADD COLUMN admin_id BIGINT;

-- Update the subject with super admin
UPDATE subject
SET admin_id = 1;

-- Add the NOT NULL constraint to the column
ALTER TABLE subject
ALTER COLUMN admin_id SET NOT NULL;