-- Drop the existing constraint
ALTER TABLE public.student
DROP CONSTRAINT IF EXISTS student_user_id_key;

-- Add the new composite unique constraint
ALTER TABLE public.student
ADD CONSTRAINT student_user_id_name_key UNIQUE (user_id, name);

ALTER TABLE public.student_session_preference
ADD COLUMN student_id BIGINT,
ADD COLUMN subject_id BIGINT,
ADD CONSTRAINT fk_student FOREIGN KEY (student_id) REFERENCES public.student(id),
ADD CONSTRAINT fk_subject FOREIGN KEY (subject_id) REFERENCES public.subject(id);

