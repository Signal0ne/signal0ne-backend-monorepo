package controllers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	_ "signalone/docs"
	"signalone/pkg/models"
	"signalone/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	e "github.com/jordan-wright/email"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogAnalysisPayload struct {
	UserId        string `json:"userId"`
	ContainerName string `json:"containerName"`
	ContainerId   string `json:"containerId"`
	Severity      string `json:"severity"`
	Logs          string `json:"logs"`
}

type Log struct {
	Logs []string `bson:"logs"`
}

type EmailClientConfig struct {
	AuthData    smtp.Auth
	HostAddress string
	From        string
	TlsConfig   *tls.Config
}

type MainController struct {
	issuesCollection        *mongo.Collection
	usersCollection         *mongo.Collection
	analysisStoreCollection *mongo.Collection
	waitlistCollection      *mongo.Collection
	emailClientData         EmailClientConfig
}

func NewMainController(issuesCollection *mongo.Collection,
	usersCollection *mongo.Collection,
	analysisStoreCollection *mongo.Collection,
	waitlistCollection *mongo.Collection,
	emailClientData EmailClientConfig) *MainController {
	return &MainController{
		issuesCollection:        issuesCollection,
		usersCollection:         usersCollection,
		analysisStoreCollection: analysisStoreCollection,
		waitlistCollection:      waitlistCollection,
		emailClientData:         emailClientData,
	}
}

func (c *MainController) ContactHandler(ctx *gin.Context) {
	var emailReqBody models.Email

	err := ctx.ShouldBindJSON(&emailReqBody)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emailObj := e.NewEmail()
	emailObj.From = c.emailClientData.From
	emailObj.To = []string{"contact@signaloneai.com"}
	emailObj.Subject = fmt.Sprintf("[CONTACT] %s", emailReqBody.MessageTitle)
	emailObj.Text = []byte(fmt.Sprintf("From: %s \nMessage: %s", emailReqBody.Email, emailReqBody.MessageContent))
	err = emailObj.SendWithStartTLS(c.emailClientData.HostAddress, c.emailClientData.AuthData, c.emailClientData.TlsConfig)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the mail"})
		return
	}

	resEmailObj := e.NewEmail()
	resEmailObj.From = c.emailClientData.From
	resEmailObj.To = []string{emailReqBody.Email}
	resEmailObj.Subject = "Thank you for contacting us"
	resEmailObj.HTML = []byte(`<img alt="Signal0ne" title="Signal0ne Logo" width="196px" height="57px" src="https://signaloneai.com/online-assets/Signal0ne.jpg"
	style="margin-top: 40px;"><h1 style="color: black">Hello,</h1> <p style="color: black">Thank you for contacting us.</p> <p style="color: black">We will get back to you as soon as possible.</p><br><p style="color: black; margin-bottom: 0; margin-top: 4px;">Best regards,</p><p style="color: black; font-family: consolas; font-size: 15px; font-weight: bold; margin-top: 6px;";>Signal0ne Team</p>`)
	err = resEmailObj.SendWithStartTLS(c.emailClientData.HostAddress, c.emailClientData.AuthData, c.emailClientData.TlsConfig)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the mail"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Email has been sent successfully",
	})
}

func (c *MainController) WaitlistHandler(ctx *gin.Context) {
	var waitlistEntry models.WaitlistEntry

	err := ctx.ShouldBindJSON(&waitlistEntry)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	waitlistQueryResult := c.waitlistCollection.FindOne(ctx, bson.M{"email": waitlistEntry.Email})
	if waitlistQueryResult.Err() != mongo.ErrNoDocuments {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists in the waitlist"})
		return
	}

	emailObj := e.NewEmail()
	emailObj.From = c.emailClientData.From
	emailObj.To = []string{waitlistEntry.Email}
	emailObj.Subject = "Thank you for joining the waitlist!"
	emailObj.HTML = []byte(utils.WaitlistEntryConfirmationEmail)
	err = emailObj.SendWithStartTLS(c.emailClientData.HostAddress, c.emailClientData.AuthData, c.emailClientData.TlsConfig)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "There was an error sending the mail"})
		return
	}

	_, err = c.waitlistCollection.InsertOne(ctx, waitlistEntry)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

}

