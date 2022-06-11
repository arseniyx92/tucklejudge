package main

import (
	"net/http"
	"log"
	"time"
	"tucklejudge/authentication"
	"tucklejudge/mainMenu"
)

func main() {
	ticker := time.NewTicker(10*time.Second)

	go func() {
		for {
			<-ticker.C
			authentication.CookieStorage.Clear()
		}
	}()

	http.HandleFunc("/login/", authentication.LoginHandler)
	http.HandleFunc("/authorize/login", authentication.AuthorizationLogHandler)

	http.HandleFunc("/register/", authentication.RegisterHandler)
	http.HandleFunc("/authorize/register", authentication.AuthorizationRegHandler)

	http.HandleFunc("/", mainMenu.mainPageHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
} 
// git branch --all
// git push origin --delete master
// git push -u main
