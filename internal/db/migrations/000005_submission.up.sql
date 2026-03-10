CREATE TABLE submission (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    assignment_id INTEGER NOT NULL,
    code TEXT NOT NULL,
    grade INTEGER,
    submitted_on TEXT NOT NULL,
    comments TEXT,
    status TEXT,
    graded_on TEXT,
    FOREIGN KEY (student_id) REFERENCES student(id),
    FOREIGN KEY (assignment_id) REFERENCES assignment(id)
);
