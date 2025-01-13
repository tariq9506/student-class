ALTER TABLE public.user ADD COLUMN dialing_code varchar(5);

ALTER TABLE public.user ALTER COLUMN timezone TYPE varchar(55);