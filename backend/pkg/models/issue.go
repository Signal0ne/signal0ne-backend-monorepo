package models

import (
	"time"
)

type IssueRateRequest struct {
	Score *int32 `json:"score" binding:"required"` // it must be a pointer because if we get 0 then the required error arises
}

type IssueResolveRequest struct {
	IsResolved *bool `json:"isResolved" binding:"required"` // it must be a pointer because if we get 'false' then the required error arises
}

type IssueSolutionPredictionSolutionSource struct {
	Title string `json:"title" bson:"title"`
	Url   string `json:"url" bson:"url"`
}

type IssueSearchResult struct {
	Id            string    `json:"id" bson:"_id"`
	ContainerName string    `json:"containerName" bson:"containerName"`
	Title         string    `json:"title" bson:"title"`
	IsResolved    bool      `json:"isResolved" bson:"isResolved"`
	TimeStamp     time.Time `json:"timestamp" bson:"timestamp"`
	Severity      string    `json:"severity" bson:"severity"`
}

type Issue struct {
	Id                        string    `json:"id" bson:"_id"`
	UserId                    string    `json:"userId" bson:"userId"`
	ContainerName             string    `json:"containerName" bson:"containerName"`
	ContainerId               string    `json:"containerId" bson:"containerId"`
	Score                     int32     `json:"score" bson:"score" binding:"odeof=-1 0 1"`
	Severity                  string    `json:"severity" bson:"severity"`
	Logs                      []string  `json:"logs" bson:"logs"`
	RelevantLogs              string    `json:"relevantLogs" bson:"relevantLogs"`
	Title                     string    `json:"title" bson:"title"`
	IsResolved                bool      `json:"isResolved" bson:"isResolved"`
	TimeStamp                 time.Time `json:"timestamp" bson:"timestamp"`
	LogSummary                string    `json:"logSummary" bson:"logSummary"`
	PredictedSolutionsSummary string    `json:"predictedSolutionsSummary" bson:"predictedSolutionsSummary"`
	PredictedSolutionsSources []string  `json:"issuePredictedSolutionsSources" bson:"issuePredictedSolutionsSources"`
	RegenerateCount           int32     `json:"regenerateCount" bson:"regenerateCount"`
	ViewCount                 int32     `json:"viewCount" bson:"viewCount"`
	CopiedSourcesLinksCount   int32     `json:"copiedSourcesLinksCount" bson:"copiedSourcesLinksCount"`
}
