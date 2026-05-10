package user

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaHandler struct {
	svc Service
}

func NewKafkaHandler(svc Service) *KafkaHandler {
	return &KafkaHandler{svc: svc}
}

func (h *KafkaHandler) HandleUserRegistered(ctx context.Context, msg kafka.Message) {
	var user UserSnapshot
	if err := json.Unmarshal(msg.Value, &user); err != nil {
		log.Printf("Error unmarshaling user event: %v", err)
		return
	}

	if err := h.svc.SyncUserSnapshot(ctx, &user); err != nil {
		log.Printf("Error processing user snapshot: %v", err)
		return
	}

	log.Printf("Successfully synced user snapshot for: %s", user.Email)
}
