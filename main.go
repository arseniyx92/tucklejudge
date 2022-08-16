package main

import (
	"net/http"
	"log"
	"time"
	"tucklejudge/authentication"
	"tucklejudge/mainMenu"
	"tucklejudge/tester/testCreator"
	"tucklejudge/tester/testViewer"
	"tucklejudge/tester/testChecker"
	"tucklejudge/utils"
)

func main() {
	utils.Init()

	cookieTicker := time.NewTicker(168*time.Hour)
	newVerificationCodeTicker := time.NewTicker(48*time.Hour)

	go func() {
		for {
			<-cookieTicker.C
			utils.LoginCookieStorage.Clear()
		}
	}()
	go func() {
		for {
			<-newVerificationCodeTicker.C
			utils.ChangeVerificationCode()
		}
	}()

	http.HandleFunc("/login/", authentication.LoginHandler)
	http.HandleFunc("/authorize/login", authentication.AuthorizationLogHandler)

	http.HandleFunc("/register/", authentication.RegisterHandler)
	http.HandleFunc("/authorize/register", authentication.AuthorizationRegHandler)

	http.HandleFunc("/", mainMenu.MainPageHandler)

	http.HandleFunc("/test/createTest", testCreator.TestCreatorHandler)
	http.HandleFunc("/test/editTest/", testCreator.TestEditHandler)
	http.HandleFunc("/test/createTest/process", testCreator.CreationProcessHandler)
	http.HandleFunc("/test/saveTest/process/", testCreator.SavingProcessHandler)
	http.HandleFunc("/test/deleteTest/process/", testCreator.TestDeletionHandler)

	http.HandleFunc("/test/view/", testViewer.TestViewHandler)
	http.HandleFunc("/test/teacherView/", testViewer.TeacherTestViewHandler)

	http.HandleFunc("/test/checkTest", testChecker.TestCheckHandler)

	// http.HandleFunc("lesson/changeMarks/", lessonEditor.ChangeMarksHandler)
	// http.HandleFunc("/test/deployToElectronicMarkBook/", lessonEditor.DeployToElectronicMarkBookHandler)

	http.HandleFunc("/clearEverything__WARNING", utils.ClearAllData)

	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("./src"))))
	http.Handle("/favicon.ico", http.NotFoundHandler()) // TODO
	log.Fatal(http.ListenAndServe(":8080", nil))
}