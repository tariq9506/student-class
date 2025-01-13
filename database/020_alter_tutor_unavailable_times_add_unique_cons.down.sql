-- Remove unique constraint on (tutor_id, start_date, end_date)
ALTER TABLE tutor_unavailable_times
DROP CONSTRAINT IF EXISTS unique_tutor_unavailable_times;
