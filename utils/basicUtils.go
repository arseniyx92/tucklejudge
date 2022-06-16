package utils

import (
	"net/http"
	"html/template"
	"tucklejudge/utils/splayMap"
	"sync"
	"math/rand"
	"time"
	"os"
	"bufio"
	"fmt"
)

var RandomGen = rand.New(rand.NewSource(time.Now().UnixNano()))

var UserFilesMutex sync.Mutex

var LoginCookieStorage = &splayMap.SplayTree[string, string]{}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func CheckForValidStandardAccess(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("user_info")
	if err != nil || LoginCookieStorage.CheckNode(c.Value) == false {
		http.Redirect(w, r, "/login", http.StatusFound)
		return false
	}
	return true
}

func CheckForAuthorizationCapability(w http.ResponseWriter, r *http.Request) bool {
	c, err := r.Cookie("user_info")
	if err == nil && LoginCookieStorage.CheckNode(c.Value) == true {
		http.Redirect(w, r, "/", http.StatusFound)
		return false
	}
	return true
}

func RenderTemplate(w http.ResponseWriter, tmpl string, page interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type User struct {
	Username string
	Name string
	Surname string
	Teacher bool
	Grade string
	Letter string
	Password string
}

func (rg *User) Save() error {
	UserFilesMutex.Lock()
	defer UserFilesMutex.Unlock()

	// adding user to user.txt (usertlist)
	f, err := os.OpenFile("authentication/users.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s;%s;%s;%v;%s;%s;%s\n", rg.Username, rg.Name, rg.Surname, rg.Teacher, rg.Grade, rg.Letter, rg.Password))
	if err != nil {
		return err
	}

	// creating a new file for the user
	return os.WriteFile("authentication/users/"+rg.Username+".txt", []byte(fmt.Sprintf("Username: %s\nName: %s\nSurname: %s\nIs teacher: %v\nGrade: %s\nLetter: %s\nPassword: %s\n", rg.Username, rg.Name, rg.Surname, rg.Teacher, rg.Grade, rg.Letter, rg.Password)), 0600)
}

func GetAccauntInfo(username string) (*User, error) {
	UserFilesMutex.Lock()
	defer UserFilesMutex.Unlock()

	f, err := os.Open("authentication/users/"+username+".txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	var user = &User{}

	scanner.Scan()
	user.Username = scanner.Text()[len("Username: "):]
	scanner.Scan()
	user.Name = scanner.Text()[len("Name: "):]
	scanner.Scan()
	user.Surname = scanner.Text()[len("Surname: "):]
	scanner.Scan()
	if scanner.Text()[len("Is teacher: "):] == "true" {
		user.Teacher = true
	} else {
		user.Teacher = false
	}
	scanner.Scan()
	user.Grade = scanner.Text()[len("Grade: "):]
	scanner.Scan()
	user.Letter = scanner.Text()[len("Letter: "):]
	scanner.Scan()
	user.Password = scanner.Text()[len("Password: "):]

	return user, scanner.Err()
}








