package main

import (
	"authSystem/controllers"
	"authSystem/initializers"
	"authSystem/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"go.uber.org/zap"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func init() {
	// Initialize configuration
	if err := initializers.LoadEnvVars(); err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	fmt.Println("Initializing application...")
	
	// Initialize logger first so we can use it for other initializations
	logger, err := middleware.InitLogger(middleware.DefaultLoggerConfig())
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer middleware.CloseLogger()

	// Initialize database connection
	if err := initializers.ConnectToDB(); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	logger.Info("Connected to database successfully")

	// Sync database schema
	if err := initializers.SyncDatabase(); err != nil {
		logger.Fatal("Failed to sync database", zap.Error(err))
	}
	logger.Info("Database schema synced successfully")
}

func main() {
	// Get logger instance
	log := middleware.GetLogger()

	// Configure graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Set Gin mode
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Middleware setup - ORDER IS IMPORTANT!
	r.Use(
		middleware.RecoveryWithLogger(), // First - handle panics
		requestid.New(),                // Add request IDs
		middleware.Logger(),            // Our custom logger
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", os.Getenv("FRONTEND_URL")},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		middleware.RateLimitMiddleware(100, 200, time.Minute), // Sensible defaults
	)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": os.Getenv("APP_VERSION"),
		})
	})

	// Authentication routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signup", controllers.SignUp)
		authGroup.POST("/login", controllers.Login)
		authGroup.GET("/validate", middleware.RequireAuth, controllers.Validate)
		authGroup.GET("/logout", middleware.RequireAuth, controllers.Logout)
	}

	// Book routes with authentication
	bookController := controllers.NewBookController()
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.RequireAuth)
	{
		apiGroup.GET("/book/:id", bookController.GetBookByID)
		apiGroup.POST("/book", bookController.CreateBook)
		apiGroup.PATCH("/book/:id", bookController.UpdateBook)
		apiGroup.DELETE("/book/:id", bookController.DeleteBook)
	}

	// Admin routes 
	adminGroup := r.Group("/admin")
	UserController := controllers.NewUserController()
	adminGroup.Use(middleware.RequireAuth, middleware.RequireAdmin)
	{
		adminGroup.GET("/users", UserController.GetAllUsers)
		adminGroup.GET("/books", bookController.GetAllBooks)
	}

	// Start server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Run server in a goroutine
	go func() {
		log.Info("Starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	
	// Restore default behavior on the interrupt signal
	stop()
	log.Info("Shutting down server...")

	// Context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited properly")
}