// IssuesSearch godoc
// @Summary Search for issues based on specified criteria.
// @Description Search for issues based on specified criteria.
// @Tags issues
// @Accept json
// @Produce json
// @Param offset query int false "Offset for paginated results"
// @Param limit query int false "Maximum number of results per page (default: 30, max: 100)"
// @Param searchString query string false "Search string for filtering issues"
// @Param container query string false "Filter by container name"
// @Param issueSeverity query string false "Filter by issue severity"
// @Param issueType query string false "Filter by issue type"
// @Param startTimestamp query string false "Filter issues starting from this timestamp (RFC3339 format)"
// @Param endTimestamp query string false "Filter issues until this timestamp (RFC3339 format)"
// @Param isResolved query bool false "Filter resolved or unresolved issues"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /issues [get]
func (c *MainController) IssuesSearch(ctx *gin.Context) {
	var max int64
	issues := make([]models.IssueSearchResult, 0)

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	container := ctx.Query("container")
	endTimestampQuery := ctx.Query("endTimestamp")
	issueSeverity := ctx.Query("issueSeverity")
	issueType := ctx.Query("issueType")
	limitQuery := ctx.Query("limit")
	offsetQuery := ctx.Query("offset")
	startTimestampQuery := ctx.Query("startTimestamp")
	_ = ctx.Query("searchString")

	isResolved, err := strconv.ParseBool(ctx.Query("isResolved"))
	if err != nil {
		isResolved = true
	}

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil || offsetQuery == "" {
		offset = 0
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit > 100 || limitQuery == "" {
		limit = 30
	}

	startTimestamp, err := time.Parse(time.RFC3339, startTimestampQuery)
	if err != nil {
		fmt.Print("Error: ", err)
		startTimestamp = time.Time{}.UTC()
	}

	endTimestamp, err := time.Parse(time.RFC3339, endTimestampQuery)
	if err != nil || endTimestampQuery == "" {
		fmt.Print("Error: ", err)
		endTimestamp = time.Now().UTC()
	}

	qOpts := options.Find()
	qOpts.SetLimit(int64(limit))
	qOpts.SetSkip(int64(offset))
	qOpts.SetSort(bson.M{"timestamp": -1})
	qOpts.SetProjection(bson.M{
		"_id":           1,
		"containerName": 1,
		"severity":      1,
		"title":         1,
		"isResolved":    1,
		"timestamp":     1,
	})

	filter := bson.M{
		"userId": userId,
		"timestamp": bson.M{
			"$gte": startTimestamp.UTC(),
			"$lte": endTimestamp.UTC(),
		},
	}

	if container != "" {
		filter["containerName"] = container
	}

	if issueSeverity != "" {
		filter["severity"] = issueSeverity
	}

	if issueType != "" {
		filter["type"] = issueType
	}

	if !isResolved {
		filter["isResolved"] = isResolved
	}

	cursor, err := c.issuesCollection.Find(ctx, filter, qOpts)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var issue models.IssueSearchResult

		if err := cursor.Decode(&issue); err != nil {
			continue
		}

		issues = append(issues, issue)
	}

	max, _ = c.issuesCollection.CountDocuments(ctx, filter)

	ctx.JSON(http.StatusOK, gin.H{
		"issues": issues,
		"max":    max,
	})
}

// GetIssue godoc
// @Summary Get information about a specific issue.
// @Description Get information about a specific issue by providing its ID.
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "ID of the issue"
// @Success 200 {object} models.Issue
// @Failure 404 {object} map[string]any
// @Router /issues/{id} [get]
func (c *MainController) GetIssue(ctx *gin.Context) {
	var issue models.Issue
	id := ctx.Param("id")

	if err := c.issuesCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&issue); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	ctx.JSON(http.StatusOK, issue)
}

