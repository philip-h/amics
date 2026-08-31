package model

import (
	"context"
	"strconv"

	"github.com/philip-h/amics/internal/storage"
)

type RegisterForm struct {
	Initial bool

	FirstName     string
	StudentNumber string
	Username      string
	Password      string
	JoinCode      string

	ServerError string

	ctx   context.Context
	store *storage.Storage
}

func NewRegisterFormGet() *RegisterForm {
	return &RegisterForm{
		Initial: true,
	}
}

func NewRegisterFormPost(
	ctx context.Context,
	store *storage.Storage,
	firstName string,
	studentNumber string,
	username string,
	password string,
	joinCode string,
) *RegisterForm {
	return &RegisterForm{
		Initial:       false,
		FirstName:     firstName,
		StudentNumber: studentNumber,
		Username:      username,
		Password:      password,
		JoinCode:      joinCode,
		ctx:           ctx,
		store:         store,
	}
}

func (f *RegisterForm) FirstNameProblem() string {
	if f.Initial {
		return ""
	}

	if f.FirstName == "" {
		return "First name is required"
	}
	return ""
}

func (f *RegisterForm) StudentNumberProblem() string {
	if f.Initial {
		return ""
	}
	if f.StudentNumber == "" {
		return "Student number must be 5-7 numbers."
	} else if _, err := strconv.Atoi(f.StudentNumber); err != nil {
		return "Student number must be only numeric values"
	} else if len(f.StudentNumber) < 5 {
		minCharsLeft := strconv.Itoa(5 - len(f.StudentNumber))
		return "You need at least " + minCharsLeft + " more numbers in your student number."
	} else if len(f.StudentNumber) > 7 {
		overflow := strconv.Itoa(len(f.StudentNumber) - 7)
		return "Your student number has " + overflow + " too many numbers."
	}
	return ""
}

func (f *RegisterForm) UsernameProblem() string {
	if f.Initial {
		return ""
	}

	if f.Username == "" {
		return "Username must be at least 4 characters."
	} else if len(f.Username) < 4 {
		minCharsLeft := strconv.Itoa(4 - len(f.Username))
		chars := "chars"
		if minCharsLeft == "1" {
			chars = "char"
		}
		return "Username needs " + minCharsLeft + " more " + chars
	} else {
		student, err := f.store.People.GetByUsername(f.ctx, f.Username)
		if err != nil {
			f.ServerError = "Sorry, something went seriously wrong on our end. Please try again in a sec."
			return ""
		} else if student != nil {
			return "This username is already taken. Please choose another one"
		}
	}
	return ""
}

func (f *RegisterForm) PasswordProblem() string {
	if f.Initial {
		return ""
	}

	if f.Password == "" {
		return "Password must be at least 10 characters"
	} else if len(f.Password) < 10 {
		minCharsLeft := strconv.Itoa(10 - len(f.Password))
		chars := "chars"
		if minCharsLeft == "1" {
			chars = "char"
		}
		return "Password needs " + minCharsLeft + " more " + chars
	}
	return ""
}

func (f *RegisterForm) JoinCodeProblem() string {
	if f.Initial {
		return ""
	}

	if f.JoinCode == "" {
		return "Join code is required"
	} else {
		course, err := f.store.Courses.GetByJoinCode(f.ctx, f.JoinCode)
		if err != nil || course == nil {
			return "I could not find a course with that code"
		}
	}
	return ""
}

func (f *RegisterForm) Validate() map[string]string {
	problems := make(map[string]string)

	if f.Initial {
		return problems
	}

	if f.FirstNameProblem() != "" {
		problems["first-name"] = f.FirstNameProblem()
	}
	if f.StudentNumberProblem() != "" {
		problems["student-number"] = f.StudentNumberProblem()
	}
	if f.UsernameProblem() != "" {
		problems["username"] = f.UsernameProblem()
	}
	if f.PasswordProblem() != "" {
		problems["password"] = f.PasswordProblem()
	}
	if f.JoinCodeProblem() != "" {
		problems["join-code"] = f.JoinCodeProblem()
	}

	return problems
}
