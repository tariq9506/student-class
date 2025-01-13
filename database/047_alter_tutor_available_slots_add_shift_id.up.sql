-- Add the `shift_id` column to `tutor_available_slots` and create a foreign key reference to `tutor_shifts(id)`
ALTER TABLE tutor_available_slots
ADD COLUMN shift_id INTEGER,
ADD CONSTRAINT fk_shift_id
    FOREIGN KEY (shift_id)
    REFERENCES tutor_shifts(id);
