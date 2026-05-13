CREATE TABLE person (
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'student'
);
