package model

import (
	"strconv"
)

type LoginForm struct {
	Initial     bool
	Username    string
	Password    string
	ServerError string
}

func NewLoginFormGet() *LoginForm {
	return &LoginForm{
		Initial: true,
	}
}

func NewLoginFormPost(
	username string,
	password string,
) *LoginForm {
	return &LoginForm{
		Initial:  false,
		Username: username,
		Password: password,
	}
}

func (f *LoginForm) UsernameProblem() string {
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

	}
	return ""
}

func (f *LoginForm) PasswordProblem() string {
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

func (f *LoginForm) Validate() map[string]string {
	problems := make(map[string]string)

	if f.Initial {
		return problems
	}

	if f.UsernameProblem() != "" {
		problems["username"] = f.UsernameProblem()
	}
	if f.PasswordProblem() != "" {
		problems["password"] = f.PasswordProblem()
	}
	return problems
}
