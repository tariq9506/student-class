ALTER TABLE session
ADD COLUMN session_confirmed BOOLEAN DEFAULT FALSE;

COMMENT ON COLUMN session.session_confirmed IS 'session_confirmed is whether session is confirmed or not';