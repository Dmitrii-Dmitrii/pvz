package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Dmitrii-Dmitrii/pvz/api"
	"github.com/Dmitrii-Dmitrii/pvz/internal"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/product_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/pvz_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/reception_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/user_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/generated"
	"github.com/Dmitrii-Dmitrii/pvz/internal/middlewares"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/product_service"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/pvz_service"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/reception_service"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/user_service"
	pvz_v1 "github.com/Dmitrii-Dmitrii/pvz/proto/generated/pvz/v1"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var registry = prometheus.NewRegistry()

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.GlobalLevel())
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	registry.MustRegister(internal.HttpRequestsTotal)
	registry.MustRegister(internal.HttpRequestDuration)
	registry.MustRegister(internal.PvzCreatedTotal)
	registry.MustRegister(internal.ReceptionCreatedTotal)
	registry.MustRegister(internal.ProductCreatedTotal)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrEnvLoading.Message)
	}

	connString := os.Getenv("CONNECTION_STRING")
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

	prometheusAddr := getPrometheusAddress()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

		log.Info().Msgf("Prometheus metrics starting on %s", prometheusAddr)
		if err := http.ListenAndServe(":9000", mux); err != nil {
			log.Error().Err(err).Msg("Failed to start Prometheus metrics server")
		}
	}()

	grpcAddr := getGrpcAddress()

	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to listen on %s", grpcAddr)
			return
		}

		grpcServer := grpc.NewServer()

		pvzGrpcHandler := api.NewGrpcHandler(pvzService)
		pvz_v1.RegisterPVZServiceServer(grpcServer, pvzGrpcHandler)

		reflection.Register(grpcServer)

		log.Info().Msgf("gRPC server starting on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error().Err(err).Msg("Failed to start gRPC server")
		}
	}()

	router := gin.Default()

	router.Use(middlewares.PrometheusMiddleware())

	authMiddleware := middlewares.NewAuthMiddleware(userService)

	apiGroup := router.Group("/")

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
		log.Info().Msg(fmt.Sprintf("Server starting on %s", server.Addr))
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

func getServerAddress() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

func getPrometheusAddress() string {
	port := os.Getenv("PROMETHEUS_PORT")
	if port == "" {
		port = "9000"
	}
	return ":" + port
}

func getGrpcAddress() string {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "3000"
	}
	return ":" + port
}
