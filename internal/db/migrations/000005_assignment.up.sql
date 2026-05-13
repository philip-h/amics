CREATE TABLE assignment(
    id SERIAL PRIMARY KEY,
    unit_name TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    required_filename TEXT  NOT NULL,
    pytest_code TEXT NOT NULL,
    points SMALLINT NOT NULL,
    due_date TIMESTAMPTZ NOT NULL,
    visible BOOLEAN NOT NULL,
    course_id INTEGER NOT NULL REFERENCES course (id)
);
