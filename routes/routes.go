package routes

import (
	"adopet-backend/handlers"
	"adopet-backend/middleware"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth         *handlers.AuthHandler
	Pet          *handlers.PetHandler
	Adoption     *handlers.AdoptionHandler
	Notification *handlers.NotificationHandler
	Upload       *handlers.UploadHandler
	Dashboard    *handlers.DashboardHandler
	User         *handlers.UserHandler
}

func SetupRoutes(r *gin.Engine, h *Handlers, authMiddleware gin.HandlerFunc) {
	api := r.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.HandleRegister)
		auth.POST("/login", h.Auth.HandleLogin)
	}

	// Authenticated routes
	authenticated := api.Group("/")
	authenticated.Use(authMiddleware)
	{
		// Auth - profile management
		authenticated.POST("/auth/logout", h.Auth.HandleLogout)
		authenticated.GET("/auth/me", h.Auth.HandleGetProfile)
		authenticated.PUT("/auth/profile", h.Auth.HandleUpdateProfile)
		authenticated.PUT("/auth/password", h.Auth.HandleChangePassword)

		// Pets - read for all authenticated users
		authenticated.GET("/pets", h.Pet.HandleGetPets)
		authenticated.GET("/pets/:id", h.Pet.HandleGetPet)

		// Pets - write for editor and admin only
		petWrite := authenticated.Group("/pets")
		petWrite.Use(middleware.RequireRole("editor", "admin"))
		{
			petWrite.POST("", h.Pet.HandleCreatePet)
			petWrite.PUT("/:id", h.Pet.HandleUpdatePet)
			petWrite.DELETE("/:id", h.Pet.HandleDeletePet)
		}

		// Upload - for editor and admin only
		upload := authenticated.Group("/upload")
		upload.Use(middleware.RequireRole("editor", "admin"))
		{
			upload.POST("/image", h.Upload.HandleUploadImage)
		}

		// Adoptions - submit for all authenticated users
		authenticated.POST("/adoptions", h.Adoption.HandleSubmitAdoption)
		authenticated.GET("/adoptions/my", h.Adoption.HandleGetMyAdoptions)

		// Adoptions - management for editor and admin
		adoptionMgmt := authenticated.Group("/adoptions")
		adoptionMgmt.Use(middleware.RequireRole("editor", "admin"))
		{
			adoptionMgmt.GET("", h.Adoption.HandleGetAllAdoptions)
			adoptionMgmt.PUT("/:id/approve", h.Adoption.HandleApproveAdoption)
			adoptionMgmt.PUT("/:id/reject", h.Adoption.HandleRejectAdoption)
		}

		// Notifications - for all authenticated users
		authenticated.GET("/notifications", h.Notification.HandleGetNotifications)
		authenticated.PUT("/notifications/:id/read", h.Notification.HandleMarkAsRead)
		authenticated.GET("/notifications/unread-count", h.Notification.HandleGetUnreadCount)

		// Admin only routes
		admin := authenticated.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/dashboard", h.Dashboard.HandleGetDashboard)
			admin.GET("/users", h.User.HandleGetUsers)
			admin.POST("/users", h.User.HandleCreateUser)
			admin.PUT("/users/:id", h.User.HandleUpdateUser)
			admin.DELETE("/users/:id", h.User.HandleDeleteUser)
			admin.GET("/active-users", h.Dashboard.HandleGetActiveUsers)
		}
	}
}
