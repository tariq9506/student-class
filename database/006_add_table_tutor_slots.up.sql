CREATE TABLE "tutor_shifts" (
    id SERIAL PRIMARY KEY,
    tutor_id BIGINT,
    day_of_week VARCHAR(10) CHECK ("day_of_week" IN ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday')),
    start_time TIME,
    end_time TIME,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tutor_id) REFERENCES tutor(id)
);

CREATE TABLE "tutor_unavailable_times" (
  id SERIAL PRIMARY KEY,
  tutor_id BIGINT,
  start_date  timestamp with time zone,
  end_date  timestamp with time zone,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
   FOREIGN KEY (tutor_id) REFERENCES tutor(id)

);

CREATE TABLE "tutor_available_slots" (
  id SERIAL PRIMARY KEY,
  tutor_id BIGINT,
  start_date  timestamp with time zone,
  end_date  timestamp with time zone,
  FOREIGN KEY (tutor_id) REFERENCES tutor(id)
);