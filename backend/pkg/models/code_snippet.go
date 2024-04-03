package models

type CodeSnippetRequest struct {
	CurrentCodeSnippet string `json:"currentCodeSnippet"`
	Logs               string `json:"logs"`
	PredictedSolutions string `json:"predictedSolutions"`
}

type CodeSnippetResponse struct {
	NewCodeSnippet string `json:"newCodeSnippet"`
}

type CodeContextRequest struct {
	Code string `json:"code"`
}
