-- Add unique_tutor_shift constraint
ALTER TABLE tutor_shifts
ADD CONSTRAINT unique_tutor_shift
UNIQUE (tutor_id, day_of_week, start_time);
