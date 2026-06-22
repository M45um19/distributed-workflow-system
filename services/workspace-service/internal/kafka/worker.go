package kafka

import (
	"context"
	"errors"
	"io"
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
	wg       sync.WaitGroup
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
	for _, r := range w.readers {
		w.wg.Add(1)
		go func(reader *kafka.Reader) {
			defer w.wg.Done()
			topic := reader.Config().Topic
			log.Printf("Kafka Background Worker started for topic: %s", topic)

			for {
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe) {
						log.Printf("Kafka worker loop stopped for topic: %s", topic)
						return
					}
					log.Printf("Kafka Read Error (%s): %v", topic, err)
					return
				}

				if handler, ok := w.handlers[msg.Topic]; ok {

					handler.Handle(context.Background(), msg)
				}
			}
		}(r)
	}

	w.wg.Wait()
}

func (w *Worker) Stop() {
	log.Println("Closing all Kafka readers...")
	for _, r := range w.readers {
		if err := r.Close(); err != nil {
			log.Printf("Error closing kafka reader: %v", err)
		}
	}
	w.wg.Wait()
}
