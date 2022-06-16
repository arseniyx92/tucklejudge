package mainMenu

import (
	"net/http"
	"tucklejudge/utils"
)

type MenuUI struct {
	Username string
	Teacher bool
}

func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	c, _ := r.Cookie("user_info")
	username, _ := utils.LoginCookieStorage.ReturnNodeValue(c.Value)
	user, err := utils.GetAccauntInfo(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	menu := &MenuUI{
		Username: user.Username,
		Teacher: user.Teacher,
	}
	utils.RenderTemplate(w, "mainMenu", menu)
}