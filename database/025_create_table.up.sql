
CREATE TABLE "student_session_preference" (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    tutor_id BIGINT,
    day_of_week VARCHAR(10) CHECK ("day_of_week" IN ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday')),
    start_time TIME,
    timezone varchar(50),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    FOREIGN KEY (tutor_id) REFERENCES tutor(id),
    FOREIGN KEY (user_id) REFERENCES public.user(id)
);
