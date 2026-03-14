-- https://stackoverflow.com/a/48382296
DO $$ BEGIN
    CREATE TYPE status_type AS ENUM ('grading', 'completed', 'failure') ;
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
CREATE TABLE submission (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES student (id),
    assignment_id INTEGER NOT NULL REFERENCES assignment (id),
    code TEXT NOT NULL,
    grade SMALLINT,
    submitted_on BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM now()),
    comments TEXT,
    status status_type,
    graded_on BIGINT
);
