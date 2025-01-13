-- Rename table back to its original name
ALTER TABLE subscription_plan RENAME TO subscription;

-- Drop the newly added columns
ALTER TABLE subscription 
DROP COLUMN name,
DROP COLUMN description,
DROP COLUMN marketing_features,
DROP COLUMN list_price,
DROP COLUMN stripe_id,
DROP COLUMN max_session;

-- Change the data type of the price column back to its original type and rename it back to amount
ALTER TABLE subscription 
ALTER COLUMN price TYPE NUMERIC USING price::NUMERIC,
RENAME COLUMN price TO amount;

-- Remove default values and constraints
ALTER TABLE subscription 
ALTER COLUMN frequency TYPE VARCHAR(255) USING frequency::VARCHAR(255),
ALTER COLUMN frequency DROP DEFAULT,
DROP CONSTRAINT IF EXISTS frequency_check;

ALTER TABLE subscription 
ALTER COLUMN recurring DROP DEFAULT;

-- Rename the column back to its original name
ALTER TABLE user2subscription
RENAME COLUMN subscription_plan_id TO subscription_id;

-- Drop the newly added columns
ALTER TABLE user2subscription 
DROP COLUMN payment_status,
DROP COLUMN invoice_pdf_link;

-- Drop the table created to store the user's customer ID
DROP TABLE IF EXISTS public.user2stripe;

-- Drop the table created to store the webhook events
DROP TABLE IF EXISTS public.stripe_webhook_event;

DROP TABLE IF EXISTS public.payment_history;
