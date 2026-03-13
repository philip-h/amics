CREATE TABLE assignment(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    unit_name TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    required_filename TEXT NOT NULL,
    pytest_code TEXT NOT NULL,
    points INTEGER NOT NULL,
    due_date INT NOT NULL,
    visible BOOLEAN NOT NULL,
    course_id INTEGER NOT NULL,
    FOREIGN KEY (course_id) REFERENCES course(id)
);
