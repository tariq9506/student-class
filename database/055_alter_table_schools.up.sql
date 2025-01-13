-- Add the additional columns to the "schools" table
ALTER TABLE public.schools
ADD COLUMN district VARCHAR(255),
ADD COLUMN street_address VARCHAR(355),
ADD COLUMN city VARCHAR(150),
ADD COLUMN state CHAR(6),
ADD COLUMN zip VARCHAR(15),
ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
