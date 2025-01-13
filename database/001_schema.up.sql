CREATE TABLE "user" (
  "id" SERIAL PRIMARY KEY,
  "email" VARCHAR(100),
  "phone_number" VARCHAR(15) UNIQUE,
  "otp" VARCHAR(4),
  "otp_valid_until" timestamp with time zone,
  "ip" INET,
  "location" TEXT,
  "phone_verified" BOOLEAN DEFAULT FALSE,
  "email_verified" BOOLEAN DEFAULT FALSE,
  "referral_code" VARCHAR(20),
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  "timezone" VARCHAR(3)
);

CREATE TABLE "student" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT UNIQUE,
  "school_id" INTEGER,
  "picture_url" TEXT,
  "name" VARCHAR(100),
  "grade" VARCHAR(50),
  "parent_name" VARCHAR(100),
  "parent_email" VARCHAR(100)
);

CREATE TABLE "ambassador" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT UNIQUE,
  "name" VARCHAR(100)
);

CREATE TABLE "roles" (
  "id" SERIAL PRIMARY KEY,
  "role" VARCHAR(20) UNIQUE CHECK ("role" IN ('student', 'ambassador'))
);
INSERT INTO roles(role)
VALUES('student'),
      ('ambassador');

CREATE TABLE "user2roles" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT,
  "role_id" BIGINT
);

CREATE TABLE "user_invitee" (
  "id" SERIAL PRIMARY KEY,
  "invitee" BIGINT NOT NULL,
  "inviter" BIGINT NOT NULL,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  "amount_receivable" FLOAT
);

CREATE TABLE "user_auth" (
  "id" SERIAL PRIMARY KEY,
  "jwt_token" TEXT,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  "valid_until" timestamp with time zone,
  "user_id" BIGINT NOT NULL REFERENCES "user"("id"),
  "device_info" TEXT,
  "browser" TEXT,
  "ip" INET,
  "location" TEXT,
  "is_active" BOOLEAN DEFAULT TRUE
  );

CREATE TABLE "tutor" (
  "id" SERIAL PRIMARY KEY,
  "email_id" VARCHAR(100) UNIQUE,
  "phone_number" VARCHAR(10),
  "password" TEXT,
  "name" VARCHAR(100),
  "picture_url" TEXT,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "tutor_auth" (
  "id" SERIAL PRIMARY KEY,
  "jwt_token" TEXT,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  "valid_until" timestamp with time zone,
  "tutor_id" BIGINT NOT NULL REFERENCES "tutor"("id"),
  "device_info" TEXT,
  "browser" TEXT,
  "ip" INET,
  "location" TEXT
);

CREATE TABLE "schools" (
  "id" SERIAL PRIMARY KEY,
  "name" VARCHAR(250)
);

CREATE TABLE "session" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT NOT NULL REFERENCES "user"("id"),
  "tutor_id" BIGINT NOT NULL REFERENCES "tutor"("id"),
  "session_start" TIMESTAMP,
  "session_end" TIMESTAMP,
  "duration_mins" INTEGER,
  "status" VARCHAR(20) CHECK ("status" IN ('active', 'started', 'cancelled')),
  "type" VARCHAR(10) CHECK ("type" IN ('demo', 'paid')),
  "recording_url" TEXT,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone
);


