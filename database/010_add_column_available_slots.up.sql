ALTER TABLE tutor_available_slots ADD COLUMN  day_of_week VARCHAR(10) CHECK ("day_of_week" IN ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'));
