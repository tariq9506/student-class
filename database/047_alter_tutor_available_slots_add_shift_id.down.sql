-- Drop the foreign key constraint if it exists
ALTER TABLE tutor_available_slots
DROP CONSTRAINT IF EXISTS fk_shift_id;

-- Remove the `shift_id` column from `tutor_available_slots`
ALTER TABLE tutor_available_slots
DROP COLUMN IF EXISTS shift_id;
