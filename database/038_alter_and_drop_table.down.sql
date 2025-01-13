ALTER TABLE public.student
DROP CONSTRAINT IF EXISTS student_user_id_name_key;

-- Drop the existing constraint
ALTER TABLE public.student
ADD CONSTRAINT student_user_id_key UNIQUE(user_id);

ALTER TABLE public.student_session_preference
DROP CONSTRAINT fk_student,
DROP CONSTRAINT fk_subject,
DROP COLUMN student_id,
DROP COLUMN subject_id;

