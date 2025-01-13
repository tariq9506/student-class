ALTER TABLE session 
    DROP CONSTRAINT session_status_check;
ALTER TABLE session 
    DROP COLUMN status;
