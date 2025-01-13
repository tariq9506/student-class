-- Create the subject2grades table
CREATE TABLE subject2grades (
    id BIGSERIAL PRIMARY KEY, -- Unique identifier
    subject_id BIGINT NOT NULL, -- Foreign key to subject table
    grade_id BIGINT NOT NULL, -- Foreign key to grade table
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp
    CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE CASCADE,
    CONSTRAINT fk_grade FOREIGN KEY (grade_id) REFERENCES grade (id) ON DELETE CASCADE,
    CONSTRAINT unique_subject_grade UNIQUE (subject_id, grade_id) -- Unique constraint
);
