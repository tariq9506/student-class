ALTER TABLE tutor_available_slots ADD CONSTRAINT idx_unique_slots UNIQUE (tutor_id,start_date);
