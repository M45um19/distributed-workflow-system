package user

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type UserRegisteredHandler struct {
	svc Service
}

func NewUserRegisteredHandler(svc Service) *UserRegisteredHandler {
	return &UserRegisteredHandler{svc: svc}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, msg kafka.Message) {

	var user UserSnapshot
	if err := json.Unmarshal(msg.Value, &user); err != nil {
		log.Printf("Error unmarshaling user event: %v", err)
		return
	}

	if err := h.svc.SyncUserSnapshot(ctx, &user); err != nil {
		log.Printf("Error processing user snapshot: %v", err)
		return
	}
}
