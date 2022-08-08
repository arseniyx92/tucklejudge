package testChecker

import (
	"net/http"
	"tucklejudge/fieldsRecognition"
	"tucklejudge/utils"
	"strings"
	"fmt"
)

func TestCheckHandler(w http.ResponseWriter, r *http.Request) {
	if utils.CheckForValidStandardAccess(w, r) == false {
		return
	}
	fileName, err := utils.SaveFormFileToSrc(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// getting file extension
	strsForGettingCorrectExtension := strings.Split(fileName, ".")
	ext := strsForGettingCorrectExtension[1]

	// applying fieldsRecognition on particular extension
	var inputInfo [][]string
	if ext == "pdf" {
		inputInfo = fieldsRecognition.BringTestResultsFromPDFs("src/"+fileName)
	} else {
		inputInfo := make([][]string, 1)
		inputInfo[0] = fieldsRecognition.BringTestResultsFromPhoto("src/"+fileName)
	}
	fmt.Println(inputInfo)

	// TODO
	// sanding each inputInfo to checker (input -> protocol)
}