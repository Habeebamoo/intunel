package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	producer := queue.NewProducer(cfg, redisClient)

	// Init providers
	router := providers.NewRouter(cfg)

	// Init consumer and reaper
	scheduler := worker.NewScheduler(schedulerRepo, producer)
	consumer := queue.NewConsumer(redisClient, router)
	reaper := worker.NewReaper(cfg, redisClient, router)

	// run consumer, reaper, and scheduler in separate goroutines
	var wg sync.WaitGroup
	wg.Add(3)

	go func(){
		defer wg.Done()
		consumer.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		reaper.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		scheduler.Start(ctx)
	}()

	//health check
	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		http.ListenAndServe(":"+cfg.WorkerPort, nil)
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers finished in-flight jobs. Clean shutdown complete.")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout (15s) reached. Forcing exit.")
	}
}