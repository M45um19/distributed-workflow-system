package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type userEvent struct {
	msg  kafka.Message
	user domain.UserSnapshot
}

type UserRegisteredHandler struct {
	userRepo  domain.UserRepository
	dlqWriter *kafka.Writer
}

func NewUserRegisteredHandler(userRepo domain.UserRepository, dlqWriter *kafka.Writer) *UserRegisteredHandler {
	return &UserRegisteredHandler{
		userRepo:  userRepo,
		dlqWriter: dlqWriter,
	}
}

func (h *UserRegisteredHandler) Handle(ctx context.Context, msg kafka.Message) {
	// Satisfies KafkaEventHandler interface, but worker uses RunLoop (KafkaLoopHandler)
}

func (h *UserRegisteredHandler) RunLoop(ctx context.Context, reader *kafka.Reader) {
	topic := reader.Config().Topic
	batchSize := 100
	flushInterval := 1 * time.Second

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe) {
				log.Printf("Kafka worker loop stopped for topic: %s", topic)
				return
			}
			log.Printf("Kafka Fetch Error (%s): %v", topic, err)
			time.Sleep(1 * time.Second)
			continue
		}

		var batch []userEvent
		var msgs []kafka.Message

		var u domain.UserSnapshot
		if err := json.Unmarshal(msg.Value, &u); err == nil {
			batch = append(batch, userEvent{msg: msg, user: u})
			msgs = append(msgs, msg)
		} else {
			log.Printf("Kafka Unmarshal Error [user-registered]: %v", err)
			h.sendToDLQ(ctx, msg, fmt.Sprintf("Unmarshal error: %v", err))
			_ = reader.CommitMessages(ctx, msg)
		}

		deadline := time.Now().Add(flushInterval)
		for len(batch) < batchSize && time.Now().Before(deadline) {
			timeout := time.Until(deadline)
			fetchCtx, cancel := context.WithTimeout(ctx, timeout)
			nextMsg, err := reader.FetchMessage(fetchCtx)
			cancel()

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					break
				}
				if errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe) {
					if len(batch) > 0 {
						h.flushBatch(ctx, reader, batch, msgs)
					}
					return
				}
				log.Printf("Kafka Fetch Error during batch accumulation (%s): %v", topic, err)
				break
			}

			var nextU domain.UserSnapshot
			if err := json.Unmarshal(nextMsg.Value, &nextU); err == nil {
				batch = append(batch, userEvent{msg: nextMsg, user: nextU})
				msgs = append(msgs, nextMsg)
			} else {
				log.Printf("Kafka Unmarshal Error [user-registered]: %v", err)
				h.sendToDLQ(ctx, nextMsg, fmt.Sprintf("Unmarshal error: %v", err))
				_ = reader.CommitMessages(ctx, nextMsg)
			}
		}

		if len(batch) > 0 {
			h.flushBatch(ctx, reader, batch, msgs)
		}
	}
}

func (h *UserRegisteredHandler) flushBatch(ctx context.Context, reader *kafka.Reader, batch []userEvent, msgs []kafka.Message) {
	log.Printf("Flushing %d user snapshots to database in batch...", len(batch))
	
	dbCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	users := make([]domain.UserSnapshot, len(batch))
	for i, ev := range batch {
		users[i] = ev.user
	}

	failedUsers, err := h.userRepo.BulkUpsertUsers(dbCtx, users)
	if err != nil {
		log.Printf("Failed to bulk upsert users due to database connection error: %v. Offsets will not be committed.", err)
		return
	}

	if len(failedUsers) > 0 {
		log.Printf("Isolated %d poisonous user snapshots to DLQ.", len(failedUsers))
		for _, fu := range failedUsers {
			for _, ev := range batch {
				if ev.user.ID == fu.ID {
					h.sendToDLQ(ctx, ev.msg, "Database constraint violation during upsert")
					break
				}
			}
		}
	}

	log.Printf("Successfully processed %d user snapshot events!", len(batch))
	if err := reader.CommitMessages(ctx, msgs...); err != nil {
		log.Printf("Failed to commit Kafka offsets for user-registered batch: %v", err)
	}
}

func (h *UserRegisteredHandler) sendToDLQ(ctx context.Context, msg kafka.Message, reason string) {
	headers := append(msg.Headers, kafka.Header{
		Key:   "x-failure-reason",
		Value: []byte(reason),
	})

	dlqMsg := kafka.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
	}

	dlqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := h.dlqWriter.WriteMessages(dlqCtx, dlqMsg)
	if err != nil {
		log.Printf("[DLQ ERROR] Failed to write poisonous user registration message to DLQ: %v", err)
	} else {
		log.Printf("[DLQ] Poisonous user registration message isolated to DLQ topic: key=%s, reason=%s", string(msg.Key), reason)
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

	sessionData, err := h.rdb.Get(ctx, redisKey).Result()
	if err == nil && sessionData != "" {
		h.rdb.Del(ctx, redisKey)
		return
	}
}
