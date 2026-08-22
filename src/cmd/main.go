// @title           HelpMe Booking Service API
// @version         1.0
// @description     Owns bookings (Postgres via GORM). Resolves user_id against the User Service over HTTP before reading/writing a booking.

// @contact.name   API Support
// @contact.email  mohit.chack@rosstechnologies.in

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.

package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go-helpme-booking/docs"
	"go-helpme-booking/src/config"
	"go-helpme-booking/src/middleware"
	"go-helpme-booking/src/routes"
	"go-helpme-booking/src/utils/database"
	"go-helpme-booking/src/utils/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "run DB migrations then exit, without starting the server")
	flag.Parse()

	// Read APP_ENV and APP_PORT directly from OS env so we can bind the port
	// and route secrets before the full config.Load call.
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = config.EnvDev
	}

	var ln net.Listener
	var addr string
	if !*migrateOnly {
		appPort := os.Getenv("APP_PORT")
		if appPort == "" {
			appPort = config.Common.DefaultPort
		}
		addr = ":" + appPort

		// Bind the port immediately — OS slot is claimed before any blocking init.
		var err error
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to bind %s: %v", addr, err)
		}
	}

	// Load secrets into env vars (AWS SecretManager in prod, no-op otherwise).
	// Must run before config.Load so injected keys are visible to it.
	secretsCtx, secretsCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer secretsCancel()
	if err := config.LoadSecrets(secretsCtx, appEnv); err != nil {
		log.Fatalf("failed to load secrets: %v", err)
	}

	config.Load(".env")

	logger.Init(config.App.App.Env)
	defer logger.Sync()

	if err := config.App.Validate(); err != nil {
		logger.Fatal("invalid configuration", zap.Error(err))
	}

	if config.App.App.IsProd {
		gin.SetMode(gin.ReleaseMode)
	}

	database.Connect()

	if *migrateOnly {
		database.Migrate()
		logger.Info("migration complete, exiting (--migrate)")
		return
	}

	drainer := new(middleware.Drainer)
	r := gin.New()
	routes.Setup(r, drainer)

	srv := &http.Server{
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server listening", zap.String("addr", addr), zap.String("env", config.App.App.Env))
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received, waiting for in-flight requests...")
	drainer.Drain()
	logger.Info("all requests completed, closing server")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited cleanly")
}
