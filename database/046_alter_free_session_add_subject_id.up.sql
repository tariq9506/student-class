-- Add two new columns 'subject_id' type BIGINT, to the 'free_session' table.
-- this column is added meaning every record must have values for subject id column.
ALTER TABLE free_session
ADD COLUMN subject_id BIGINT;
COMMENT ON COLUMN free_session.subject_id IS 'subject id identify subject of session';