-- Drop the foreign key constraint 'fk_grade_id' from the 'student' table.
ALTER TABLE student 
DROP CONSTRAINT fk_grade_id;
-- Drop the 'grade_id' column from the 'student' table. This permanently removes the column and all its associated data.
ALTER TABLE student 
DROP COLUMN grade_id;


