-- Add unique constraint on (tutor_id, start_date, end_date)
ALTER TABLE tutor_unavailable_times
ADD CONSTRAINT unique_tutor_unavailable_times 
UNIQUE (tutor_id, start_date, end_date);
