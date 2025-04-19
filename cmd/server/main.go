package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"pvz/api"
	"pvz/internal/drivers/product_driver"
	"pvz/internal/drivers/pvz_driver"
	"pvz/internal/drivers/reception_driver"
	"pvz/internal/drivers/user_driver"
	"pvz/internal/generated"
	"pvz/internal/middlewares"
	"pvz/internal/models/custom_errors"
	"pvz/internal/services/product_service"
	"pvz/internal/services/pvz_service"
	"pvz/internal/services/reception_service"
	"pvz/internal/services/user_service"
	"syscall"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrEnvLoading.Message)
	}

	setupLogger()

	connString := "postgres://pvz_user:pvz_password@localhost:5432/pvz_database?sslmode=disable"
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCreatePool.Message)
	}
	log.Info().Msg("Connected to database")
	defer dbpool.Close()

	pvzDriver := pvz_driver.NewPvzDriver(dbpool)
	receptionDriver := reception_driver.NewReceptionDriver(dbpool)
	productDriver := product_driver.NewProductDriver(dbpool)
	userDriver := user_driver.NewUserDriver(dbpool)

	pvzService := pvz_service.NewPvzService(pvzDriver)
	receptionService := reception_service.NewReceptionService(receptionDriver)
	productService := product_service.NewProductService(productDriver, receptionService)
	userService := user_service.NewUserService(userDriver)

	httpHandler := api.NewHttpHandler(pvzService, receptionService, productService, userService)

	router := gin.Default()

	authMiddleware := middlewares.NewAuthMiddleware(userService)

	apiGroup := router.Group("/api/v1")

	generated.RegisterHandlersWithOptions(apiGroup, httpHandler, generated.GinServerOptions{
		Middlewares: []generated.MiddlewareFunc{
			authMiddleware.AuthMiddleware,
		},
	})

	server := &http.Server{
		Addr:         getServerAddress(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Msg(fmt.Sprintf("Server starting on %s\n", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg(custom_errors.ErrStartServer.Message)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrShutdownServer.Message)
	}

	log.Info().Msg("Server exiting")
}

func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.GlobalLevel())
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func getServerAddress() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
