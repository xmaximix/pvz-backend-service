package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"pvz-backend-service/config"
	"pvz-backend-service/internal/api"
	"pvz-backend-service/internal/auth"
	"pvz-backend-service/internal/logger"
	"pvz-backend-service/internal/metrics"
	"pvz-backend-service/internal/repo"
	"pvz-backend-service/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := connectDB(ctx, cfg)
	defer db.Close()

	rep := repo.New(db)
	svc := service.New(rep, cfg.JWTSecret)

	router := gin.New()
	router.Use(logger.Middleware(), metrics.Middleware(), auth.Middleware(cfg.JWTSecret))
	api.RegisterHTTP(router, svc)

	go func() {
		addr := ":" + cfg.AppPort
		log.Printf("HTTP listening on %s", addr)
		if err := router.Run(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		addr := ":" + cfg.MetricsPort
		srv := &http.Server{Addr: addr, Handler: mux}
		log.Printf("Metrics listening on %s/metrics", addr)
		go func() { <-ctx.Done(); srv.Shutdown(context.Background()) }()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Metrics server error: %v", err)
		}
	}()

	go func() {
		addr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("gRPC listen error: %v", err)
		}
		grpcSrv := grpc.NewServer()
		api.RegisterGRPC(grpcSrv, rep)
		reflection.Register(grpcSrv)
		log.Printf("gRPC listening on %s", addr)
		go func() { <-ctx.Done(); grpcSrv.GracefulStop() }()
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, exiting…")
}

func connectDB(ctx context.Context, cfg config.Config) *pgxpool.Pool {
	for i := 0; i < cfg.DBMaxRetries; i++ {
		if db, err := pgxpool.New(ctx, cfg.DatabaseURL); err == nil {
			return db
		} else {
			log.Printf("DB connect failed (%d/%d): %v", i+1, cfg.DBMaxRetries, err)
			time.Sleep(cfg.DBRetryDelay)
		}
	}
	log.Fatal("DB connection retries exhausted")
	return nil
}
