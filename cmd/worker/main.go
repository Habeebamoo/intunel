package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Habeebamoo/intunel-backend/internal/configs"
	"github.com/Habeebamoo/intunel-backend/internal/database"
	"github.com/Habeebamoo/intunel-backend/internal/providers"
	"github.com/Habeebamoo/intunel-backend/internal/queue"
	"github.com/Habeebamoo/intunel-backend/internal/repositories"
	"github.com/Habeebamoo/intunel-backend/internal/worker"
)

func main() {
	cfg := configs.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisClient, err := configs.InitRedis(cfg, ctx)
	if err != nil {
		log.Fatal("couldn't connect to redis: ", err)
		os.Exit(1)
	}

	//Init db
	db := database.NewPostgres(cfg)
	schedulerRepo := repositories.NewScheduledNotificationRepository(db)
	producer := queue.NewProducer(redisClient)

	// Init providers
	router := providers.NewRouter(cfg)

	// Init consumer and reaper
	scheduler := worker.NewScheduler(schedulerRepo, producer)
	consumer := queue.NewConsumer(redisClient, router)
	reaper := worker.NewReaper(redisClient, router)

	// run consumer, reaper, and scheduler in separate goroutines
	go consumer.Start(ctx)
	go reaper.Start(ctx)
	go scheduler.Start(ctx)

	//health check
	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		http.ListenAndServe(":"+cfg.Port, nil)
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker shutting down...")
	cancel()
}