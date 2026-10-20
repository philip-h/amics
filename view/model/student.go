package model

import "github.com/philip-h/amics/internal/storage"

type StudentData struct {
	Student *storage.Person

	Assignments []*storage.AssignmentWithGrade

	StudentAverage string
}

func NewStudentData(student *storage.Person, assignments []*storage.AssignmentWithGrade, studentAverage string) *StudentData {
	return &StudentData{
		Student:        student,
		Assignments:    assignments,
		StudentAverage: studentAverage,
	}
}
