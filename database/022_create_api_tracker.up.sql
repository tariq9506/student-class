CREATE TABLE api_tracker(
    id serial,
    full_url character varying(1000),
    ip character varying(20),
    query_param character varying(1000),
    created_at timestamp with time zone,
    user_session character varying(1000),
    body_param TEXT,
    user_agent TEXT,
    api_method varchar CHECK (api_method IN ('GET','POST','PUT','DELETE','PATCH'))
);

-- Add comment on the column
COMMENT ON COLUMN api_tracker.id IS 'Primary key for this table';
COMMENT ON COLUMN api_tracker.full_url IS 'full url of the api request';
COMMENT ON COLUMN api_tracker.ip IS 'ip address of the client making request';
COMMENT ON COLUMN api_tracker.query_param IS 'query_param is list of all query parameters used in an api request';
COMMENT ON COLUMN api_tracker.created_at IS 'time when api request is made';
COMMENT ON COLUMN api_tracker.user_session IS 'token of the user making api calls';
COMMENT ON COLUMN api_tracker.body_param IS 'body_param is the list of parameters that are used in body of an api request';
COMMENT ON COLUMN api_tracker.user_agent IS 'user_agent is the agent which is calling the api';
COMMENT ON COLUMN api_tracker.api_method IS 'method of api request like GET,POST,DELETE,PUT,PATCH';