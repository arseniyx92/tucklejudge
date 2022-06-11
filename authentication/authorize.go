package authentication

import (
	"os"
	"net/http"
	"fmt"
	"crypto/sha256"
	"strings"
	"bufio"
	"time"
	"tucklejudge/utils"
)

func generateCookie(w http.ResponseWriter, username string, password string) {
	key := fmt.Sprintf("%x", createPasswordHash(username+password))
	cookie := http.Cookie {
		Name: "user_info",
		Value: key,
		Expires: time.Now().Add(48*time.Hour),
		Path: "/",
	}
	utils.LoginCookieStorage.AddNode(key, username)
	http.SetCookie(w, &cookie)
}

func readSpecificLineFromFile(filepath string, line int) (string, error) {
	utils.UserFilesMutex.Lock()
	defer utils.UserFilesMutex.Unlock()

	f, err := os.Open(filepath)
	if (err != nil) {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var current_line int
	for scanner.Scan() {
		current_line++
		if current_line == line {
			return scanner.Text(), nil
		}
	}
	return "", scanner.Err()
}

type Login struct {
	Message string
	Prev_username string
}

type Registration struct {
	Message string
	Prev_username string
	Prev_name string
	Prev_surname string
	Prev_grade byte
	Prev_letter string
	Grades []byte
	Letters []string
}

type User struct {
	username string
	name string
	surname string
	grade string
	letter string
	password [32]byte
}


func (rg *User) save() error {
	utils.UserFilesMutex.Lock()
	defer utils.UserFilesMutex.Unlock()

	// adding user to user.txt (usertlist)
	f, err := os.OpenFile("authentication/users.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s;%s;%s;%s;%s;%x\n", rg.username, rg.name, rg.surname, rg.grade, rg.letter, rg.password))
	if err != nil {
		return err
	}

	// creating a new file for the user
	return os.WriteFile("authentication/users/"+rg.username+".txt", []byte(fmt.Sprintf("Username: %s\nName: %s\nSurname: %s\nGrade: %s\nLetter: %s\nPassword: %x\n", rg.username, rg.name, rg.surname, rg.grade, rg.letter, rg.password)), 0600)
}

func createPasswordHash(password string) [32]byte {
	return sha256.Sum256([]byte(password))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForAuthorizationCapability(w, r) == false {
		return
	}
	current_login := Login{}

	infoS := r.URL.Path[len("/login/"):]

	if len(infoS) > 0 {
		info := strings.NewReader(infoS)
		_, _ = fmt.Fscanf(info, "%s %s", 
			&current_login.Message, 
			&current_login.Prev_username)

		message := current_login.Message
		current_login.Message = ""
		for _, ch := range message {
			if string(ch) == "$" {
				current_login.Message += " "
			} else {
				current_login.Message += string(ch)
			}
		}
	}

	utils.RenderTemplate(w, "login", &current_login)
}

func AuthorizationLogHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForAuthorizationCapability(w, r) == false {
		return
	}
	message := "";
	failure := false
	if _, err := os.Stat("authentication/users/"+r.FormValue("username")+".txt"); err != nil || r.FormValue("username") == "" {
		message = "User$with$such$\"Username\"$does$not$exist!"
		failure = true
	} else {
		pw, err := readSpecificLineFromFile("authentication/users/"+r.FormValue("username")+".txt", 6)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		real_pw := pw[len("Password: "):]
		current_pw := fmt.Sprintf("%x", createPasswordHash(r.FormValue("password")))
		if real_pw != current_pw {
			message = "Wrong$\"password\""
			failure = true
		}
	}
	if failure {
		info := fmt.Sprintf("%s %s", message, r.FormValue("username"))
		http.Redirect(w, r, "/login/"+info, http.StatusFound)
		return
	}

	generateCookie(w, r.FormValue("username"), r.FormValue("password"))
	http.Redirect(w, r, "/", http.StatusFound)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForAuthorizationCapability(w, r) == false {
		return
	}
	current_registration := usual_registration

	infoS := r.URL.Path[len("/register/"):]

	if len(infoS) > 0 {
		info := strings.NewReader(infoS)
		_, _ = fmt.Fscanf(info, "%s %s %s %s %s %s", 
			&current_registration.Message, 
			&current_registration.Prev_username, 
			&current_registration.Prev_name, 
			&current_registration.Prev_surname,
			&current_registration.Prev_grade,
			&current_registration.Prev_letter)

		message := current_registration.Message
		current_registration.Message = ""
		for _, ch := range message {
			if string(ch) == "$" {
				current_registration.Message += " "
			} else {
				current_registration.Message += string(ch)
			}
		}
	}

	utils.RenderTemplate(w, "register", &current_registration)
}

func AuthorizationRegHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForAuthorizationCapability(w, r) == false {
		return
	}
	message := "";
	failure := false
	if _, err := os.Stat("authentication/users/"+r.FormValue("username")+".txt"); err == nil || r.FormValue("username") == "" {
		message = "\"Username\"$has$already$been$registered$:($$Try$to$choose$another$one"
		failure = true
	} else if r.FormValue("password") != r.FormValue("password_check") { // or any other conditions
		message = "\"Password\"$doesn't$match$with\"Password$check\""
		failure = true
	}
	if (failure) {
		info := fmt.Sprintf("%s %s %s %s %s %s", message, r.FormValue("username"), r.FormValue("name"), r.FormValue("surname"), r.FormValue("grade"), r.FormValue("letter"))
		http.Redirect(w, r, "/register/"+info, http.StatusFound)
		return
	}

	var newUser User
	newUser.username = r.FormValue("username")
	newUser.name = r.FormValue("name")
	newUser.surname = r.FormValue("surname")
	newUser.grade = r.FormValue("grade")
	newUser.letter = r.FormValue("letter")
	newUser.password = createPasswordHash(r.FormValue("password"))
	err := newUser.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	generateCookie(w, newUser.username, r.FormValue("password"))
	http.Redirect(w, r, "/", http.StatusFound)
}

var usual_registration = Registration{
	Grades: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	Letters: []string{"А", "Б", "В", "Г", "Д", "Е", "Ж", "З", "И", "К", "Л", "М", "Н", "О", "П", "Р", "С", "Т", "У", "Ф", "Х", "Ц", "Ч", "Ш", "Щ", "Э", "Ю", "Я"},
}
