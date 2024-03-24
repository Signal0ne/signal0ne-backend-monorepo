package models

type IssueAnalysis struct {
	Title              string   `json:"title"`
	LogSummary         string   `json:"logsummary"`
	PredictedSolutions string   `json:"predictedSolutions"`
	Sources            []string `json:"sources"`
}

type IssueAnalysisReportRequest struct {
	IssueId string `json:"issueId" bson:"_id"`
	Reason  string `json:"reason" bson:"reason"`
	Delete  bool   `json:"delete" bson:"delete"`
}

type IssueAnalysisReportResponse struct {
	Acknowledged bool `json:"acknowledged"`
	Deleted      bool `json:"deleted"`
}
