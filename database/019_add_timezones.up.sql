CREATE Table timezones (
    id SERIAL PRIMARY KEY,
    identifier text,
    abbreviation varchar(20),
    country_name varchar(100),
    country_code varchar (4) 
);


