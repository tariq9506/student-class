ALTER TABLE public.user2subscription 
ADD CONSTRAINT user_subscription_unique UNIQUE(user_id, subscription_plan_id);
