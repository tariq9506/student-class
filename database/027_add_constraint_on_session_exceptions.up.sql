ALTER TABLE session_exceptions
DROP CONSTRAINT session_exceptions_status_check;

ALTER TABLE session_exceptions
ADD CONSTRAINT session_exceptions_status_check 
CHECK (status::text = ANY (ARRAY['active', 'started', 'cancelled', 'completed', 'absent']::text[]));