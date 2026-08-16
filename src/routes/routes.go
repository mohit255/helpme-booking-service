package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-helpme-booking/src/clients"
	"go-helpme-booking/src/config"
	"go-helpme-booking/src/controllers"
	"go-helpme-booking/src/middleware"
	"go-helpme-booking/src/repositories"
	"go-helpme-booking/src/services"
	"go-helpme-booking/src/utils/database"
)

type healthResponse struct {
	Status   string `json:"status"    example:"ok"`
	Env      string `json:"env"       example:"dev"`
	Database string `json:"database"  example:"ok"`
}

// Health godoc
// @Summary      Health check
// @Description  Returns service liveness and database connectivity status
// @Tags         Health
// @Produce      json
// @Success      200  {object}  healthResponse
// @Success      503  {object}  healthResponse
// @Router       /health [get]
func Health(c *gin.Context) {
	dbStatus := "ok"
	httpStatus := http.StatusOK

	if err := database.Ping(); err != nil {
		dbStatus = "unreachable"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, healthResponse{
		Status:   "ok",
		Env:      config.App.App.Env,
		Database: dbStatus,
	})
}

func Setup(r *gin.Engine, drainer *middleware.Drainer) {
	r.Use(drainer.Handler())
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.RateLimit(config.DefaultRateLimitPerMin))

	// Swagger UI — disabled in prod
	if !config.App.App.IsProd {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// wire up dependencies
	userClient := clients.NewUserServiceClient()
	bookingRepo := repositories.NewBookingRepository()
	bookingSvc := services.NewBookingService(bookingRepo)
	bookingCtrl := controllers.NewBookingController(userClient, bookingSvc)

	v1 := r.Group(config.APIV1)
	{
		// Public endpoints
		v1.GET("/health", Health)

		// Protected endpoints — require valid JWT
		protected := v1.Group("")
		protected.Use(middleware.Authenticate())
		{
			bookings := protected.Group("/bookings")
			{
				bookings.GET("", bookingCtrl.List)
				bookings.POST("", bookingCtrl.Create)
			}
		}
	}
}
