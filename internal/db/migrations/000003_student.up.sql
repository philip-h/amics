CREATE TABLE student (
    id SERIAL PRIMARY KEY,
    student_number VARCHAR(10) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    course_id INTEGER NOT NULL REFERENCES course (id)
);
