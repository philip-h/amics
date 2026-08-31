package model

import (
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/storage"
)

type AssignmentForm struct {
	Initial bool

	Id               string
	UnitName         string
	Name             string
	Description      string
	RequiredFilename string
	PytestCode       string
	Points           string
	DueDate          string
	Visible          bool
	CourseId         string

	ServerError string
}

func NewAssignmentFormGet() *AssignmentForm {
	return &AssignmentForm{
		Initial: true,
	}
}

func NewAssignmentFormPost(
	id string,
	unitName string,
	name string,
	description string,
	requiredFilename string,
	pytestCode string,
	points string,
	dueDate string,
	visible bool,
	courseId string,
) *AssignmentForm {
	return &AssignmentForm{
		Initial:          false,
		Id:               id,
		UnitName:         unitName,
		Name:             name,
		Description:      description,
		RequiredFilename: requiredFilename,
		PytestCode:       pytestCode,
		Points:           points,
		DueDate:          dueDate,
		Visible:          visible,
		CourseId:         courseId,
	}
}

func NewFromAssignment(assignment *storage.Assignment) *AssignmentForm {
	return &AssignmentForm{
		Initial:          false,
		Id:               strconv.Itoa(assignment.Id),
		UnitName:         assignment.UnitName,
		Name:             assignment.Name,
		Description:      assignment.Description,
		RequiredFilename: assignment.RequiredFilename,
		PytestCode:       assignment.PytestCode,
		Points:           strconv.Itoa(assignment.Points),
		DueDate:          assignment.DueDate.Format("2006-01-02T15:04"),
		Visible:          assignment.Visible,
		CourseId:         strconv.Itoa(assignment.CourseId),
	}
}

func (f *AssignmentForm) UnitNameProblem() string {
	if f.Initial {
		return ""
	}

	if f.UnitName == "" {
		return "Unit name is required"
	}
	return ""
}

func (f *AssignmentForm) NameProblem() string {
	if f.Initial {
		return ""
	}

	if f.Name == "" {
		return "Name is required"
	}
	return ""
}

func (f *AssignmentForm) DescriptionProblem() string {
	if f.Initial {
		return ""
	}

	if f.Description == "" {
		return "Description is required"
	}
	return ""
}

func (f *AssignmentForm) RequiredFilenameProblem() string {
	if f.Initial {
		return ""
	}

	if f.RequiredFilename == "" {
		return "Required filename is required"
	}
	return ""
}

func (f *AssignmentForm) PointsProblem() string {
	if f.Initial {
		return ""
	}

	points, err := strconv.Atoi(f.Points)
	if err != nil {
		return "Points must be a valid number"
	}
	if points <= 0 {
		return "Points must be greater than 0"
	}
	if points > 100 {
		return "Points must be less than or equal to 100"
	}
	return ""
}

func (f *AssignmentForm) DueDateProblem() string {
	if f.Initial {
		return ""
	}

	// Check if the due date is in the correct format
	_, err := time.Parse("2006-01-02T15:04", f.DueDate)
	if err != nil {
		return "Due date must be in valid datetime format"
	}
	return ""
}

func (f *AssignmentForm) CourseIdProblem() string {
	if f.Initial {
		return ""
	}

	if f.CourseId == "" {
		return "Course Id is required"
	}

	_, err := strconv.Atoi(f.CourseId)
	if err != nil {
		return "Course Id must be a valid number"
	}
	return ""
}

func (f *AssignmentForm) Validate() map[string]string {
	problems := make(map[string]string)

	if f.Initial {
		return problems
	}

	if f.UnitNameProblem() != "" {
		problems["UnitName"] = f.UnitNameProblem()
	}
	if f.NameProblem() != "" {
		problems["Name"] = f.NameProblem()
	}
	if f.DescriptionProblem() != "" {
		problems["Description"] = f.DescriptionProblem()
	}
	if f.RequiredFilenameProblem() != "" {
		problems["RequiredFilename"] = f.RequiredFilenameProblem()
	}
	if f.PytestCode == "" {
		problems["PytestCode"] = "Pytest code is required"
	}
	if f.PointsProblem() != "" {
		problems["Points"] = f.PointsProblem()
	}
	if f.DueDateProblem() != "" {
		problems["DueDate"] = f.DueDateProblem()
	}
	if f.CourseIdProblem() != "" {
		problems["CourseId"] = f.CourseIdProblem()
	}

	return problems
}
