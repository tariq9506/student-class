ALTER TABLE tutor
ADD COLUMN timezone VARCHAR(10),
ADD COLUMN teaches_to VARCHAR(50)[] CHECK (ARRAY['Elementary','Intermediate','High School']::VARCHAR[] @> teaches_to),
ADD COLUMN zoom_link VARCHAR;

ALTER TABLE session
ADD COLUMN game_url varchar(100);

COMMENT ON COLUMN session.game_url IS 'link of the game that userr has created';
COMMENT ON COLUMN tutor.timezone IS 'timezone of the tutor which would be added by admin while tutor onboarding';
COMMENT ON COLUMN tutor.teaches_to IS 'grade of the students upto which this tutor can teach';
COMMENT ON COLUMN tutor.zoom_link IS 'zoom link for students to join';

ALTER TABLE session
    DROP CONSTRAINT session_status_check,
    ADD CONSTRAINT session_status_check CHECK (status IN ('active', 'started', 'cancelled', 'completed', 'absent'));