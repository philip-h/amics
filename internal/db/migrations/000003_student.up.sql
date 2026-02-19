CREATE TABLE student (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_number TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    course_id INTEGER NOT NULL,
    FOREIGN KEY (course_id) REFERENCES course(id)
);