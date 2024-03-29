package routers

import (
	"signalone/pkg/controllers"
	middlewares "signalone/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type MainRouter struct {
	mainController            *controllers.MainController
	userAuthController        *controllers.UserAuthController
	userController            *controllers.UserController
	userIssuesController      *controllers.UserIssuesController
	integrationController     *controllers.IntegrationController
	integrationAuthController *controllers.IntegrationAuthController
}

func NewMainRouter(mainController *controllers.MainController,
	userAuthController *controllers.UserAuthController,
	userController *controllers.UserController,
	userIssuesController *controllers.UserIssuesController,
	integrationController *controllers.IntegrationController,
	integrationAuthController *controllers.IntegrationAuthController) *MainRouter {
	return &MainRouter{
		mainController:            mainController,
		userAuthController:        userAuthController,
		userController:            userController,
		userIssuesController:      userIssuesController,
		integrationController:     integrationController,
		integrationAuthController: integrationAuthController,
	}
}

func (mr *MainRouter) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/contact", mr.mainController.ContactHandler)
	rg.POST("/waitlist", mr.mainController.WaitlistHandler)

	authorizationRouterGroup := rg.Group("/auth")
	authorizationRouterGroup.POST("/email-confirmation", mr.userAuthController.VerifyEmail)
	authorizationRouterGroup.POST("/email-confirmation-link-resend", mr.userAuthController.ResendConfirmationEmail)
	authorizationRouterGroup.POST("/login", mr.userAuthController.LoginHandler)
	authorizationRouterGroup.POST("/login-with-github", mr.userAuthController.LoginWithGithubHandler)
	authorizationRouterGroup.POST("/login-with-google", mr.userAuthController.LoginWithGoogleHandler)
	authorizationRouterGroup.POST("/register", mr.userAuthController.RegisterHandler)
	authorizationRouterGroup.POST("/token/refresh", mr.userAuthController.RefreshTokenHandler)

	userRouterGroup := rg.Group("/user", middlewares.CheckAuthorization)
	{
		userRouterGroup.POST("/agent/authenticate", mr.integrationAuthController.AuthenticateAgent)
		userRouterGroup.GET("/containers", mr.userIssuesController.GetContainers)
		userRouterGroup.GET("/issues", mr.userIssuesController.IssuesSearch)
		userRouterGroup.GET("/issues/:id", mr.userIssuesController.GetIssue)
		userRouterGroup.PUT("/issues/:id/regenerate", mr.userIssuesController.RegenerateSolution)
		userRouterGroup.PUT("/issues/:id/resolve", mr.userIssuesController.ResolveIssue)
		userRouterGroup.PUT("/issues/:id/score", mr.userIssuesController.RateIssue)
		userRouterGroup.POST("/issues/report", mr.userIssuesController.ReportIssueAnalysis)
		userRouterGroup.GET("/last-activity", mr.userController.LastActivityHandler)
		userRouterGroup.GET("/settings", func(c *gin.Context) {})
		userRouterGroup.POST("/settings", func(c *gin.Context) {})
	}

	agentRouterGroup := rg.Group("/agent", mr.integrationAuthController.CheckAgentAuthorization)
	{
		agentRouterGroup.DELETE("/issues/:containerId", mr.integrationController.DeleteIssues)
		agentRouterGroup.PUT("/issues/analysis", mr.integrationController.LogAnalysisTask)
	}
}
