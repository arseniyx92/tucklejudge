package testCreator

import (
	"net/http"
	"tucklejudge/utils"
)

const NUMBER_OF_QUESTIONS = 20

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