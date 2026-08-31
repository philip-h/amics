CREATE TABLE IF NOT EXISTS course (
    id SERIAL PRIMARY KEY,
    course_code TEXT NOT NULL,
    section SMALLINT NOT NULL,
    name TEXT NOT NULL,
    year SMALLINT NOT NULL,
    semester SMALLINT NOT NULL,
    join_code VARCHAR(20) NOT NULL UNIQUE
);
