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
	"strconv"
)

var RandomGen = rand.New(rand.NewSource(time.Now().UnixNano()))

var UserFilesMutex sync.Mutex

var IDtoUsername = &splayMap.SplayTree[int, string]{}
var LoginCookieStorage = &splayMap.SplayTree[string, string]{}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func Init() {
	// initializing IDs to Users (using users.txt)
	f, err := os.Open("authentication/users.txt")
	if err != nil {
		panic(err.Error())
	}
	scanner := bufio.NewScanner(f)
	for id := 0; scanner.Scan(); id++ {
		IDtoUsername.AddNode(id, scanner.Text())
	}
}

func GetUsername(r *http.Request) (string, error) {
	c, err := r.Cookie("user_info")
	if err != nil {
		return "", err
	}
	UserFilesMutex.Lock()
	username, _ := LoginCookieStorage.ReturnNodeValue(c.Value)
	UserFilesMutex.Unlock()
	return username, nil
}

func CheckForValidStandardAccess(w http.ResponseWriter, r *http.Request) bool {
	UserFilesMutex.Lock()
	defer UserFilesMutex.Unlock()
	c, err := r.Cookie("user_info")
	if err != nil || LoginCookieStorage.CheckNode(c.Value) == false {
		http.Redirect(w, r, "/login", http.StatusFound)
		return false
	}
	return true
}

func CheckForAuthorizationCapability(w http.ResponseWriter, r *http.Request) bool {
	UserFilesMutex.Lock()
	defer UserFilesMutex.Unlock()
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
	ID string
	Username string
	Name string
	Surname string
	Teacher bool
	Grade string
	Letter string
	Password string
	Tests []string
}

func (rg *User) Create() error {
	UserFilesMutex.Lock()
	defer UserFilesMutex.Unlock()

	// getting currently free ID
	f, err := os.Open("authentication/currentID.txt")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	id, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return err
	}
	f.Close()
	string_id := fmt.Sprintf("%d", id)
	for len(rg.ID)+len(string_id) < 4 {
		rg.ID += "0"
	}
	rg.ID += string_id
	os.WriteFile("authentication/currentID.txt", []byte(fmt.Sprintf("%d", id+1)), 0600)

	// adding ID to local memory
	IDtoUsername.AddNode(id, rg.Username)

	// adding user to user.txt (usertlist)
	f, err = os.OpenFile("authentication/users.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s\n", rg.Username))
	if err != nil {
		return err
	}

	// merging test ids to one string
	tests_string := ""
	for _, s := range rg.Tests {
		tests_string += s + " "
	}

	// creating a new file for the user
	return os.WriteFile("authentication/users/"+rg.Username+".txt", []byte(fmt.Sprintf("ID: %s\nUsername: %s\nName: %s\nSurname: %s\nIs teacher: %v\nGrade: %s\nLetter: %s\nPassword: %s\nTests: %s", rg.ID, rg.Username, rg.Name, rg.Surname, rg.Teacher, rg.Grade, rg.Letter, rg.Password, tests_string)), 0600)
}

func (rg *User) Save() error {
	UserFilesMutex.Lock()
	defer UserFilesMutex.Unlock()

	// merging test IDs to one string
	tests_string := ""
	for _, s := range rg.Tests {
		tests_string += s + " "
	}

	// updating user's file
	return os.WriteFile("authentication/users/"+rg.Username+".txt", []byte(fmt.Sprintf("ID: %s\nUsername: %s\nName: %s\nSurname: %s\nIs teacher: %v\nGrade: %s\nLetter: %s\nPassword: %s\nTests: %s", rg.ID, rg.Username, rg.Name, rg.Surname, rg.Teacher, rg.Grade, rg.Letter, rg.Password, tests_string)), 0600)
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
	user.ID = scanner.Text()[len("ID: "):]
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
	scanner.Scan()
	tests_string := string(scanner.Text()[len("Tests: "):])
	for i := 0; i < len(tests_string); i += 5 {
		str := ""
		for j := 0; j < 4; j++ {
			str += string(tests_string[i+j])
		}
		user.Tests = append(user.Tests, str)
	}

	return user, scanner.Err()
}

type Question struct {
	Answer string
	Points int
	Punishment int
}

type Test struct {
	ID string
	Name string
	Questions []Question
}

var TestFilesMutex sync.Mutex

func (test *Test) CreateIDAndSave() error {
	TestFilesMutex.Lock()
	defer TestFilesMutex.Unlock()

	// getting current test ID
	f, err := os.Open("tester/currentID.txt")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	id, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return err
	}
	f.Close()
	string_id := fmt.Sprintf("%d", id)
	for len(test.ID)+len(string_id) < 4 {
		test.ID += "0"
	}
	test.ID += string_id
	os.WriteFile("tester/currentID.txt", []byte(fmt.Sprintf("%d", id+1)), 0600)

	// creating new test file
	testInfo := fmt.Sprintf("Name: %s\nQuestions (%d)\n", test.Name, len(test.Questions))
	for i, q := range test.Questions {
		testInfo += fmt.Sprintf("Question %d.\n%s\n%d\n%d\n", i, q.Answer, q.Points, q.Punishment)
	}

	return os.WriteFile(fmt.Sprintf("tester/tests/%s.txt", test.ID), []byte(testInfo), 0600)
}

func GetTestByID(id string) (Test, error) {
	var test Test
	test.ID = id

	TestFilesMutex.Lock()
	defer TestFilesMutex.Unlock()

	f, err := os.Open(fmt.Sprintf("tester/tests/%s.txt", id))
	if err != nil {
		return test, err
	}
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	test.Name = scanner.Text()[len("Name: "):]
	scanner.Scan()
	n, err := strconv.Atoi(scanner.Text()[len("Questions ("):len(scanner.Text())-1])
	if err != nil {
		return test, err
	}
	test.Questions = make([]Question, n)
	for _, q := range test.Questions {
		scanner.Scan()
		scanner.Scan()
		q.Answer = scanner.Text()
		scanner.Scan()
		q.Points, err = strconv.Atoi(scanner.Text())
		if err != nil {
			return test, err
		}
		scanner.Scan()
		q.Punishment, err = strconv.Atoi(scanner.Text())
		if err != nil {
			return test, err
		}
	}
	return test, nil
}

func AddTestToTeachersList(username string, testID string) error {
	user, err := GetAccauntInfo(username)
	if err != nil {
		return err
	}
	user.Tests = append(user.Tests, testID)
	return user.Save()
}



