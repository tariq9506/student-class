CREATE TABLE courses (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    admin_id BIGINT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE modules (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    admin_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    description TEXT,
    sort_order int,
    updated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE SET NULL

);
CREATE TABLE lessons (
    id BIGSERIAL PRIMARY KEY, -- Unique identifier for the lesson
    course_id BIGINT NULL, -- Foreign key to Courses (nullable)
    module_id BIGINT NULL, -- Foreign key to Modules (nullable)
    admin_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL, -- Name of the lesson
    description TEXT, -- Optional detailed description of the lesson\
    lesson_pdf TEXT, 
    duration int, -- lesson duration in minutes
    sort_order int,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Last update timestamp
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE SET NULL, -- Foreign key to Courses
    CONSTRAINT fk_module FOREIGN KEY (module_id) REFERENCES modules (id) ON DELETE SET NULL -- Foreign key to Modules 
);
CREATE TABLE subject2course (
    id BIGSERIAL PRIMARY KEY, -- Unique identifier
    subject_id BIGINT NOT NULL, -- Foreign key to Modules
    course_id BIGINT NOT NULL, -- Foreign key to Grades
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp
    CONSTRAINT fk_module FOREIGN KEY (subject_id) REFERENCES subject (id) ON DELETE CASCADE,
    CONSTRAINT fk_grade FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);
ALTER TABLE subject2course ADD CONSTRAINT unique_subject_course UNIQUE (subject_id,course_id);