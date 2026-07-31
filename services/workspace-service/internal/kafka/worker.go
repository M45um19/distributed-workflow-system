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

type KafkaLoopHandler interface {
	RunLoop(ctx context.Context, reader *kafka.Reader)
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

			handler := w.handlers[topic]
			if loopHandler, ok := handler.(KafkaLoopHandler); ok {
				loopHandler.RunLoop(ctx, reader)
				return
			}

			for {
				msg, err := reader.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe) {
						log.Printf("Kafka worker loop stopped for topic: %s", topic)
						return
					}
					log.Printf("Kafka Read Error (%s): %v", topic, err)
					return
				}

				handler.Handle(ctx, msg)

				if err := reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("Kafka Commit Error (%s): %v", topic, err)
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
