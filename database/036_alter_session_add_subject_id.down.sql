-- Drop the foreign key constraint on 'student_id'.
ALTER TABLE session 
DROP CONSTRAINT fk_student_id;

-- Drop the foreign key constraint on 'subject_id'.
ALTER TABLE session 
DROP CONSTRAINT fk_subject_id;

-- Drop the 'subject_id' and 'student_id' columns from the 'session' table.
ALTER TABLE session 
DROP COLUMN subject_id,
DROP COLUMN student_id;
