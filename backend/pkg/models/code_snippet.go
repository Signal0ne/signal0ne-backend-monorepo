package models

type CodeSnippetRequest struct {
	CurrentCodeSnippet string `json:"currentCodeSnippet"`
	Logs               string `json:"logs"`
	PredictedSolutions string `json:"predictedSolutions"`
	LanguageId         string `json:"languageId"`
	IsUserPro          bool   `json:"isUserPro"`
}

type CodeSnippetResponse struct {
	Explanation string `json:"explanation"`
	Error       string `json:"error"`
}

type CodeContextRequest struct {
	Code string `json:"code"`
	Lang string `json:"lang"`
}
