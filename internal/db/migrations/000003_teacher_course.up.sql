CREATE TABLE teacher_course (
  teacher_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  course_id INTEGER NOT NULL REFERENCES course(id) ON DELETE CASCADE,
  PRIMARY KEY(teacher_id, course_id)
);
