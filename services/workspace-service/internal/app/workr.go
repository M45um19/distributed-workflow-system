package app

import (
	"context"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/user"
	"github.com/segmentio/kafka-go"
)

type KafkaWorker struct {
	reader  *kafka.Reader
	handler *user.KafkaHandler
}

func NewKafkaWorker(reader *kafka.Reader, handler *user.KafkaHandler) *KafkaWorker {
	return &KafkaWorker{
		reader:  reader,
		handler: handler,
	}
}

func (w *KafkaWorker) Start(ctx context.Context) {
	log.Println("Kafka Worker started.")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka Worker...")
			return
		default:
			msg, err := w.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Kafka Read Error: %v", err)
				return
			}

			if msg.Topic == "user-registered" {
				w.handler.HandleUserRegistered(ctx, msg)
			}
		}
	}
}
