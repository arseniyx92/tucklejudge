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
	"strings"
	"io"
	"errors"
	"image"
	_ "image/png"
	"image/png"
)

var RandomGen = rand.New(rand.NewSource(time.Now().UnixNano()))

var UserFilesMutex sync.Mutex

var IDtoUsername = &splayMap.SplayTree[int, string]{}
var LoginCookieStorage = &splayMap.SplayTree[string, string]{}
var VerificationCode string // length = 6

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
	ChangeVerificationCode()
}

func ChangeVerificationCode() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	VerificationCode = ""
	for i := 0; i < 6; i++ {
		currentChar := rng.Int63()%36
		if currentChar >= 26 {
			VerificationCode += string(currentChar-26+'0')
		} else {
			VerificationCode += string(currentChar+'A');
		}
	}
}

func GetUsernameByID(string_id string) (string, error) {
	id, err := strconv.Atoi(string_id)
	if (err != nil) {
		return "", err
	}
	username, ok := IDtoUsername.ReturnNodeValue(id)
	if ok == false {
		return "", errors.New(fmt.Sprintf("Username with such (%s) userID does not exist", string_id))
	}
	return username, nil
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

func GetCurrentlyFreeID(folderPath string, maxChars int) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s/currentID.txt", folderPath))
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	id, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return "", err
	}
	f.Close()
	string_id := fmt.Sprintf("%d", id)
	var ID string
	for len(ID)+len(string_id) < maxChars {
		ID += "0"
	}
	ID += string_id
	os.WriteFile(fmt.Sprintf("%s/currentID.txt", folderPath), []byte(fmt.Sprintf("%d", id+1)), 0600)
	return ID, nil
}

func SaveFormFileToSrc(r *http.Request) (string, error) {
	// receiving image info from http.Request
	in, header, err := r.FormFile("file")
	if err != nil {
		return "", err // maybe http.Redirect(w, r, "/", http.StatusFound)
	}
	defer in.Close()
	// getting currently free ID for a new image
	UserFilesMutex.Lock()
	fileName, err := GetCurrentlyFreeID("src", 12)
	UserFilesMutex.Unlock()
	if err != nil {
		return "", err
	}
	// getting file extension
	strsForGettingCorrectExtension := strings.Split(header.Filename, ".")
	fileName += "." + strsForGettingCorrectExtension[len(strsForGettingCorrectExtension)-1]
	// creating new image
	out, err := os.Create("src/"+fileName)
	if err != nil {
		return "", err
	}
	defer out.Close()
    io.Copy(out, in)
    return fileName, nil
}

