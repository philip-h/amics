CREATE TABLE assignment(
    id SERIAL PRIMARY KEY,
    unit_name VARCHAR(20) NOT NULL,
    name VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    required_filename VARCHAR(20)  NOT NULL,
    pytest_code TEXT NOT NULL,
    points SMALLINT NOT NULL,
    due_date BIGINT NOT NULL,
    visible BOOLEAN NOT NULL,
    course_id INTEGER NOT NULL REFERENCES course (id)
);
