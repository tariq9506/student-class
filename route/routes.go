package route

import (
	"net/http"
	"os"
	"tutree/student-apis/controllers/ambassador"
	email "tutree/student-apis/controllers/email"
	"tutree/student-apis/controllers/metadata"
	"tutree/student-apis/controllers/middleware"
	phonecontroller "tutree/student-apis/controllers/phoneController"
	"tutree/student-apis/controllers/session"
	"tutree/student-apis/controllers/stripe"
	studentscontroller "tutree/student-apis/controllers/studentsController"
	"tutree/student-apis/controllers/timezone"
	twiliowebhook "tutree/student-apis/controllers/twilio_webhook"
	"tutree/student-apis/controllers/user"

	"tutree/student-apis/controllers/tutor"
	"tutree/student-apis/docs"
	"tutree/student-apis/utility"

	"github.com/gin-gonic/gin"

	"tutree/student-apis/controllers"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// AddRoutes is responsible for adding all the routes so the server can handle
// new routes. this means that we can reuse this function for multiple prefixes.
// prefixes like job_portal are necessary for legacy url handling.
func AddRoutes(router *gin.RouterGroup) {
	router.GET("/token", controllers.GetZohoAccesstoken)
	// NOTE :- all api must be in this group and for every particular feature apis must be create new group.
	api := router.Group("/v1", controllers.ApiTracker)
	{
		// ALL THE STUDENT API'S WILL COME UNDER THIS GROUP

		// This api is responsible for student registration
		api.POST("/student/authenticate", studentscontroller.StudentRegistration)
		// This api is responsible for ambassador registration
		api.POST("/ambassador/authenticate", ambassador.AmbassadorRegistration)
		// This api is resposible to fetch ambassador statistics
		api.GET("/ambassador/stats", ambassador.GetAmbassadorStats)
		// This api is responsile for varify the OTP
		api.POST("/otp/verify", phonecontroller.VerifyCode)
		// This api is responsible for resend otp on phone number.
		api.POST("/otp/send", phonecontroller.ResendVerificationCode)
		//This api is responsible for fetching student profile from database.
		api.GET("/student/profile", middleware.TokenAuthMiddleware, studentscontroller.GetStudentProfile)
		// This api is responsible for user logout process by invalidating the user's active session
		api.DELETE("/logout", middleware.TokenAuthMiddleware, user.DeactivateUserSession)
		// This aoi is responsible to add a chiled for user.
		api.POST("/user/child", middleware.TokenAuthMiddleware, studentscontroller.AddChildForUser)
		api.GET("/user/child", middleware.TokenAuthMiddleware, studentscontroller.GetChildOfUser)
		api.PUT("/user/child", middleware.TokenAuthMiddleware, studentscontroller.UpdateChildInformation)

		api.POST("/twilio-webhook", twiliowebhook.TwilioWebhook)

		sessionGroup := api.Group("/session", middleware.TokenAuthMiddleware)
		{
			// this api is responsible to book student's demo session.
			sessionGroup.POST("/book-demo", session.BookDemoSession)

			// this api is responsible to retrieve the sessionlist of logged in user.
			sessionGroup.GET("/list", session.GetSessionlist)

			// this api is responsible to retrieve the session detail of logged in user.
			sessionGroup.GET("", session.GetSession)

			// this api is responsible to cancel the session scheduled session before starting of the session(before X minutes).
			sessionGroup.DELETE("/cancel", session.CancelSession)

			// this api is responsible to add the feedback of logged in user for thier session id.
			sessionGroup.POST("/feedback", session.AddSessionFeedback)

			sessionGroup.OPTIONS("/book-demo", optionsHandler)
			// This api is responsile for re-schedule student session.
			sessionGroup.PUT("/demo-reschedule", session.RescheduleDemoSession)
			// This api is responsible for update student profile.
			sessionGroup.PUT("/student/profile", studentscontroller.UpdateStudentProfileDetails)
			// This api is responsible for booking regualr session for students.
			sessionGroup.POST("/book", session.BookRegularSession)
			// This api is responsible for session re-scheduling later for students.
			sessionGroup.POST("/reschedule-later", session.SessionRescheduleLater)
			// This api is responsible for delete session from session table.
			sessionGroup.DELETE("", session.DeleteSession)
			// This api is responsible for get list of pending session of any student.
			sessionGroup.GET("/pending-book", session.GetExpectedSession)
			// This api is responsible for booking pending sessions.
			sessionGroup.POST("/pending-book", session.BookPendingSession)

			preference := sessionGroup.Group("/preference")
			{
				preference.GET("", session.GetSessionPreferenceById)
				preference.GET("/all", session.GetAllSessionPreferences)
				preference.POST("", session.AddSessionPreference)
				preference.PUT("", session.UpdateSessionPreference)
				preference.DELETE("", session.DeleteSessionPreference)
			}

		}

		api.GET("/tutor/slots/demo", middleware.TokenAuthMiddleware, tutor.GetTutorAvailableDemoSlots)
		api.GET("/tutor/slots/reschedule", middleware.TokenAuthMiddleware, tutor.GetTutorAvailableSlotsForReschedule)
		api.GET("/tutor/slots", middleware.TokenAuthMiddleware, tutor.GetTutorAvailableSlots)
		api.GET("/timezones", timezone.GetTimezones)

		subscription := api.Group("/subscription")
		{
			subscription.GET("", stripe.GetSubscription)
			subscription.GET("/plans", stripe.ListSubscriptionPlans)
			subscription.POST("", middleware.TokenAuthMiddleware, stripe.BuySubscription)
			subscription.POST("/webhook", stripe.StripeWebhook)
			subscription.GET("/active-plan", middleware.TokenAuthMiddleware, stripe.GetActivePlanDetails)
			subscription.POST("/renew", stripe.RenewSubscription)
			subscription.PUT("/user-info-stripe", stripe.GetAllCustomerInfoToUpdateOnstripe)

			subscription.POST("/payment-method", middleware.TokenAuthMiddleware, stripe.SetPaymentMethod)

		}
		api.GET("/location-from-ip", controllers.GetLocationDetailsFromIP)

		meta := api.Group("/metadata")
		{
			// This API is used to fetch all the subjects
			meta.GET("/subjects", metadata.GetSubjects)
			// This meta is used to fetch all the grades
			meta.GET("/grades", metadata.GetGrades)
		}
		api.GET("/zoho-leads", controllers.GetPhoneOfColdLeads)
		api.GET("/call-details", controllers.GetCallDetails)

	}
	// Grouping routes for APIs that are triggered by cron jobs.
	cron := router.Group("/cron")
	{
		// This route is used to retrieve the list of users who have scheduled their first regular session today.
		// It then updates the billing cycle for users who haven't scheduled their only class.
		cron.GET("/schedule-today", session.GetUsersWithTodaySession)
		// this api is resposible for sent reminder email before start demo session.
		cron.POST("/session/reminder-email", email.SendReminderEmailForSession)
		// This API is  responsible for sending reminder email for book regular session after purchasing plan.
		cron.POST("/session/schedule/reminder-email", email.SendReminderEmailForBookRegularSession)
		// This API is responsible for sending reminder SMS after sign in for scheduling demo session.
		cron.POST("/session/reminder-sms", studentscontroller.SendReminderSMSForDemoBooking)
		// This API is responsible for sending reminder emails and SMS messages to students 2 hours before their class session starts.
		cron.POST("/session/reminder/2hours", email.SessionReminderNotificationBefore2Hour)
		// This API is responsible for sending reminder emails and SMS messages to students 24 hours before their class session starts.
		cron.POST("/session/reminder/24hours", email.SessionReminderNotificationBefore24Hour)
		cron.POST("/paid-session/48hours", email.ScheduleOneDayBeforeReminderForPaidSession)

		cron.PUT("/session/cancel/not-confirmed", session.SessionCancelIfNotConfirmed)

		cron.POST("reminder/schedule-class", email.SendReminderToScheduleClass)

	}

	ppc := router.Group("/ppc")
	{
		// ALL THE STUDENT API'S WILL COME UNDER THIS GROUP

		// This api is responsible for student registration
		ppc.POST("/student/authenticate", studentscontroller.StudentRegistrationPPC)
	}

}

// SetupRouter sets up routes
func SetupRouter() *gin.Engine {

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/v1"
	url := ginSwagger.URL(utility.GetHostURL() + "/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))
	//router.Use(cors.Default())
	// corsConfig := cors.Config{
	// 	AllowOrigins:     []string{"*"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }

	// // Handle preflight requests
	// if c.Request.Method == "OPTIONS" {
	// 	c.AbortWithStatus(http.StatusNoContent)
	// 	return
	// }
	// // Apply CORS middleware with logging
	// router.Use(func(c *gin.Context) {
	// 	log.Println("CORS middleware executed")
	// 	c.Next()
	// })
	router.Use(corsMiddleware())
	gin.SetMode(os.Getenv("GIN_MODE"))
	router.Static("/domain", "./domainInfo")
	router.Static("/pdf", "./pdf")
	router.Static("/og_images", "./og_images")

	// Add all current URls
	AddRoutes(&router.RouterGroup)

	return router
}

// corsMiddleware sets the necessary headers for CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization,token")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// optionsHandler handles preflight OPTIONS requests
func optionsHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, token")
	c.Status(http.StatusOK)
}
