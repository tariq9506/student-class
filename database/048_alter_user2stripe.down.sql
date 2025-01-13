-- Remove the `updated_at` and `created_at` columns from `user2stripe`
ALTER TABLE public.user2stripe
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS updated_at;
