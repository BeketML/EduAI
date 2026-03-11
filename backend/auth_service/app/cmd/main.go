// @title management auth
// @version 1.0
// @description API for project management auth
// @termsOfService http://example.com/terms/
// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @Security ApiKeyAuth
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	_ "auth_service/docs"
	"auth_service/internal/infrastructure/auth"
	"auth_service/internal/infrastructure/logger"
	"auth_service/internal/infrastructure/postgres"
	"auth_service/internal/infrastructure/repository"
	"auth_service/internal/interfaces/http/handler"
	"auth_service/internal/interfaces/http/middleware"
	"auth_service/internal/usecase"
	"context"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := logger.New("dev")
	ctx := context.Background()
	log.Info(ctx, "App is running")
	if err := initConfig(); err != nil {
		log.Error(ctx, "init config error : ", err.Error())
	}
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

	if accessSecret == "" || refreshSecret == "" {
		log.Error(ctx, "JWT secrets are not set")
	}

	retryCfg := postgres.RetryConfig{
		MaxAttempts: 10,
		Delay:       3 * time.Second,
		Timeout:     30 * time.Second,
	}
	log.Info(ctx, "db config",
		"sslmode", viper.GetString("db.sslmode"),
	)

	db, err := postgres.ConnectWithRetry(
		ctx,
		retryCfg,
		viper.GetString("db.username"),
		os.Getenv("DB_PASSWORD"),
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.dbname"),
		viper.GetString("db.sslmode"),
	)
	if err != nil {
		log.Error(ctx, "db connect failed", "error", err)
		return
	}
	tokenManager := auth.NewTokenManager(accessSecret, refreshSecret)

	repos := repository.NewRepository(db, log)
	services := usecase.NewService(repos, log, tokenManager)
	handlers := handler.NewHandler(services, log)
	router := handlers.InitRouter()
	routerWithMiddleware := middleware.RequestID(router)
	srv := new(handler.Server)
	go func() {
		log.Info(ctx, "Leaderboard app starting", "port", viper.GetString("port"))
		if err := srv.Run(viper.GetString("port"), routerWithMiddleware); err != nil {
			log.Error(ctx, "server run error", "error", err)
		}
	}()

	log.Info(ctx, "pm project app starting")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Info(ctx, "pm project is shutting down")
	if err := srv.Shutdown(); err != nil {
		log.Error(ctx, "Error occured on server shutting down: ", err.Error())
	}
	if err := db.Close(); err != nil {
		log.Error(ctx, "Error occured on db connection close: ", err.Error())
	}
}

func initConfig() error {
	viper.SetConfigName("config") // config.yml
	viper.SetConfigType("yaml")   // 🔥 важно
	viper.AddConfigPath(".")
	return viper.ReadInConfig()
}
