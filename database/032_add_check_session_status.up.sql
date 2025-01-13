ALTER TABLE session
DROP CONSTRAINT session_status_check;

ALTER TABLE session
ADD CONSTRAINT session_status_check
CHECK(status::text=ANY(ARRAY['active','started','cancelled','completed','absent','noshow']::text[]));