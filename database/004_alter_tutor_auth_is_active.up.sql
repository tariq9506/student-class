ALTER TABLE tutor_auth
ADD COLUMN is_active boolean DEFAULT TRUE;

COMMENT ON COLUMN tutor_auth.is_active IS 'whether the session is active or not';