-- Adding the 'duration' column
ALTER TABLE tutor_shifts
ADD COLUMN duration INTEGER NOT NULL;

COMMENT ON COLUMN tutor_shifts.duration IS 'Duration of the tutor shift in minutes';
