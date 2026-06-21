package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/redis/go-redis/v9"
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

type UserLogoutHandler struct {
	rdb *redis.Client
}

func NewUserLogoutHandler(rdb *redis.Client) *UserLogoutHandler {
	return &UserLogoutHandler{rdb: rdb}
}

func (h *UserLogoutHandler) Handle(ctx context.Context, msg kafka.Message) {

	var userLogout domain.UserLogoutPayload
	if err := json.Unmarshal(msg.Value, &userLogout); err != nil {
		log.Printf("Kafka Unmarshal Error [user-logout]: %v", err)
		return
	}
	userID := fmt.Sprintf("%v", userLogout.UserID)
	deviceId := fmt.Sprintf("%v", userLogout.DeviceID)
	redisKey := fmt.Sprintf("session:%s:%s", userID, deviceId)

	// remove local cache session
	sessionData, err := h.rdb.Get(ctx, redisKey).Result()
	if err == nil && sessionData != "" {
		h.rdb.Del(ctx, redisKey)
		return
	}
}
