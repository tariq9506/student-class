ALTER TABLE payment_history
ADD COLUMN card_last4 VARCHAR(5),
ADD COLUMN payment_type VARCHAR(10),
ADD COLUMN brand VARCHAR(20),
ADD COLUMN exp_year VARCHAR(5),
ADD COLUMN exp_month VARCHAR(2),
ADD COLUMN billing_address TEXT;