func (rg *User) Create() error {
	UserFilesMutex.Lock()
	defer UserFilesMutex.Unlock()

	// getting currently free ID
	s, err := GetCurrentlyFreeID("authentication", 4)
	rg.ID = s
	if err != nil {
		return err
	}
	id, err := strconv.Atoi(rg.ID)
	if err != nil {
		return err
	}
	// f, err := os.Open("authentication/currentID.txt")
	// if err != nil {
	// 	return err
	// }
	// scanner := bufio.NewScanner(f)
	// scanner.Scan()
	// id, err := strconv.Atoi(scanner.Text())
	// if err != nil {
	// 	return err
	// }
	// f.Close()
	// string_id := fmt.Sprintf("%d", id)
	// for len(rg.ID)+len(string_id) < 4 {
	// 	rg.ID += "0"
	// }
	// rg.ID += string_id
	// os.WriteFile("authentication/currentID.txt", []byte(fmt.Sprintf("%d", id+1)), 0600)

	// adding ID to local memory
	IDtoUsername.AddNode(id, rg.Username)

	// adding user to user.txt (usertlist)
	f, err := os.OpenFile("authentication/users.txt", os.O_APPEND|os.O_WRONLY, 0600)
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
	for i, s := range rg.Tests {
		tests_string += s
		if i+1 != len(rg.Tests) {
			tests_string += " "
		}
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
	tests_strings := strings.Split(scanner.Text()[len("Tests: "):], " ")
	if len(scanner.Text()[len("Tests: "):]) == 0 {
		tests_strings = make([]string, 0)
	}
	user.Tests = make([]string, len(tests_strings))
	for i, str := range tests_strings {
		user.Tests[i] = str
	}

	return user, scanner.Err()
}

type Question struct {
	Answer string
	Points int
	IndexForTemplate int
}

type Test struct {
	ID string
	Name string
	Questions []Question
	PointsToMark [3]int // < 2, 3, 4
	NumberOfQuestionsForTemplate int
}

var TestFilesMutex sync.Mutex

func (test *Test) CreateIDAndSave() error {
	TestFilesMutex.Lock()
	defer TestFilesMutex.Unlock()

	// getting current test ID
	string_id, err := GetCurrentlyFreeID("tester", 4)
	test.ID = string_id
	if err != nil {
		return err
	}
	// f, err := os.Open("tester/currentID.txt")
	// if err != nil {
	// 	return err
	// }
	// scanner := bufio.NewScanner(f)
	// scanner.Scan()
	// id, err := strconv.Atoi(scanner.Text())
	// if err != nil {
	// 	return err
	// }
	// f.Close()
	// string_id := fmt.Sprintf("%d", id)
	// for len(test.ID)+len(string_id) < 4 {
	// 	test.ID += "0"
	// }
	// test.ID += string_id
	// os.WriteFile("tester/currentID.txt", []byte(fmt.Sprintf("%d", id+1)), 0600)

	// creating new test file
	return SaveTestToFile(test)
}

func SaveTestToFile(test *Test) error {
	testInfo := fmt.Sprintf("Name: %s\nQuestions (%d)\n", test.Name, len(test.Questions))
	for i, q := range test.Questions {
		testInfo += fmt.Sprintf("Question %d.\n%s\n%d\n", i, q.Answer, q.Points)
	}
	testInfo += "Points to mark: 2, 3, 4\n"
	for _, q := range test.PointsToMark {
		testInfo += fmt.Sprintf("%d\n", q)
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
	for i, _ := range test.Questions {
		q := &test.Questions[i]
		scanner.Scan()
		scanner.Scan()
		q.Answer = scanner.Text()
		scanner.Scan()
		q.Points, err = strconv.Atoi(scanner.Text())
		if err != nil {
			return test, err
		}
	}
	scanner.Scan() // scanning "Points to mark: 2, 3, 4"
	for i, _ := range test.PointsToMark {
		scanner.Scan()
		test.PointsToMark[i], _ = strconv.Atoi(scanner.Text())
	}
	scanner.Scan()
	return test, nil
}

func CheckForTeacher(r *http.Request) bool {
	c, _ := r.Cookie("user_info")
	UserFilesMutex.Lock()
	username, _ := LoginCookieStorage.ReturnNodeValue(c.Value)
	UserFilesMutex.Unlock()
	user, err := GetAccauntInfo(username)
	if err != nil || !user.Teacher {
		return false
	}
	return true
}


func CheckForAdmin(r *http.Request) bool {
	c, _ := r.Cookie("user_info")
	UserFilesMutex.Lock()
	username, _ := LoginCookieStorage.ReturnNodeValue(c.Value)
	UserFilesMutex.Unlock()
	if username == "_admin" {
		return true
	}
	return false
}

func AddTestToUsersList(username string, testID string) error {
	user, err := GetAccauntInfo(username)
	if err != nil {
		return err
	}
	for _, test := range user.Tests {
		if test == testID {
			return nil
		}
	}
	user.Tests = append(user.Tests, testID)
	return user.Save()
}

func DeleteTestFromUsersList(username string, testID string) error {
	user, err := GetAccauntInfo(username)
	if err != nil {
		return err
	}
	for i := 0; i < len(user.Tests); i++ {
		if user.Tests[i] == testID {
			user.Tests[i], user.Tests[len(user.Tests)-1] = user.Tests[len(user.Tests)-1], user.Tests[i]
			user.Tests = user.Tests[:len(user.Tests)-1]
			i--
		}
	}
	return user.Save()
}

func GetTestUsersResultByID(testID string, username string) (string, error) {
	b, err := os.ReadFile(fmt.Sprintf("tester/testResults/%s$%s.txt", testID, username))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type PersonalQuestion struct {
	Index string
	UserAnswer string
	CorrectAnswer string
	Points string
}

type PersonalTest struct {
	UserName string
	TestName string
	Mark string
	InputImageName string
	ProcessedImageName string
	Questions []PersonalQuestion
	PointsSum string
	PointsToMark [3]string
}

func CreateTestResultFile(personalTestName string, results *PersonalTest) error {
	filePath := fmt.Sprintf("tester/testResults/%s.txt", personalTestName)
	var out string
	out += fmt.Sprintf("TestName: %s\n", results.TestName)
	out += fmt.Sprintf("Mark: %s\n", results.Mark)
	out += fmt.Sprintf("Input image name: %s\n", results.InputImageName)
	out += fmt.Sprintf("Processed image name: %s\n", results.ProcessedImageName)
	out += fmt.Sprintf("Questions (%d)\n", len(results.Questions))

	for _, q := range results.Questions {
		out += fmt.Sprintf("%s) %s %s %s\n", q.Index, q.UserAnswer, q.CorrectAnswer, q.Points)
	}

	out += fmt.Sprintf("Points sum: %s\n", results.PointsSum)
	out += "Points to mark: 2, 3, 4\n"
	out += fmt.Sprintf("%s\n%s\n%s\n", results.PointsToMark[0], results.PointsToMark[1], results.PointsToMark[2])

	return os.WriteFile(filePath, []byte(out), 0600)
}

type PersonalResult struct {
	TestID, Username, FullName, Mark string
	IndexForTemplate int
}

type ShortTestResultsInfo struct {
	Results []PersonalResult
	IDForTemplate string
}

func SaveShortResultsInfoToFile(filename string, results *ShortTestResultsInfo) error {
	out := fmt.Sprintf("Results (%d)\n", len(results.Results))
	for _, r := range results.Results {
		out += fmt.Sprintf("%s %s %s %s\n", r.TestID, r.Username, r.FullName, r.Mark)
	}
	return os.WriteFile(fmt.Sprintf("tester/teacherTestResults/%s.txt", filename), []byte(out), 0600)
}

func LoadShortResultsFromFile(filename string) (*ShortTestResultsInfo, error) {
	byte_in, err := os.ReadFile(fmt.Sprintf("tester/teacherTestResults/%s.txt", filename))
	if err != nil {
		return nil, err
	}
	in := strings.Split(string(byte_in), "\n")
	n, err := strconv.Atoi(in[0][len("Results ("):len(in[0])-1])
	results := &ShortTestResultsInfo {
		Results: make([]PersonalResult, n),
		IDForTemplate: filename,
	}
	for i := range results.Results {
		cur_line := strings.Split(in[i+1], " ")
		results.Results[i].TestID = cur_line[0]
		results.Results[i].Username = cur_line[1]
		results.Results[i].FullName = cur_line[2] + " " + cur_line[3]
		results.Results[i].Mark = cur_line[4]
		results.Results[i].IndexForTemplate = i+1
	}
	return results, nil
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func ClearAllData(w http.ResponseWriter, r *http.Request) {
	if CheckForValidStandardAccess(w, r) == false {
		return
	}
	if CheckForAdmin(r) == false {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// clear all users and set currentID to zero
	admin, _ := GetAccauntInfo("_admin")
	admin.Tests = make([]string, 0);
	admin.Save()
	b, err := os.ReadFile("authentication/users/_admin.txt")
	if err != nil {
		panic(err)
	}
	Must(os.RemoveAll("authentication/users"))
	Must(os.Mkdir("authentication/users", 0755))
	Must(os.WriteFile("authentication/currentID.txt", []byte("1"), 0600))
	os.WriteFile("authentication/users/_admin.txt", b, 0600)
	os.WriteFile("authentication/users.txt", []byte("_admin\n"), 0600)
	// clear all tests and set currentID to zero
	Must(os.RemoveAll("tester/tests"))
	Must(os.Mkdir("tester/tests", 0755))
	Must(os.WriteFile("tester/currentID.txt", []byte("0"), 0600))
	// clear all teacherTestResults and set currentID to zero
	Must(os.RemoveAll("tester/teacherTestResults"))
	Must(os.Mkdir("tester/teacherTestResults", 0755))
	Must(os.WriteFile("tester/teacherTestResults/currentID.txt", []byte("0"), 0600))
	// clear all testResults
	Must(os.RemoveAll("tester/testResults"))
	Must(os.Mkdir("tester/testResults", 0755))
	// clear all src and set currentID to zero
	Must(os.RemoveAll("src"))
	Must(os.Mkdir("src", 0755))
	Must(os.WriteFile("src/currentID.txt", []byte("0"), 0600))
}


func SaveImageToSrc(img image.Image) string {
	// getting currently free ID for a new image
	UserFilesMutex.Lock()
	fileName, _ := GetCurrentlyFreeID("src", 12)
	UserFilesMutex.Unlock()
	fileName += ".png"
	filepath := "src/"+fileName
	// creating new image
    if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
	}
	f, _ := os.Create(filepath)
	_ = png.Encode(f, img)
	return fileName
}
