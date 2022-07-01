package mainMenu

import (
	"net/http"
	"tucklejudge/utils"
)

type MenuUI struct {
	UserID string
	Username string
	Teacher bool
	Tests []TestUI
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

	tests := make([]TestUI, len(user.Tests))
	for i, id := range user.Tests {
		tests[i].TestID = id
		t, err := utils.GetTestByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		tests[i].TestName = t.Name
	}

	menu := &MenuUI{
		UserID: user.ID,
		Username: user.Username,
		Teacher: user.Teacher,
		Tests: tests,
	}
	utils.RenderTemplate(w, "mainMenu", menu)
}

