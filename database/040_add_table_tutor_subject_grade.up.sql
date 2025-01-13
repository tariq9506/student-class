CREATE TABLE public.tutor_subject_grade (
    id BIGSERIAL PRIMARY KEY,
    tutor_id BIGINT NOT NULL,
    grade_id BIGINT NOT NULL,
    subject_id BIGINT NOT NULL,
    updated_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT tutor_subject_grade_unique UNIQUE(tutor_id, subject_id, grade_id),
    CONSTRAINT fk_tutor
        FOREIGN KEY(tutor_id) 
        REFERENCES tutor(id),
    CONSTRAINT fk_grade
        FOREIGN KEY(grade_id) 
        REFERENCES grade(id),
    CONSTRAINT fk_subject
        FOREIGN KEY(subject_id) 
        REFERENCES subject(id)
);