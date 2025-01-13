CREATE TABLE subscription_prices (
    id BIGSERIAL PRIMARY KEY,
    stripe_product_id VARCHAR(255) NOT NULL,
    subscription_plan_id BIGINT NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    price FLOAT,
    list_price FLOAT,
    default_price BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);