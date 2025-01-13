CREATE TABLE session_exceptions
(
    "id" SERIAL PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "user"("id"),
    "tutor_id" BIGINT NOT NULL REFERENCES "tutor"("id"),
    "session_start" TIMESTAMP,
    "session_end" TIMESTAMP,
    "duration_mins" INTEGER,
    "status" VARCHAR(20) CHECK ("status" IN ('active', 'started', 'cancelled', 'completed')),
    "type" VARCHAR(10) CHECK ("type" IN ('demo', 'paid')),
    "timezone" VARCHAR(50),
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Comments for table and columns
COMMENT ON COLUMN session_exceptions.id IS 'Primary key for this table';
COMMENT ON COLUMN session_exceptions.user_id IS 'user_id is a reference to table public.user';
COMMENT ON COLUMN session_exceptions.tutor_id IS 'tutor_id is a reference to table tutor';
COMMENT ON COLUMN session_exceptions.session_start IS 'Start time of the session';
COMMENT ON COLUMN session_exceptions.session_end IS 'End time of the session';
COMMENT ON COLUMN session_exceptions.duration_mins IS 'Duration of the session in minutes';
COMMENT ON COLUMN session_exceptions.status IS 'Status of the session (active, started, cancelled, completed)';
COMMENT ON COLUMN session_exceptions.type IS 'Type of session (demo or paid)';
COMMENT ON COLUMN session_exceptions.timezone IS 'Timezone for the session chosen by the student';
COMMENT ON COLUMN session_exceptions.created_at IS 'Timestamp with timezone indicating the creation time of the session record';
