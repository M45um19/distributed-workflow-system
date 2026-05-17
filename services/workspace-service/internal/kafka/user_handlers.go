package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type UserRegisteredHandler struct {
	svc domain.UserService
}

func NewUserRegisteredHandler(svc domain.UserService) *UserRegisteredHandler {
	return &UserRegisteredHandler{svc: svc}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, msg kafka.Message) {
	var user domain.UserSnapshot
	if err := json.Unmarshal(msg.Value, &user); err != nil {
		log.Printf("Kafka Unmarshal Error [user-registered]: %v", err)
		return
	}

	if err := h.svc.SyncUserSnapshot(ctx, &user); err != nil {
		log.Printf("Failed to sync user snapshot via Kafka: %v", err)
		return
	}
}
