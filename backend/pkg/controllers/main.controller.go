package controllers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/smtp"
	_ "signalone/docs"
	"signalone/pkg/models"
	"signalone/pkg/utils"

	"github.com/gin-gonic/gin"
	e "github.com/jordan-wright/email"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
