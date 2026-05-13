package server

import (
	"strconv"

	"github.com/philip-h/amics/internal/storage"
)

func validateLoginReq(username, password string) (problems map[string]string) {
	problems = make(map[string]string)

	if username == "" {
		problems["username"] = "Username must be at least 4 characters."
	} else if len(username) < 4 {
		minCharsLeft := strconv.Itoa(4 - len(username))
		problems["username"] = "You need at least " + minCharsLeft + " more chars in your username."

	}

	if password == "" {
		problems["password"] = "Password must be at least 10 characters"
	} else if len(password) < 10 {
		minCharsLeft := strconv.Itoa(10 - len(password))
		problems["password"] = "You need at least " + minCharsLeft + " more chars in your password."
	}
	return
}

func validateRegisterReq(store *storage.Storage, studentNumber, firstName, username, password, joinCode string) (problems map[string]string) {
	problems = make(map[string]string)

	if studentNumber == "" {
		problems["student_number"] = "Student number must be 5-7 numbers."
	} else if _, err := strconv.Atoi(studentNumber); err != nil {
		problems["student_number"] = "Student number must be only numeric values"
	} else if len(studentNumber) < 5 {
		minCharsLeft := strconv.Itoa(5 - len(studentNumber))
		problems["student_number"] = "You need at least " + minCharsLeft + " more numbers in your student number."
	} else if len(studentNumber) > 7 {
		overflow := strconv.Itoa(len(studentNumber) - 7)
		problems["student_number"] = "Your student number has " + overflow + " too many numbers."
	}

	if firstName == "" {
		problems["first_name"] = "First name cannot be blank"
	}

	if username == "" {
		problems["username"] = "Username must be at least 4 characters."
	} else if len(username) < 4 {
		minCharsLeft := strconv.Itoa(4 - len(username))
		problems["username"] = "You need at least " + minCharsLeft + " more chars in your username."
	} else {
		student, err := store.People.GetByUsername(username)
		if err != nil {
			problems["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
		} else if student != nil {
			problems["username"] = "This username is already taken. Please choose another one"
		}
	}

	if password == "" {
		problems["password"] = "Password must be at least 10 characters"
	} else if len(password) < 10 {
		minCharsLeft := strconv.Itoa(10 - len(password))
		problems["password"] = "You need at least " + minCharsLeft + " more chars in your password."
	}

	if joinCode == "" {
		problems["join_code"] = "Join code is required"
	} else {
		course, err := store.Courses.GetByJoinCode(joinCode)
		if err != nil || course == nil {
			problems["join_code"] = "I could not find a course with that code"
		}
	}
	return
}

func validateCourseReq(id, joinCode, year, semester, name string) (problems map[string]string) {
	problems = make(map[string]string)

	if id != "new" {
		if _, err := strconv.Atoi(id); err != nil {
			problems["id"] = "Id must be a number"
		}
	}

	if joinCode == "" {
		problems["join_code"] = "Join code required"

	}
	if name == "" {
		problems["name"] = "Name required"
	}

	if year == "" {
		problems["year"] = "Year must be YYYY"
	} else if _, err := strconv.Atoi(year); err != nil {
		problems["year"] = "Year must be all numbers of the form YYYY"
	}

	if semester != "1" && semester != "2" {
		problems["semester"] = "Semester must be 1 or 2"
	}

	return
}

func validateAssignmentReq(
	unitName,
	name,
	description,
	requiredFilename,
	pytestCode,
	points,
	dueDate string) (problems map[string]string) {

	problems = make(map[string]string)

	if unitName == "" {
		problems["unitName"] = "Unit name cannot be blank"

	}

	if name == "" {
		problems["name"] = "Name cannot be blank"

	}

	if description == "" {
		problems["description"] = "Description cannot be blank"

	}

	if requiredFilename == "" {
		problems["requiredFilename"] = "RequiredFilename cannot be blank"

	}

	if pytestCode == "" {
		problems["pytestCode"] = "PytestCode cannot be blank"

	}

	if points == "" {
		problems["points"] = "Points cannot be blank"
	} else if pointsInt, err := strconv.Atoi(points); err != nil {
		problems["points"] = "Points must be a number"
	} else if pointsInt < 0 {
		problems["points"] = "Points cannot be negative"

	}

	if dueDate == "" {
		problems["dueDate"] = "Due date cannot be blank"
	}

	return
}
