package testCreator

import (
	"net/http"
	"tucklejudge/utils"
	"strconv"
	"fmt"
)

const NUMBER_OF_QUESTIONS = 30

func TestEditHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	if !utils.CheckForTeacher(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	testID := r.URL.Path[len("/test/editTest/"):]
	test, err := utils.GetTestByID(testID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	test.NumberOfQuestionsForTemplate = len(test.Questions)

	for i := len(test.Questions); i < NUMBER_OF_QUESTIONS; i++ {
		test.Questions = append(test.Questions, utils.Question{
			Answer: "",
			Points: 1,
		})
	}
	for i := range test.Questions {
		test.Questions[i].IndexForTemplate = i+1
	}

	utils.RenderTemplate(w, "testEditor", test)
}

func TestCreatorHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	if !utils.CheckForTeacher(r) {
		http.Redirect(w, r, "/", http.StatusFound)
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
	if !utils.CheckForTeacher(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	var test utils.Test
	test.Name = r.FormValue("testName")
	n, _ := strconv.Atoi(r.FormValue("numberOfQuestions"))
	test.Questions = make([]utils.Question, n)
	for i, _ := range test.Questions {
		test.Questions[i].Answer = r.FormValue(fmt.Sprintf("answer%d", i+1))
		test.Questions[i].Points, _ = strconv.Atoi(r.FormValue(fmt.Sprintf("points%d", i+1)))
	}
	for i, _ := range test.PointsToMark {
		test.PointsToMark[i], _ = strconv.Atoi(r.FormValue(fmt.Sprintf("pointsTo%d", i+2)))
	}
	err := test.CreateIDAndSave()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, err := utils.GetUsername(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	utils.AddTestToUsersList(username, test.ID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func SavingProcessHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	if !utils.CheckForTeacher(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	
	var test utils.Test
	test.ID = r.URL.Path[len("/test/saveTest/process/"):]
	test.Name = r.FormValue("testName")
	n, _ := strconv.Atoi(r.FormValue("numberOfQuestions"))
	test.Questions = make([]utils.Question, n)
	for i, _ := range test.Questions {
		test.Questions[i].Answer = r.FormValue(fmt.Sprintf("answer%d", i+1))
		test.Questions[i].Points, _ = strconv.Atoi(r.FormValue(fmt.Sprintf("points%d", i+1)))
	}
	for i, _ := range test.PointsToMark {
		test.PointsToMark[i], _ = strconv.Atoi(r.FormValue(fmt.Sprintf("pointsTo%d", i+2)))
	}


	// saving test to file
	err := utils.SaveTestToFile(&test)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func TestDeletionHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	username, err := utils.GetUsername(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	testID := r.URL.Path[len("/test/deleteTest/process/"):]
	
	err = utils.DeleteTestFromUsersList(username, testID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

