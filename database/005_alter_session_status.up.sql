ALTER TABLE session
    DROP CONSTRAINT session_status_check,
    ADD CONSTRAINT session_status_check CHECK (status IN ('active', 'started', 'cancelled', 'completed'));
