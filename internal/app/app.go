package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/database"
	"github.com/Habeebamoo/intunel-backend/internal/handlers"
	"github.com/Habeebamoo/intunel-backend/internal/queue"
	"github.com/Habeebamoo/intunel-backend/internal/repositories"
	"github.com/Habeebamoo/intunel-backend/internal/services"
	"github.com/Habeebamoo/intunel-backend/internal/store"
	"github.com/gin-gonic/gin"
)

type App struct {
	server *http.Server
	router *gin.Engine
	config *configs.Config
}

func New() *App {
	cfg := configs.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisClient, err := configs.InitRedis(cfg, ctx)
	if err != nil {
		log.Fatal("couldn't connect to redis: ", err)
		os.Exit(1)
	}

	// Postgres init
	db := database.NewPostgres(cfg)
	database.Migrate(db)

	// Gin init
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()

	//idempotent store
	idempotencyStore := store.NewIdempotencyStore(redisClient)

	//repositories init
	schedulerRepo := repositories.NewScheduledNotificationRepository(db)

	// Services init
	producer := queue.NewProducer(cfg, redisClient)
	notificationService := services.NewNotificationService(producer, schedulerRepo)

	// Handlers init
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Routes
	RegisterRoutes(
		router, 
		notificationHandler, 
		idempotencyStore,
	)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		server: server,
		router: router,
		config: cfg,
	}
}

func (a *App) Run() error {
	go func() {
		log.Printf("intunl running on PORT :%s", a.config.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.server.Shutdown(ctx)
}