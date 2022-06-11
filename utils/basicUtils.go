package utils

import (
	"net/http"
	"html/template"
	"tucklejudge/utils/splayMap"
	"sync"
)

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