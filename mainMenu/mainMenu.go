package mainMenu

import (
	"net/http"
	"tucklejudge/utils"
	"fmt"
)

type MenuUI struct {
	UserID string
	Username string
	Teacher bool
	Tests []TestUI
	Classes []TestUI
}

type TestUI struct {
	TestID string
	TestName string
}

func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	c, _ := r.Cookie("user_info")
	utils.UserFilesMutex.Lock()
	username, _ := utils.LoginCookieStorage.ReturnNodeValue(c.Value)
	utils.UserFilesMutex.Unlock()
	user, err := utils.GetAccauntInfo(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var tests []TestUI
	var classes []TestUI
	if user.Teacher {
		for i, id := range user.Tests {
			if len(id) == 4 {
				tests = append(tests, TestUI{})
				tests[len(tests)-1].TestID = id
				t, err := utils.GetTestByID(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				tests[len(tests)-1].TestName = t.Name
			} else {
				classes = append(classes, TestUI{})
				classes[len(classes)-1].TestID = id
				classes[len(classes)-1].TestName = fmt.Sprintf("Check №%d", i+1)
			}
		}
	} else {
		tests = make([]TestUI, len(user.Tests))
		for i, id := range user.Tests {
			tests[i].TestID = id
			t, err := utils.GetTestByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			tests[i].TestName = t.Name
		}
	}

	for i := 0; i < len(tests)/2; i++ {
		tests[i], tests[len(tests)-i-1] = tests[len(tests)-i-1], tests[i]
	}
	for i := 0; i < len(classes)/2; i++ {
		classes[i], classes[len(classes)-i-1] = classes[len(classes)-i-1], classes[i]
	}

	menu := &MenuUI{
		UserID: user.ID,
		Username: user.Username,
		Teacher: user.Teacher,
		Tests: tests,
		Classes: classes,
	}
	utils.RenderTemplate(w, "mainMenu", menu)
}

