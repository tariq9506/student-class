ALTER TABLE free_session 
DROP COLUMN message_status;

ALTER TABLE free_session
DROP CONSTRAINT free_session_status_check;s