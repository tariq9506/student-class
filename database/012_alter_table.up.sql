
--Renamed the table name
ALTER TABLE subscription RENAME TO subscription_plan;

-- Add new columns
ALTER TABLE subscription_plan 
DROP COLUMN stripe_id,
DROP COLUMN max_session,
ADD COLUMN name TEXT NOT NULL,
ADD COLUMN description TEXT,
ADD COLUMN marketing_features TEXT,
ADD COLUMN list_price INTEGER,
ADD COLUMN stripe_id VARCHAR(255),
ADD COLUMN max_session INTEGER NOT NULL;

-- Change the data type of the amount column to INTEGER and rename it to price
ALTER TABLE subscription_plan 
ALTER COLUMN amount TYPE INTEGER USING amount::INTEGER;

ALTER TABLE subscription_plan 
RENAME COLUMN amount TO price;


-- Add default values and constraints where applicable
ALTER TABLE subscription_plan 
ALTER COLUMN frequency TYPE VARCHAR(9) USING frequency::VARCHAR(9),
ALTER COLUMN frequency SET DEFAULT 'month',
ADD CONSTRAINT frequency_check CHECK (frequency IN ('month', 'quarter', 'year'));

ALTER TABLE subscription_plan 
ALTER COLUMN recurring SET DEFAULT TRUE;

ALTER TABLE subscription_plan
ADD COLUMN per_week_limit INTEGER;



ALTER TABLE user2subscription
RENAME COLUMN subscription_id TO subscription_plan_id;

ALTER TABLE user2subscription
ADD COLUMN status VARCHAR(20) 
CHECK (status IN ('inactive', 'active', 'cancelled', 'past_due'));

ALTER TABLE user2subscription
DROP COLUMN stripe_txn_id,
ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;


--Create table "user2stripe" to store the user's customer ID. 
CREATE TABLE public.user2stripe (
  id SERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES "user"(id),
  stripe_id VARCHAR(255)
);


--Created table to store the webhook events.
CREATE TABLE public.stripe_webhook_event (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES "user"(id),
  event_id VARCHAR(255),
  event_type VARCHAR(255),
  data JSONB,
  received_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE payment_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES "user"(id),
    purchased_date TIMESTAMP WITH TIME ZONE,
    invoice_pdf_link VARCHAR(400),
    subscription_plan_id INTEGER REFERENCES "subscription_plan"(id),
    amount VARCHAR(60),
    stripe_intent_id VARCHAR(300),
    status VARCHAR(20) CHECK (status IN ('succeeded', 'failed', 'pending')),
    currency CHARACTER VARYING(3),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
