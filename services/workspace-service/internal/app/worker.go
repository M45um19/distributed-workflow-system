package app

import (
	"context"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"
)

type EventHandler interface {
	Handle(ctx context.Context, msg kafka.Message)
}

type KafkaWorker struct {
	readers  []*kafka.Reader
	handlers map[string]EventHandler
}

func NewKafkaWorker() *KafkaWorker {
	return &KafkaWorker{
		handlers: make(map[string]EventHandler),
	}
}

func (w *KafkaWorker) AddTopicHandler(reader *kafka.Reader, handler EventHandler) {
	w.readers = append(w.readers, reader)
	w.handlers[reader.Config().Topic] = handler
}

func (w *KafkaWorker) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for _, r := range w.readers {
		wg.Add(1)
		go func(reader *kafka.Reader) {
			defer wg.Done()
			topic := reader.Config().Topic
			log.Printf("Worker started for topic: %s", topic)

			for {
				select {
				case <-ctx.Done():
					log.Printf("Stopping worker for topic: %s", topic)
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