CREATE TABLE "session_feedback" (
  "id" SERIAL PRIMARY KEY,
  "session_id" BIGINT NOT NULL REFERENCES "session"("id"),
  "rating" FLOAT,
  "comment" TEXT,
  "session_liked" BOOLEAN,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "subscription" (
  "id" SERIAL PRIMARY KEY,
  "about" TEXT,
  "max_session" INTEGER,
  "frequency" VARCHAR(10) CHECK ("frequency" IN ('monthly', 'quarterly', 'yearly')),
  "amount" FLOAT,
  "currency" VARCHAR(3),
  "recurring" BOOLEAN,
  "stripe_id" TEXT,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE "user2subscription" (
  "id" SERIAL PRIMARY KEY,
  "user_id" BIGINT NOT NULL REFERENCES "user"("id"),
  "subscription_id" BIGINT NOT NULL REFERENCES "subscription"("id"),
  "purchased_date" timestamp with time zone,
  "valid_until" timestamp with time zone,
  "stripe_txn_id" TEXT
);

CREATE EXTENSION postgis;

CREATE TABLE cities (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    state_id integer NOT NULL,
    state_code character varying(255) NOT NULL,
    country_id integer NOT NULL,
    country_code character(2) NOT NULL,
    latitude numeric(10,8) NOT NULL,
    longitude numeric(11,8) NOT NULL,
    created_at timestamp with time zone DEFAULT '2014-01-01 01:01:01+00'::timestamp with time zone NOT NULL,
    updated_on timestamp with time zone,
    flag boolean DEFAULT true NOT NULL,
    wikidataid character varying(255),
    doc_vectors tsvector
);

CREATE TABLE countries (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    iso3 character(3),
    iso2 character(2),
    phonecode character varying(255),
    capital character varying(255),
    currency character varying(255),
    native character varying(255),
    emoji character varying(191),
    emojiu character varying(191),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    flag boolean DEFAULT true NOT NULL,
    wikidataid character varying(255)
);

CREATE TABLE states (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    country_id integer NOT NULL,
    country_code character(2) NOT NULL,
    fips_code character varying(255),
    iso2 character varying(255),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    flag boolean DEFAULT true NOT NULL,
    wikidataid character varying(255),
    latitude double precision,
    longitude double precision
);



CREATE TABLE zipcode (
    id bigint NOT NULL,
    zip character varying(255),
    state character(3),
    reference_number character varying(255),
    city_name character varying(255),
    name character varying(255),
    population bigint,
    status bigint,
    latitude double precision,
    longitude double precision,
    country character varying(3),
    istatus bigint,
    city_id bigint,
    doc_vectors tsvector,
    job_count bigint,
    geom public.geometry(Point,4326),
    cities_id bigint
);



ALTER TABLE cities OWNER TO postgres;

COMMENT ON COLUMN cities.wikidataid IS 'Rapid API GeoDB Cities';

CREATE SEQUENCE cities_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE cities_id_seq OWNER TO postgres;

ALTER SEQUENCE cities_id_seq OWNED BY cities.id;

ALTER TABLE ONLY cities ALTER COLUMN id SET DEFAULT nextval('cities_id_seq'::regclass);

ALTER TABLE ONLY cities
    ADD CONSTRAINT idx_983214_primary PRIMARY KEY (id);

ALTER TABLE ONLY cities
    ADD CONSTRAINT unique_city UNIQUE (name, state_id, country_id);

CREATE INDEX idx_983214_cities_test_ibfk_1 ON cities USING btree (state_id);

CREATE INDEX idx_983214_cities_test_ibfk_2 ON cities USING btree (country_id);

CREATE INDEX idx_cities_state_code ON cities USING btree (state_code);


ALTER TABLE countries OWNER TO postgres;

COMMENT ON COLUMN countries.wikidataid IS 'Rapid API GeoDB Cities';

CREATE SEQUENCE countries_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE countries_id_seq OWNER TO postgres;

ALTER SEQUENCE countries_id_seq OWNED BY countries.id;

ALTER TABLE ONLY countries ALTER COLUMN id SET DEFAULT nextval('countries_id_seq'::regclass);

ALTER TABLE ONLY countries
    ADD CONSTRAINT idx_983225_primary PRIMARY KEY (id);

ALTER TABLE ONLY cities
    ADD CONSTRAINT cities_ibfk_2 FOREIGN KEY (country_id) REFERENCES countries(id) ON UPDATE RESTRICT ON DELETE RESTRICT;




ALTER TABLE states OWNER TO postgres;

COMMENT ON COLUMN states.wikidataid IS 'Rapid API GeoDB Cities';

CREATE SEQUENCE states_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE states_id_seq OWNER TO postgres;

ALTER SEQUENCE states_id_seq OWNED BY states.id;

ALTER TABLE ONLY states ALTER COLUMN id SET DEFAULT nextval('states_id_seq'::regclass);

ALTER TABLE ONLY states
    ADD CONSTRAINT idx_983235_primary PRIMARY KEY (id);

CREATE INDEX idx_983235_country_region ON states USING btree (country_id);

ALTER TABLE ONLY cities
    ADD CONSTRAINT cities_ibfk_1 FOREIGN KEY (state_id) REFERENCES states(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


ALTER TABLE ONLY states
    ADD CONSTRAINT country_region_final FOREIGN KEY (country_id) REFERENCES countries(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


ALTER TABLE zipcode OWNER TO postgres;

CREATE SEQUENCE zipcode_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE zipcode_id_seq OWNER TO postgres;

ALTER SEQUENCE zipcode_id_seq OWNED BY zipcode.id;

ALTER TABLE ONLY zipcode ALTER COLUMN id SET DEFAULT nextval('zipcode_id_seq'::regclass);

ALTER TABLE ONLY zipcode
    ADD CONSTRAINT idx_63189_primary PRIMARY KEY (id);

ALTER TABLE ONLY zipcode
    ADD CONSTRAINT unique_name UNIQUE (zip);

CREATE INDEX idx_cities_id ON zipcode USING btree (cities_id);

CREATE INDEX idx_cities_zipcode_country ON zipcode USING btree (country);

CREATE INDEX idx_zipcode_geom ON zipcode USING gist (geom);

CREATE INDEX idx_zipcode_gist_geom ON zipcode USING gist (geom);

CREATE INDEX zipcode_doc_vectors_idx ON zipcode USING gin (doc_vectors);

CREATE INDEX zipcode_doc_vectors_idx1 ON zipcode USING gin (doc_vectors);

CREATE INDEX zipcode_zip_idx ON zipcode USING btree (zip);

ALTER TABLE ONLY zipcode
    ADD CONSTRAINT zipcode_cities_id_fkey FOREIGN KEY (cities_id) REFERENCES cities(id);



ALTER TABLE "user_auth" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "student" ADD FOREIGN KEY ("school_id") REFERENCES "schools" ("id");

ALTER TABLE "session" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "session" ADD FOREIGN KEY ("tutor_id") REFERENCES "tutor" ("id");

ALTER TABLE "tutor_auth" ADD FOREIGN KEY ("id") REFERENCES "tutor" ("id");

ALTER TABLE "session_feedback" ADD FOREIGN KEY ("session_id") REFERENCES "session" ("id");

ALTER TABLE "user2subscription" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "user2subscription" ADD FOREIGN KEY ("subscription_id") REFERENCES "subscription" ("id");

ALTER TABLE "user_invitee" ADD FOREIGN KEY ("invitee") REFERENCES "user" ("id");

ALTER TABLE user2roles ADD CONSTRAINT unique_user_role UNIQUE (user_id, role_id);

ALTER TABLE user_invitee ADD CONSTRAINT unique_invitee_inviter UNIQUE (invitee, inviter);