package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/app"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rdb := config.ConnectRedis(cfg.RedisURI)

	container := app.NewContainer(cfg, db, rdb, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Workspace Background Worker is running...")

	container.KafkaWorker.Start(ctx)

	<-ctx.Done()
	log.Println("Shutting down workers gracefully...")
}
