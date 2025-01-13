-- Add the `updated_at` and `created_at` columns to `user2stripe`
ALTER TABLE public.user2stripe
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
