package testChecker

import (
	"net/http"
	"tucklejudge/fieldsRecognition"
	"tucklejudge/utils"
	"strings"
	"fmt"
)

func createProtocol(input []string, inputPictureName, processedPictureName string) (*utils.PersonalResult, error) {
	// userID := input[0]
	userID := "0001"
	testID := input[1]
	test, err := utils.GetTestByID(testID)
	if err != nil {
		return nil, err
	}
	username, err := utils.GetUsernameByID(userID)
	if err != nil {
		return nil, err
	}
	user, err := utils.GetAccauntInfo(username)
	if err != nil {
		return nil, err
	}
	results := &utils.PersonalTest {
		UserName: username,
		TestName: test.Name,
		InputImageName: inputPictureName,
		ProcessedImageName: processedPictureName,
		Questions: make([]utils.PersonalQuestion, len(test.Questions)),
	}
	numericPointsSum := 0
	for i, q := range test.Questions {
		ind := 2
		if i < 15 {
			ind += 2*i
		} else {
			ind += 2*(i-15)+1
		}
		userAnswer := input[i+2]
		results.Questions[i].Index = fmt.Sprint(i+1)
		results.Questions[i].UserAnswer = userAnswer[:len(q.Answer)]
		results.Questions[i].CorrectAnswer = q.Answer
		results.Questions[i].Points = "0"
		if (q.Answer == userAnswer[:len(q.Answer)]) {
			results.Questions[i].Points = fmt.Sprint(q.Points)
			numericPointsSum += q.Points
		}
	}
	results.PointsSum = fmt.Sprint(numericPointsSum)
	mark := 5
	for i, points := range test.PointsToMark {
		results.PointsToMark[i] = fmt.Sprint(points)
		if (numericPointsSum < points && mark == 5) {
			mark = i+2
		}
	}
	results.Mark = fmt.Sprint(mark)

	err = utils.CreateTestResultFile(testID+"$"+username, results)
	if err != nil {
		return nil, err
	}
	err = utils.AddTestToUsersList(username, testID)
	if err != nil {
		return nil, err
	}

	short_result := &utils.PersonalResult {
		TestID: testID,
		Username: username,
		FullName: user.Surname + " " + user.Name,
		Mark: results.Mark,
	}
	return short_result, nil
}

func TestCheckHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	fileName, err := utils.SaveFormFileToSrc(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// getting file extension
	strsForGettingCorrectExtension := strings.Split(fileName, ".")
	ext := strsForGettingCorrectExtension[1]

	// applying fieldsRecognition on particular extension
	var inputInfo [][]string
	var imagesNames [][]string
	if ext == "pdf" {
		// TODO should be inputInfo, imagesNames
		inputInfo = fieldsRecognition.BringTestResultsFromPDFs("src/"+fileName)
	} else {
		// TODO should be inputInfo[0], imagesNames[0]
		inputInfo := make([][]string, 1)
		inputInfo[0] = fieldsRecognition.BringTestResultsFromPhoto("src/"+fileName)
	}

	testingInfo := &utils.ShortTestResultsInfo {
		Results: make([]utils.PersonalResult, len(inputInfo)),
	}
	for i, str := range inputInfo {
		imagesNames = append(imagesNames, []string{fileName, fileName}) // TODO make redundant
		res, err := createProtocol(str, imagesNames[i][0], imagesNames[i][1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		testingInfo.Results[i] = *res
	}
	string_id, err := utils.GetCurrentlyFreeID("tester/teacherTestResults", 6)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = utils.SaveShortResultsInfoToFile(string_id, testingInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	username, err := utils.GetUsername(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = utils.AddTestToUsersList(username, string_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.RenderTemplate(w, "testChecker", testingInfo)
}
