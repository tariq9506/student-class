-- Remove the added columns from the "schools" table
ALTER TABLE public.schools
DROP COLUMN district,
DROP COLUMN street_address,
DROP COLUMN city,
DROP COLUMN state,
DROP COLUMN zip,
DROP COLUMN created_at;
