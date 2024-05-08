package models

type CodeSnippetRequest struct {
	CurrentCodeSnippet string `json:"currentCodeSnippet"`
	Logs               string `json:"logs"`
	PredictedSolutions string `json:"predictedSolutions"`
	LanguageId         string `json:"languageId"`
	IsUserPro          bool   `json:"isUserPro"`
}

type CodeSnippetResponse struct {
	Code string `json:"code"`
}

type CodeContextRequest struct {
	Code string `json:"code"`
	Lang string `json:"lang"`
}
