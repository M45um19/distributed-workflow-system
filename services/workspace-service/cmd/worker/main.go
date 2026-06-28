package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/app"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/monitoring"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	cleanup := monitoring.InitTracer(cfg.ServiceName, cfg.OtelExporterOtlpEndpoint)
	defer cleanup()

	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	rdb := config.ConnectRedis(cfg.RedisURI)

	container := app.NewContainer(cfg, db, rdb, nil, true)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Workspace Background Worker is running...")

	go container.KafkaWorker.Start(ctx)
	go container.TemporalWorker.Start(ctx)

	<-ctx.Done()
	log.Println("Shutting down workers gracefully...")

	container.KafkaWorker.Stop()
	log.Println("Kafka Worker components stopped.")

	if container.TemporalClient != nil {
		container.TemporalClient.Close()
		log.Println("Temporal Client closed.")
	}

	if rdb != nil {
		rdb.Close()
		log.Println("Redis client closed.")
	}

	db.Close()
	log.Println("PostgreSQL connection closed safely. Goodbye!")
}
