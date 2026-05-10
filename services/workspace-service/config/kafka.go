package config

import (
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewKafkaReader(brokers string, topic string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(brokers, ","),
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        time.Second * 1,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})
}
