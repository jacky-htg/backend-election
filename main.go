package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-election/internal/pkg/config"
	"backend-election/internal/pkg/database"
	"backend-election/internal/pkg/logger"
	"backend-election/internal/pkg/redis"
	"backend-election/internal/route"

	_ "github.com/lib/pq"
)

// @title Rest Skeleton API
// @version 1.0
// @description This is a sample server API.
// @Schemes http
// @BasePath /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	if _, ok := os.LookupEnv("APP_NAME"); !ok {
		if err := config.Setup(".env"); err != nil {
			fmt.Printf("failed to setup config: %v", err)
			os.Exit(1)
		}
	}

	log := logger.New()

	fmt.Println("Starting Server at : "+os.Getenv("APP_PORT"), "")

	db, err := database.NewDatabase()
	if err != nil {
		fmt.Printf("Could not connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Conn.Close()

	redisClient, err := redis.NewCache(context.Background(), os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PASSWORD"), 24*time.Hour)
	if err != nil {
		fmt.Printf("Could not connect to redis: %v", err)
	}
	defer redisClient.Close()

	srv := &http.Server{
		Addr:         ":" + os.Getenv("APP_PORT"),
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 30,
		Handler:      route.ApiRoute(log, db, redisClient),
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println("listen and serve", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutdown Server ...", "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("Server shutdown", err)
	}

	fmt.Println("Server exiting")
}
