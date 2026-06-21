package config

import (
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewKafkaReader(brokers string, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(brokers, ","),
		Topic:          topic,
		GroupID:        "workspace-service-group",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        time.Second * 1,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})
}

func NewKafkaWriter(brokers string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        true,
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		WriteTimeout: 5 * time.Second,
	}
}
