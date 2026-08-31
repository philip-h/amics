package model

import "github.com/philip-h/amics/internal/storage"

type StudentData struct {
	Student *storage.Person

	Assignments []*storage.AssignmentWithGrade

	StudentAverage float64
}

func NewStudentData(student *storage.Person, assignments []*storage.AssignmentWithGrade, studentAverage float64) *StudentData {
	return &StudentData{
		Student:        student,
		Assignments:    assignments,
		StudentAverage: studentAverage,
	}
}
