-- Drop the 'subject_id' and 'student_id' columns from the 'session_exceptions' table.
ALTER TABLE session_exceptions 
DROP COLUMN subject_id,
DROP COLUMN student_id;