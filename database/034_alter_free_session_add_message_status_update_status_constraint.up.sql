ALTER TABLE free_session
ADD COLUMN "message_status" VARCHAR(10) 
CHECK ("message_status" IN ('draft','sent'));


ALTER TABLE free_session
DROP CONSTRAINT free_session_status_check;


ALTER TABLE free_session
ADD CONSTRAINT free_session_status_check
CHECK (status:: text = ANY (ARRAY['active','expired','used']::text[]));