-- Add a new column 'grade_id' of type INT to the 'student' table with a NOT NULL constraint.
ALTER TABLE student
ADD COLUMN grade_id BIGINT;
COMMENT ON COLUMN student.grade_id IS 'grade id identify the grade of students';
-- Add a foreign key constraint on 'grade_id' in the 'student' table, ensuring it references the 'grade_id' in the 'grade' table.
ALTER TABLE student
ADD CONSTRAINT fk_grade_id FOREIGN KEY (grade_id) REFERENCES grade(id);
