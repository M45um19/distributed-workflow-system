package app

import (
	"context"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/user"
	"github.com/segmentio/kafka-go"
)

type Worker struct {
	reader  *kafka.Reader
	handler *user.KafkaHandler
}

func NewWorker(reader *kafka.Reader, handler *user.KafkaHandler) *Worker {
	return &Worker{
		reader:  reader,
		handler: handler,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Kafka Worker started. Listening for events...")

	for {
		msg, err := w.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if msg.Topic == "user-registered" {
			w.handler.HandleUserRegistered(ctx, msg)
		}
	}
}
