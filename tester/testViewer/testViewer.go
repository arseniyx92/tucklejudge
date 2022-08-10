package testViewer

import (
	"net/http"
	"tucklejudge/utils"
	"strings"
	"strconv"
	"fmt"
)

func TestViewHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	info := strings.Split(r.URL.Path[len("/test/view/"):], "$")
	givenTestID := info[0]
	givenUsername := info[1]
	c, _ := r.Cookie("user_info")
	utils.UserFilesMutex.Lock()
	username, _ := utils.LoginCookieStorage.ReturnNodeValue(c.Value)
	utils.UserFilesMutex.Unlock()
	if username != givenUsername {
		user, err := utils.GetAccauntInfo(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !user.Teacher {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}
	// receiving test from system files
	str, err := utils.GetTestUsersResultByID(givenTestID, givenUsername)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	strs := strings.Split(str, "\n")
	testInfo := &utils.PersonalTest{
		UserName: givenUsername,
		TestName: strs[0][len("TestName: "):],
		Mark: strs[1][len("Mark: "):],
		InputImageName: strs[2][len("Input image name: "):],
		ProcessedImageName: strs[3][len("Processed image name: "):],
		//Questions: []PersonalQuestion
		PointsSum: strs[len(strs)-6][len("Points sum: "):],
		PointsToMark: [3]string{strs[len(strs)-4], strs[len(strs)-3], strs[len(strs)-2]},
	}
	n, _ := strconv.Atoi(strs[4][len("Questions ("):len(strs[4])-1])
	testInfo.Questions = make([]utils.PersonalQuestion, n)
	for i := 0; i < n; i++ {
		s := strings.Split(strs[5+i][len(fmt.Sprintf("%d) ", i+1)):], " ")
		testInfo.Questions[i].Index = string(i+1+'0')
		testInfo.Questions[i].UserAnswer = s[0]
		testInfo.Questions[i].CorrectAnswer = s[1]
		testInfo.Questions[i].Points = s[2]
	}
	utils.RenderTemplate(w, "testViewer", testInfo)
}

func TeacherTestViewHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	filename := r.URL.Path[len("/test/teacherView/"):]
	testingInfo, err := utils.LoadShortResultsFromFile(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.RenderTemplate(w, "testChecker", testingInfo)
}

