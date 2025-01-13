CREATE TABLE free_session
(
    "id" SERIAL PRIMARY KEY,
    "user_id" BIGINT NOT NULL REFERENCES "user"("id"),
    "type" VARCHAR(10) CHECK ("type" IN ('given', 'earned')),
    "active_subscription_required" BOOLEAN,
    "admin_id" BIGINT NOT NULL,
    "referred_user_id" text,
    "valid_until" TIMESTAMP WITH TIME ZONE,
    "status" VARCHAR(10) CHECK("status" IN ('active','expired')),
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Comments for table and columns
COMMENT ON COLUMN free_session.id IS 'Unique identifier for each free session record.';
COMMENT ON COLUMN free_session.user_id IS 'Foreign key referencing the "user" table, representing the user who owns this free session.';
COMMENT ON COLUMN free_session.type IS 'Indicates whether the free session was "given" (by an admin) or "earned" (by the user).';
COMMENT ON COLUMN free_session.active_subscription_required IS 'Indicates if an active subscription is required for booking the free session.';
COMMENT ON COLUMN free_session.admin_id IS 'Foreign key referencing the admin who assigned the free session.';
COMMENT ON COLUMN free_session.referred_user_id IS 'Stores information about how the user was referred to receive a free session.';
COMMENT ON COLUMN free_session.valid_until IS 'Expiration date for the free session, indicating its booking validity.';
COMMENT ON COLUMN free_session.status IS 'Current status of the free session (either "active" or "expired").';
COMMENT ON COLUMN free_session.created_at IS 'Timestamp of when the free session record was created.';
COMMENT ON COLUMN free_session.updated_at IS 'Timestamp of the most recent update to the free session record.';


 
