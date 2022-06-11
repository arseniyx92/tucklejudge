package mainMenu

import (
	"net/http"
	"tucklejudge/utils"
)

type MenuUI struct {
	Username string
}

func MainPageHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	c, _ := r.Cookie("user_info")
	user, _ := utils.LoginCookieStorage.ReturnNodeValue(c.Value)

	menu := &MenuUI{
		Username: user,
	}
	utils.RenderTemplate(w, "mainMenu", menu)
}