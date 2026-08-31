package model

import (
	"regexp"
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/storage"
)

type CourseForm struct {
	Initial bool

	Id         string
	CourseCode string
	Section    string
	Name       string
	Year       string
	Semester   string
	JoinCode   string

	ServerError string
}

func NewCourseFormGet() *CourseForm {
	return &CourseForm{
		Initial: true,
	}
}

func NewCourseFormPost(
	id string,
	courseCode string,
	section string,
	name string,
	year string,
	semester string,
	joinCode string,
) *CourseForm {
	return &CourseForm{
		Initial:    false,
		Id:         id,
		CourseCode: courseCode,
		Section:    section,
		Name:       name,
		Year:       year,
		Semester:   semester,
		JoinCode:   joinCode,
	}
}

func NewFromCourse(course *storage.Course) *CourseForm {
	return &CourseForm{
		Initial:    false,
		Id:         strconv.Itoa(course.Id),
		CourseCode: course.CourseCode,
		Section:    strconv.Itoa(course.Section),
		Name:       course.Name,
		Year:       strconv.Itoa(course.Year),
		Semester:   strconv.Itoa(course.Semester),
		JoinCode:   course.JoinCode,
	}
}

func (f *CourseForm) CourseCodeProblem() string {
	if f.Initial {
		return ""
	}

	if f.CourseCode == "" {
		return "Course code is required"
	}

	re := regexp.MustCompile(`^[A-Z]{3}[1-4][UMCO]$`)
	if !re.MatchString(f.CourseCode) {
		return "Course code does not match ontario code format."
	}
	return ""
}

func (f *CourseForm) SectionProblem() string {
	if f.Initial {
		return ""
	}

	if f.Section != "1" && f.Section != "2" {
		return "Section must be 1 or 2"
	}
	return ""
}

func (f *CourseForm) NameProblem() string {
	if f.Initial {
		return ""
	}

	if f.Name == "" {
		return "Name is required"
	}
	return ""
}

func (f *CourseForm) JoinCodeProblem() string {
	if f.Initial {
		return ""
	}

	if f.JoinCode == "" {
		return "Join code is required"
	}
	return ""
}

func (f *CourseForm) SemesterProblem() string {
	if f.Initial {
		return ""
	}

	if f.Semester != "1" && f.Semester != "2" {
		return "Semester must be 1 or 2"
	}
	return ""
}

func (f *CourseForm) YearProblem() string {
	if f.Initial {
		return ""
	}

	// get current year
	thisYear := strconv.Itoa(time.Now().Year())
	nextYear := strconv.Itoa(time.Now().Year() + 1)

	if f.Year != thisYear && f.Year != nextYear {
		return "Year must be " + nextYear + " or " + thisYear
	}
	return ""
}

func (f *CourseForm) Validate() map[string]string {
	problems := make(map[string]string)

	if f.Initial {
		return problems
	}

	if f.CourseCodeProblem() != "" {
		problems["course-code"] = f.CourseCodeProblem()
	}
	if f.SectionProblem() != "" {
		problems["section"] = f.SectionProblem()
	}
	if f.NameProblem() != "" {
		problems["name"] = f.NameProblem()
	}
	if f.SemesterProblem() != "" {
		problems["semester"] = f.SemesterProblem()
	}
	if f.YearProblem() != "" {
		problems["year"] = f.YearProblem()
	}
	if f.JoinCodeProblem() != "" {
		problems["join-code"] = f.JoinCodeProblem()
	}

	return problems
}
