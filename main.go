package main

import (
	"net/http"
	"log"
	"time"
	"tucklejudge/authentication"
	"tucklejudge/mainMenu"
	"tucklejudge/tester/testCreator"
	"tucklejudge/utils"
)

func main() {
	utils.Init()

	ticker := time.NewTicker(48*time.Hour)

	go func() {
		for {
			<-ticker.C
			utils.LoginCookieStorage.Clear()
		}
	}()

	http.HandleFunc("/login/", authentication.LoginHandler)
	http.HandleFunc("/authorize/login", authentication.AuthorizationLogHandler)

	http.HandleFunc("/register/", authentication.RegisterHandler)
	http.HandleFunc("/authorize/register", authentication.AuthorizationRegHandler)

	http.HandleFunc("/", mainMenu.MainPageHandler)

	http.HandleFunc("/test/createTest", testCreator.TestCreatorHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}