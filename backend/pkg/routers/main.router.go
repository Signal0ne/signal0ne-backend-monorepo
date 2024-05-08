package routers

import (
	"signalone/pkg/controllers"
	middlewares "signalone/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type MainRouter struct {
	mainController            *controllers.MainController
	paymentsController        *controllers.PaymentController
	userAuthController        *controllers.UserAuthController
	userController            *controllers.UserController
	userIssuesController      *controllers.UserIssuesController
	integrationController     *controllers.IntegrationController
	integrationAuthController *controllers.IntegrationAuthController
}

func NewMainRouter(mainController *controllers.MainController,
	paymentsController *controllers.PaymentController,
	userAuthController *controllers.UserAuthController,
	userController *controllers.UserController,
	userIssuesController *controllers.UserIssuesController,
	integrationController *controllers.IntegrationController,
	integrationAuthController *controllers.IntegrationAuthController) *MainRouter {
	return &MainRouter{
		mainController:            mainController,
		paymentsController:        paymentsController,
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
		userRouterGroup.POST("/complete-upgrade-pro", mr.paymentsController.StripeCheckoutCompleteHandler)
		userRouterGroup.GET("/containers", mr.userIssuesController.GetContainers)
		userRouterGroup.GET("/issues", mr.userIssuesController.IssuesSearch)
		userRouterGroup.GET("/issues/:id", mr.userIssuesController.GetIssue)
		userRouterGroup.PUT("/issues/:id/regenerate", mr.userIssuesController.RegenerateSolution)
		userRouterGroup.PUT("/issues/:id/resolve", mr.userIssuesController.ResolveIssue)
		userRouterGroup.PUT("/issues/:id/score", mr.userIssuesController.RateIssue)
		userRouterGroup.GET("/issues/:id/metrics/copied-sources-links", mr.userIssuesController.MetricsCopiedSourcesLinksHandler)
		userRouterGroup.POST("/issues/report", mr.userIssuesController.ReportIssueAnalysis)
		userRouterGroup.GET("/last-activity", mr.userController.LastActivityHandler)
		userRouterGroup.GET("/manage-pro", mr.paymentsController.StripeCreateBillingPortalHandler)
		userRouterGroup.GET("/settings", func(c *gin.Context) {})
		userRouterGroup.POST("/settings", func(c *gin.Context) {})
		userRouterGroup.POST("/upgrade-pro", mr.paymentsController.UpgradeProHandler)
	}

	agentRouterGroup := rg.Group("/agent", mr.integrationAuthController.CheckAgentAuthorization)
	{
		agentRouterGroup.DELETE("/issues/:containerId", mr.integrationController.DeleteIssues)
		agentRouterGroup.PUT("/issues/analysis", mr.integrationController.LogAnalysisTask)
	}

	integrationRouterGroup := rg.Group("/integration", middlewares.CheckAuthorization)
	{
		integrationRouterGroup.POST("/issues/:id/add-code-as-context", mr.integrationController.AddCodeAsContext)
	}

	metricsRouterGroup := rg.Group("/metrics", middlewares.CheckAuthorization)
	{
		metricsRouterGroup.POST("/overall-score", mr.userController.MetricsOverallScoreHandler)
		metricsRouterGroup.GET("/pro-btn-clicks", mr.userController.MetricsProButtonClickHandler)
		metricsRouterGroup.GET("/pro-checkout-clicks", mr.userController.MetricsProCheckoutClickHandler)
	}
}
