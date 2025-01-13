ALTER TABLE session
ADD CONSTRAINT unique_tutor_session_time UNIQUE (tutor_id,session_start);