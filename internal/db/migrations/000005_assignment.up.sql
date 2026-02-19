CREATE TABLE assignment(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    unit_name TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    required_filename TEXT NOT NULL,
    points INTEGER NOT NULL,
    due_date TEXT NOT NULL,
    visible BOOLEAN NOT NULL,
    pytest_file_id INTEGER NOT NULL,
    course_id INTEGER NOT NULL,
    FOREIGN KEY (pytest_file_id) REFERENCES file(id),
    FOREIGN KEY (course_id) REFERENCES course(id)
);