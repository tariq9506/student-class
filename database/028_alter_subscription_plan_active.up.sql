ALTER TABLE public.subscription_plan 
ADD COLUMN "status" VARCHAR(20) DEFAULT 'active' 
CHECK ("status" IN ('active', 'inactive'));