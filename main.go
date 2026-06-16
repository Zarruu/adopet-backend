package main

import (
	"log"
	"net/http"
	"os"

	"adopet-backend/config"
	"adopet-backend/handlers"
	"adopet-backend/middleware"
	"adopet-backend/routes"
	"adopet-backend/services"
	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Set Gin mode
	ginMode := config.GetEnv("GIN_MODE", "debug")
	gin.SetMode(ginMode)

	// Connect to databases
	authDB, petsDB, err := config.ConnectDatabases()
	if err != nil {
		log.Fatalf("Gagal menghubungkan database: %v", err)
	}
	defer authDB.Close()
	defer petsDB.Close()

	// Initialize Google Drive service
	gdriveSvc := services.NewGDriveService()
	if err := gdriveSvc.Init(); err != nil {
		log.Printf("Peringatan: Gagal inisialisasi Google Drive: %v", err)
		log.Println("Menggunakan penyimpanan lokal sebagai fallback")
	}

	// Initialize services
	authService := services.NewAuthService(authDB)
	petService := services.NewPetService(petsDB)
	notificationService := services.NewNotificationService(petsDB)
	adoptionService := services.NewAdoptionService(petsDB, notificationService)
	userService := services.NewUserService(authDB, petsDB)

	// Create default admin user
	adminUsername := config.GetEnv("ADMIN_USERNAME", "admin")
	adminEmail := config.GetEnv("ADMIN_EMAIL", "admin@adopet.com")
	adminPassword := config.GetEnv("ADMIN_PASSWORD", "admin123")
	adminName := config.GetEnv("ADMIN_NAME", "Administrator")

	if err := authService.CreateDefaultAdmin(adminUsername, adminEmail, adminPassword, adminName); err != nil {
		log.Printf("Peringatan: Gagal membuat admin default: %v", err)
	} else {
		log.Println("Admin default sudah tersedia")
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	petHandler := handlers.NewPetHandler(petService)
	adoptionHandler := handlers.NewAdoptionHandler(adoptionService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	uploadHandler := handlers.NewUploadHandler(gdriveSvc)
	dashboardHandler := handlers.NewDashboardHandler(userService)
	userHandler := handlers.NewUserHandler(userService)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		utils.Success(c, http.StatusOK, "Adopet Backend is running", gin.H{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// Serve static files for local uploads
	if gdriveSvc.IsLocal() {
		r.Static("/uploads", "./uploads")
		log.Println("Mode penyimpanan lokal aktif - file disajikan dari /uploads")
	}

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(authDB)

	// Setup routes
	h := &routes.Handlers{
		Auth:         authHandler,
		Pet:          petHandler,
		Adoption:     adoptionHandler,
		Notification: notificationHandler,
		Upload:       uploadHandler,
		Dashboard:    dashboardHandler,
		User:         userHandler,
	}

	routes.SetupRoutes(r, h, authMiddleware)

	// Start server
	port := config.GetEnv("PORT", "8080")
	log.Printf("Server Adopet berjalan di port %s", port)

	serverAddr := ":" + port
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		serverAddr = "0.0.0.0:" + port
	}

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
