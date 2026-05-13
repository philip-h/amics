CREATE TABLE student_course (
  student_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  course_id INTEGER NOT NULL REFERENCES course(id) ON DELETE CASCADE,
  -- currently I only want any student enrolled in at most 1 class
  -- it's not a field in person because of teachers
  PRIMARY KEY(student_id)
);
