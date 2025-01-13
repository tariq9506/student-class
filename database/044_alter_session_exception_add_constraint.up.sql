-- Adds a unique constraint to the session_exceptions table on the combination of tutor_id and session_start.
-- This ensures that each tutor cannot have multiple sessions starting at the same time, preventing duplicate entries
-- based on the tutor's ID and session start time.
ALTER TABLE session_exceptions
ADD CONSTRAINT unique_session_time_tutor_id UNIQUE(tutor_id, session_start);
