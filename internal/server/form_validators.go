package server

import (
	"strconv"
)

func (app *Application) validateLoginReq(body *LoginReq) map[string]string {

	errors := make(map[string]string)

	if body.Username == "" {
		errors["username"] = "Username must be at least 4 characters."
	} else if len(body.Username) < 4 {
		minCharsLeft := strconv.Itoa(4 - len(body.Username))
		errors["username"] = "You need at least " + minCharsLeft + " more chars in your username."

	}

	if body.Password == "" {
		errors["password"] = "Password must be at least 10 characters"
	} else if len(body.Password) < 10 {
		minCharsLeft := strconv.Itoa(10 - len(body.Password))
		errors["password"] = "You need at least " + minCharsLeft + " more chars in your password."
	}
	return errors
}

func (app *Application) validateRegisterReq(body *RegisterReq) map[string]string {
	errors := make(map[string]string)
	if body.StudentNumber == "" {
		errors["student_number"] = "Student number must be 5-7 numbers."
	} else if _, err := strconv.Atoi(body.StudentNumber); err != nil {
		errors["student_number"] = "Student number must be only numeric values"
	} else if len(body.StudentNumber) < 5 {
		minCharsLeft := strconv.Itoa(5 - len(body.StudentNumber))
		errors["student_number"] = "You need at least " + minCharsLeft + " more numbers in your student number."
	} else if len(body.StudentNumber) > 7 {
		overflow := strconv.Itoa(len(body.StudentNumber) - 7)
		errors["student_number"] = "Your student number has " + overflow + " too many numbers."
	}

	if body.Username == "" {
		errors["username"] = "Username must be at least 4 characters."
	} else if len(body.Username) < 4 {
		minCharsLeft := strconv.Itoa(4 - len(body.Username))
		errors["username"] = "You need at least " + minCharsLeft + " more chars in your username."
	} else {
		student, err := app.Store.Students.GetByUsername(body.Username)
    if err != nil {
      errors["server"] = "Sorry, something went seriously wrong on our end. Please try again in a sec."
    }else if student != nil {
			errors["username"] = "This username is already taken. Please choose another one"
		}
	}

	if body.Password == "" {
		errors["password"] = "Password must be at least 10 characters"
	} else if len(body.Password) < 10 {
		minCharsLeft := strconv.Itoa(10 - len(body.Password))
		errors["password"] = "You need at least " + minCharsLeft + " more chars in your password."
	}

	if body.JoinCode == "" {
		errors["join_code"] = "Join code is required"
	} else {
		course, err := app.Store.Courses.GetByJoinCode(body.JoinCode)
		if err != nil || course == nil {
			errors["join_code"] = "I could not find a course with that code"
		}
	}

	return errors
}

func (app *Application) validateCourseCreate(body *CreateCourseReq) map[string]string {
	errors := make(map[string]string)
	if body.JoinCode == "" {
		errors["join_code"] = "Join code required"

	}
	if body.Name == "" {
		errors["name"] = "Name required"
	}

	if body.Year == "" {
		errors["year"] = "Year must be YYYY"
	} else if _, err := strconv.Atoi(body.Year); err != nil {
		errors["year"] = "Year must all numbers of the form YYYY"
	}

	if body.Semester == "" {
		errors["semester"] = "Semester must be 1 or 2"
	} else if body.Semester != "1" && body.Semester != "2" {
		errors["semester"] = "Semester must be 1 or 2"
	}
	return errors
}
