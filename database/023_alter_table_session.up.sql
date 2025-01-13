ALTER TABLE public.session ADD COLUMN credited BOOL DEFAULT TRUE;

COMMENT ON COLUMN public.session.credited IS 'This indicates if money is given for this session.';
