package temporal

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	temporalWorker "go.temporal.io/sdk/worker"
)

type Worker struct {
	client client.Client
	worker temporalWorker.Worker
}

func NewWorker(c client.Client, taskQueue string) *Worker {
	w := temporalWorker.New(c, taskQueue, temporalWorker.Options{})
	return &Worker{
		client: c,
		worker: w,
	}
}

func (tw *Worker) Register(registryFn func(temporalWorker.Worker)) {
	registryFn(tw.worker)
}

func (tw *Worker) Start(ctx context.Context) {
	log.Println("Temporal Background Worker is starting...")

	go func() {
		<-ctx.Done()
		log.Println("Context cancelled. Stopping Temporal worker workflow listeners...")
		tw.worker.Stop()
	}()

	if err := tw.worker.Run(temporalWorker.InterruptCh()); err != nil {
		log.Printf("Temporal worker error: %v", err)
	}
	log.Println("Temporal Worker stopped gracefully.")
}
