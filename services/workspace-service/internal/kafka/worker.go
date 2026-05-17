package kafka

import (
	"context"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"
)

type KafkaEventHandler interface {
	Handle(ctx context.Context, msg kafka.Message)
}

type Worker struct {
	readers  []*kafka.Reader
	handlers map[string]KafkaEventHandler
}

func NewWorker() *Worker {
	return &Worker{
		handlers: make(map[string]KafkaEventHandler),
	}
}

func (w *Worker) AddTopicHandler(reader *kafka.Reader, handler KafkaEventHandler) {
	w.readers = append(w.readers, reader)
	w.handlers[reader.Config().Topic] = handler
}

func (w *Worker) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for _, r := range w.readers {
		wg.Add(1)
		go func(reader *kafka.Reader) {
			defer wg.Done()
			topic := reader.Config().Topic
			log.Printf("Kafka Background Worker started for topic: %s", topic)

			for {
				select {
				case <-ctx.Done():
					log.Printf("Stopping Kafka worker for topic: %s", topic)
					reader.Close()
					return
				default:
					msg, err := reader.ReadMessage(ctx)
					if err != nil {
						log.Printf("Kafka Read Error (%s): %v", topic, err)
						return
					}
					if handler, ok := w.handlers[msg.Topic]; ok {
						handler.Handle(ctx, msg)
					}
				}
			}
		}(r)
	}

	wg.Wait()
}
