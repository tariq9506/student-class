-- Add two new columns 'subject_id' and 'student_id', both of type BIGINT, to the 'session_exceptions' table.
-- These columns are added meaning every record must have values for these columns.
ALTER TABLE session_exceptions
ADD COLUMN subject_id BIGINT,
ADD COLUMN student_id BIGINT;
COMMENT ON COLUMN session_exceptions.subject_id IS 'subject id identify subject of session';
COMMENT ON COLUMN session_exceptions.student_id IS 'student id identify which session assign which student';