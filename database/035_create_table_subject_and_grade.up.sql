-- Creating the 'subject' table to store details about subjects
CREATE TABLE subject (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Adding comments to columns of the 'subject' table
COMMENT ON COLUMN subject.id IS 'Unique identifier for each subject (auto-incrementing)';
COMMENT ON COLUMN subject.name IS 'Name of the subject (e.g., Mathematics, Science)';
COMMENT ON COLUMN subject.description IS 'Optional description of the subject';
COMMENT ON COLUMN subject.updated_at IS 'Timestamp for the last update of the subject record';
COMMENT ON COLUMN subject.created_at IS 'Timestamp for when the subject was created, defaults to the current time';

-- Creating the 'grade' table to store details about grade levels
CREATE TABLE grade (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    sort_level INT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Adding comments to columns of the 'grade' table
COMMENT ON COLUMN grade.id IS 'Unique identifier for each grade (auto-incrementing)';
COMMENT ON COLUMN grade.name IS 'Name of the grade (e.g., Grade 1, Grade 2)';
COMMENT ON COLUMN grade.sort_level IS 'Level used for sorting grades (numeric value)';
COMMENT ON COLUMN grade.description IS 'Optional description of the grade';
COMMENT ON COLUMN grade.updated_at IS 'Timestamp for the last update of the grade record';
COMMENT ON COLUMN grade.created_at IS 'Timestamp for when the grade was created, defaults to the current time';

-- Insert default subject
INSERT INTO subject (name, description)
VALUES 
    ('Coding', 'Study of programming and software development.');

-- Insert default grades
INSERT INTO grade (name, sort_level, description)
VALUES 
    ('Grade 1', 1, 'First grade'),
    ('Grade 2', 2, 'Second grade'),
    ('Grade 3', 3, 'Third grade'),
    ('Grade 4', 4, 'Fourth grade'),
    ('Grade 5', 5, 'Fifth grade'),
    ('Grade 6', 6, 'Sixth grade'),
    ('Grade 7', 7, 'Seventh grade'),
    ('Grade 8', 8, 'Eighth grade'),
    ('Grade 9', 9, 'Ninth grade'),
    ('Grade 10', 10, 'Tenth grade');