func (c *MainController) RateIssue(ctx *gin.Context) {
	var issue models.Issue
	var issueRateReq models.IssueRateRequest
	var user models.User

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = ctx.ShouldBindJSON(&issueRateReq)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if *issueRateReq.Score != -1 && *issueRateReq.Score != 0 && *issueRateReq.Score != 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Score must be one of: -1, 0, 1"})
		return
	}

	userResult := c.usersCollection.FindOne(ctx, bson.M{"userId": userId})

	err = userResult.Decode(&user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := ctx.Param("id")

	issueConditions := bson.M{
		"_id":    id,
		"userId": userId,
	}

	filter := utils.GenerateFilter(issueConditions, "$and")
	issueResult := c.issuesCollection.FindOne(ctx, filter)

	err = issueResult.Decode(&issue)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var currentIssueScore = issue.Score

	if currentIssueScore == *issueRateReq.Score {
		ctx.JSON(http.StatusOK, gin.H{"message": "Issue already rated with the same score"})
		return
	}

	updatedIssueResult, err := c.issuesCollection.UpdateOne(ctx,
		filter,
		bson.M{
			"$set": bson.M{
				"score": issueRateReq.Score,
			},
		})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updatedIssueResult.MatchedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Issue cannot be found"})
		return
	}

	counter := user.Counter
	counter = utils.CalculateNewCounter(currentIssueScore, *issueRateReq.Score, counter)

	updatedUserResult, err := c.usersCollection.UpdateOne(ctx,
		bson.M{"userId": userId},
		bson.M{
			"$set": bson.M{
				"counter": counter,
			},
		})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updatedUserResult.MatchedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User cannot be found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}

func (c *MainController) RegenerateSolution(ctx *gin.Context) {
	var analysisResponse models.IssueAnalysis
	var issue models.Issue
	var user models.User

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id := ctx.Param("id")
	issueResult := c.issuesCollection.FindOne(ctx, bson.M{"_id": id, "userId": userId})

	err = issueResult.Decode(&issue)
	if err != nil && err.Error() == mongo.ErrNoDocuments.Error() {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userResult := c.usersCollection.FindOne(ctx, bson.M{"userId": userId})

	err = userResult.Decode(&user)
	if err != nil && err.Error() == mongo.ErrNoDocuments.Error() {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var formattedAnalysisRelevantLogs = utils.FilterForRelevantLogs(issue.Logs)
	data := map[string]string{"logs": strings.Join(formattedAnalysisRelevantLogs, "\n")}
	jsonData, _ := json.Marshal(data)

	analysisResponse, err = utils.CallPredictionAgentService(jsonData)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	if !user.IsPro {
		c.analysisStoreCollection.InsertOne(ctx, models.SavedAnalysis{
			Logs:       strings.Join(issue.Logs, "\n"),
			LogSummary: analysisResponse.LogSummary,
		})
	}
	_, err = c.issuesCollection.UpdateOne(ctx, bson.M{"_id": id, "userId": userId}, bson.M{"$set": bson.M{
		"title":                          analysisResponse.Title,
		"timestamp":                      time.Now(),
		"predictedSolutionsSummary":      analysisResponse.PredictedSolutions,
		"issuePredictedSolutionsSources": analysisResponse.Sources,
		"logSummary":                     analysisResponse.LogSummary,
	}})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	issueResult = c.issuesCollection.FindOne(ctx, bson.M{"_id": id, "userId": userId})

	err = issueResult.Decode(&issue)
	if err != nil && err.Error() == mongo.ErrNoDocuments.Error() {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	ctx.JSON(http.StatusOK, issue)
}

// ResolveIssue godoc
// @Summary Mark issue as resolved/unresolved.
// @Description Resolve an issue by providing its ID and resolve state of the issue.
// @Tags issues
// @Accept json
// @Produce json
// @Param id path string true "ID of the issue to be resolved"
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /issues/{id}/resolve [put]
// @RequestBody application/json isResolved boolean
func (c *MainController) ResolveIssue(ctx *gin.Context) {
	var requestData models.IssueResolveRequest

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := ctx.Param("id")

	issueResult, err := c.issuesCollection.UpdateOne(ctx, bson.M{"_id": id, "userId": userId}, bson.M{"$set": bson.M{"isResolved": *requestData.IsResolved}})
	if issueResult.MatchedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Issue not found"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})
}

// GetContainers godoc
// @Summary Get a list of containers based on the provided user ID.
// @Description Get a list of containers based on the provided user ID.
// @Tags containers
// @Accept json
// @Produce json
// @Param userId query string true "User ID to filter containers"
// @Success 200 {array} string
// @Failure 500 {object} map[string]any
// @Router /containers [get]
func (c *MainController) GetContainers(ctx *gin.Context) {
	containers := make([]string, 0)

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	results, err := c.issuesCollection.Distinct(ctx, "containerName", bson.M{"userId": userId})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, result := range results {
		if container, ok := result.(string); ok {
			containers = append(containers, container)
		}
	}
	ctx.JSON(http.StatusOK, containers)
}
