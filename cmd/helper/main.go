package main

import (
	"bufio"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:@localhost/amics?sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Create teacher account
	// Get teacher data from stdio
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the teacher employee number\n> ")
	employee_number, _ := reader.ReadString('\n')
	fmt.Print("Enter the teacher username\n> ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Enter the teacher password\n> ")
	teacher_password, _ := reader.ReadString('\n')
	for len(teacher_password) < 10 {
		fmt.Print("Password must be 10 or more characters.\nEnter the teacher password\n> ")
		teacher_password, _ = reader.ReadString('\n')
	}

	// Hash the pass
	hashed, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(teacher_password)), bcrypt.DefaultCost)

	// Store it
	_, err = db.Exec("INSERT INTO teacher (employee_number, username, password) VALUES ($1,$2,$3);", strings.TrimSpace(employee_number), strings.TrimSpace(username), string(hashed))
	if err != nil {
		panic(err)
	}
}
