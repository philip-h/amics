CREATE TABLE submission (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    assignment_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    grade INTEGER,
    comments TEXT,
    FOREIGN KEY (student_id) REFERENCES student(id),
    FOREIGN KEY (assignment_id) REFERENCES assignment(id),
    FOREIGN KEY (file_id) REFERENCES file(id)
);