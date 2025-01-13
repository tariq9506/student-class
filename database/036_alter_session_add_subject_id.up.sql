-- Add two new columns 'subject_id' and 'student_id', both of type BIGINT, to the 'session' table.
-- These columns are added with a NOT NULL constraint, meaning every record must have values for these columns.
ALTER TABLE session 
ADD COLUMN subject_id BIGINT,
ADD COLUMN student_id BIGINT;
COMMENT ON COLUMN session.subject_id IS 'subject id identify subject of session';
COMMENT ON COLUMN session.student_id IS 'student id identify which session assign which student';

-- Add a foreign key constraint to the 'student_id' column, referencing the 'student_id' column in the 'student' table.
-- This ensures that any value in the 'student_id' column of the 'session' table must match a valid 'student_id' in the 'student' table.
ALTER TABLE session
ADD CONSTRAINT fk_student_id FOREIGN KEY (student_id) REFERENCES student(id),

-- Add a foreign key constraint to the 'subject_id' column, referencing the 'subject_id' column in the 'subject' table.
-- This ensures that any value in the 'subject_id' column of the 'session' table must match a valid 'subject_id' in the 'subject' table.
ADD CONSTRAINT fk_subject_id FOREIGN KEY (subject_id) REFERENCES subject(id);
