CREATE TABLE "user2zoho_leads" (
    id SERIAL PRIMARY KEY,
    user_id BIGINT,
    lead_id BIGINT,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    FOREIGN KEY (user_id) REFERENCES public.user(id)
);
COMMENT ON COLUMN user2zoho_leads.user_id IS 'User id referenced from user table.';
COMMENT ON COLUMN user2zoho_leads.lead_id IS 'Store lead id which i created on zoho CRM.';
COMMENT ON COLUMN user2zoho_leads.created_at IS 'Created date when record create of user.';
COMMENT ON COLUMN user2zoho_leads.updated_at IS 'Updated date when record update of user.';