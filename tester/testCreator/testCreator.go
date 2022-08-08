package testCreator

import (
	"net/http"
	"tucklejudge/utils"
	"strconv"
	"fmt"
)

const NUMBER_OF_QUESTIONS = 30

func TestCreatorHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	questions_iterator := make([]int, NUMBER_OF_QUESTIONS)
	for i := range questions_iterator {
		questions_iterator[i] = i+1
	}
	test := struct {
		Questions []int
	}{
		Questions: questions_iterator,
	}
	utils.RenderTemplate(w, "testCreator", test)
}

func CreationProcessHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	var test utils.Test
	test.Name = r.FormValue("testName")
	n, _ := strconv.Atoi(r.FormValue("numberOfQuestions"))
	test.Questions = make([]utils.Question, n)
	for i, _ := range test.Questions {
		test.Questions[i].Answer = r.FormValue(fmt.Sprintf("answer%d", i+1))
		test.Questions[i].Points, _ = strconv.Atoi(r.FormValue(fmt.Sprintf("points%d", i+1)))
		test.Questions[i].Punishment, _ = strconv.Atoi(r.FormValue(fmt.Sprintf("punishment%d", i+1)))
	}
	for i, _ := range test.PointsToMark {
		test.PointsToMark[i], _ = strconv.Atoi(r.FormValue(fmt.Sprintf("pointsTo%d", i+2)))
	}
	err := test.CreateIDAndSave()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	username, err := utils.GetUsername(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	utils.AddTestToTeachersList(username, test.ID)
	http.Redirect(w, r, "/", http.StatusFound)
}


