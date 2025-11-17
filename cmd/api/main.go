package main

import (
	"file-sharing/internal/config"
	"file-sharing/internal/storage"
	"file-sharing/internal/transport/http"
	"file-sharing/internal/share"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// ---- 1. Khởi tạo Database Connection ----
	config, err := config.LoadConfig("./env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()
	log.Println("Connected to database")

	// ---- 2. Khởi tạo Repository ----
	userRepo := storage.NewUserRepository(db)
	shareRepo := storage.NewShareRepository(db)

	// ---- 3. Khởi tạo Gin Router ----
	router := gin.Default()

	// ---- Khởi tạo Service ----
    shareService := share.NewShareService(shareRepo)

	// --- Khởi tạo Auth Service ---
	authService := share.NewAuthService(shareRepo)

	// ---- 4. Khởi tạo Handlers & Middlewares ----
	userHandler := http.NewUserHandler()
	authMiddleware := http.AuthMiddleware(userRepo)
	fileHandler := http.InitFileUploadHandler(db.DB)
	listFilesHandler := http.ListUserFilesHandler(db.DB)
	shareHandler := http.NewShareHandler(shareService)
	// Authoize Password Handler
	authorizePasswordHandler := http.NewAuthorizePasswordHandler(authService)
    
    // Report handlers (register these so client endpoints exist)
    reportCompleteHandler := http.ReportUploadCompleteHandler(db.DB)
    getReportHandler := http.GetUploadReportHandler(db.DB)
    listReportsHandler := http.ListUploadReportsHandler(db.DB)

	// ---- 5. Đăng ký API Routes ----
	api := router.Group("/api")
	{
		authed := api.Group("/")
		authed.Use(authMiddleware)
		{
			authed.GET("/me", userHandler.GetCurrentUser)
			authed.POST("/v1/files", fileHandler)
			authed.GET("/v1/files", listFilesHandler)
            // report endpoints
            authed.POST("/v1/files/:file_id/report-complete", reportCompleteHandler)
            authed.GET("/v1/files/:file_id/report", getReportHandler)
            authed.GET("/v1/upload-reports", listReportsHandler)
			authed.POST("/v1/shares/:id/revoke", shareHandler.HandleRevoke)
			// authorize password endpoint
			authed.POST("/v1/shares/:id/authorize", authorizePasswordHandler.HandleAuthorizePassword)

			authed.GET("/v1/shares", shareHandler.HandleListShares)
		}
	}

	// ---- Swagger UI ----
	router.StaticFile("/openapi.yaml", "./api/openapi.yaml")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))

	// ---- 6. Khởi động HTTP Server ----
	log.Printf("Starting server on %s", config.HTTPServerAddress)
	if err := router.Run(config.HTTPServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
