ALTER TABLE tutor
DROP COLUMN timezone,
DROP COLUMN teaches_to,
DROP COLUMN zoom_link;

ALTER TABLE session
DROP game_url;

ALTER TABLE session
    DROP CONSTRAINT session_status_check,
    ADD CONSTRAINT session_status_check CHECK (status IN ('active', 'started', 'cancelled', 'completed'));