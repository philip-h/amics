CREATE TABLE assignment_category (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES course (id),
    name TEXT NOT NULL,
    weight SMALLINT NOT NULL CHECK (weight BETWEEN 0 AND 100),
    UNIQUE (course_id, name)
);
