package controllers

import (
	"encoding/json"
	"net/http"
	"signalone/pkg/models"
	"signalone/pkg/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IntegrationController struct {
	issuesCollection        *mongo.Collection
	usersCollection         *mongo.Collection
	analysisStoreCollection *mongo.Collection
}

func NewIntegrationController(issuesCollection *mongo.Collection,
	usersCollection *mongo.Collection,
	analysisStoreCollection *mongo.Collection) *IntegrationController {
	return &IntegrationController{
		issuesCollection:        issuesCollection,
		usersCollection:         usersCollection,
		analysisStoreCollection: analysisStoreCollection,
	}
}

// LogAnalysisTask godoc
// @Summary Perform log analysis and generate solutions.
// @Description Perform log analysis based on the provided logs and generate solutions.
// @Tags analysis
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param logAnalysisPayload body LogAnalysisPayload true "Log analysis payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Router /issues/analysis [put]
func (c *IntegrationController) LogAnalysisTask(ctx *gin.Context) {
	var user models.User
	var analysisResponse models.IssueAnalysis

	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		ctx.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	var logAnalysisPayload LogAnalysisPayload
	if err := ctx.ShouldBindJSON(&logAnalysisPayload); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := utils.GetUser(ctx, c.usersCollection, bson.M{"userId": logAnalysisPayload.UserId}, &user)
	if err != nil {
		return
	}

	issueId := uuid.New().String()
	go func() {
		var issueLogs = make([]string, 0)
		var issueLog Log
		var isNewIssue = true

		logAnalysisPayload.Logs = utils.AnonymizePII(logAnalysisPayload.Logs)
		logAnalysisPayload.Logs = utils.MaskSecrets(logAnalysisPayload.Logs)

		formattedAnalysisLogs := strings.Split(logAnalysisPayload.Logs, "\n")
		formattedAnalysisRelevantLogs := utils.FilterForRelevantLogs(formattedAnalysisLogs)
		if len(formattedAnalysisRelevantLogs) < 1 {
			formattedAnalysisRelevantLogs = formattedAnalysisLogs
		}

		qOpts := options.Find()
		qOpts.Projection = bson.M{"logs": 1}

		cursor, err := c.issuesCollection.Find(ctx, bson.M{
			"containerId": logAnalysisPayload.ContainerId,
			"isResolved":  false,
		}, qOpts)
		if err != nil {
			return
		}

		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			if err := cursor.Decode(&issueLog); err != nil {
				continue
			}
			issueLogs = append(issueLogs, issueLog.Logs...)
		}

		//Compare logs with previous logs and if they are similar enough, don't call the prediction agent
		if len(issueLogs) > 0 {
			isNewIssue = utils.CompareLogs(formattedAnalysisRelevantLogs, issueLogs)
			if !isNewIssue {
				return
			}
		}

		initialInsertResult, _ := c.issuesCollection.InsertOne(ctx, models.Issue{
			Id:                        issueId,
			UserId:                    "",
			ContainerName:             logAnalysisPayload.ContainerName,
			ContainerId:               logAnalysisPayload.ContainerId,
			Score:                     0,
			Severity:                  logAnalysisPayload.Severity,
			Title:                     analysisResponse.Title,
			TimeStamp:                 time.Now(),
			IsResolved:                false,
			Logs:                      formattedAnalysisLogs,
			RelevantLogs:              "",
			LogSummary:                "",
			PredictedSolutionsSummary: "",
			PredictedSolutionsSources: []string{},
			RegenerateCount:           0,
			ViewCount:                 0,
		})

		data := map[string]any{
			"logs":      strings.Join(formattedAnalysisRelevantLogs, "\n"),
			"isUserPro": user.IsPro,
		}
		jsonData, _ := json.Marshal(data)
		analysisResponse, err = utils.CallPredictionAgentService(jsonData)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !user.IsPro {
			c.analysisStoreCollection.InsertOne(ctx, models.SavedAnalysis{
				Logs:       logAnalysisPayload.Logs,
				LogSummary: analysisResponse.LogSummary,
			})
		}

		_, err = c.issuesCollection.UpdateOne(ctx,
			bson.M{
				"_id":         initialInsertResult.InsertedID,
				"containerId": logAnalysisPayload.ContainerId,
			},
			bson.M{"$set": bson.M{
				"userId":                         logAnalysisPayload.UserId,
				"title":                          analysisResponse.Title,
				"timestamp":                      time.Now(),
				"predictedSolutionsSummary":      analysisResponse.PredictedSolutions,
				"issuePredictedSolutionsSources": analysisResponse.Sources,
				"relevantLogs":                   analysisResponse.RelevantLogs,
				"logSummary":                     analysisResponse.LogSummary,
			},
			})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}()

	ctx.JSON(200, gin.H{
		"message": "Acknowledged",
		"issueId": issueId,
	})
}

// DeleteIssues godoc
// @Summary Delete issues based on the provided container name.
// @Description Delete issues based on the provided container name.
// @Tags issues
// @Accept json
// @Produce json
// @Param container query string true "Container name to delete issues from"
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /issues [delete]
func (c *IntegrationController) DeleteIssues(ctx *gin.Context) {
	containerId := ctx.Param("containerId")
	res, err := c.issuesCollection.DeleteMany(ctx, bson.M{"containerId": containerId})
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{
		"message": "Success",
		"count":   res.DeletedCount,
	})
}

func (c *IntegrationController) AddCodeAsContext(ctx *gin.Context) {
	var issue models.Issue
	var codeContext models.CodeContextRequest
	var user models.User
	id := ctx.Param("id")

	if err := ctx.ShouldBindJSON(&codeContext); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.issuesCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&issue); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if err := c.usersCollection.FindOne(ctx, bson.M{"userId": issue.UserId}).Decode(&user); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	codeSnippetRequest := models.CodeSnippetRequest{
		CurrentCodeSnippet: codeContext.Code,
		Logs:               issue.RelevantLogs,
		PredictedSolutions: issue.PredictedSolutionsSummary,
		LanguageId:         codeContext.Lang,
		IsUserPro:          user.IsPro,
	}

	jsonData, _ := json.Marshal(codeSnippetRequest)
	analysisResponse, err := utils.CallCodeGenAgentService(jsonData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err = c.issuesCollection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"codeContext": analysisResponse.Explanation,
		},
		}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if analysisResponse.Error != "" {
		ctx.JSON(http.StatusOK, gin.H{
			"message":     "Error",
			"error":       analysisResponse.Error,
			"explanation": analysisResponse.Explanation,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Success",
		"explanation": analysisResponse.Explanation,
	})
}
