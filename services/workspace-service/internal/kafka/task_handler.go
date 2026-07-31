package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type TaskCreatedHandler struct {
	taskRepo  domain.TaskRepository
	dlqWriter *kafka.Writer
}

func NewTaskCreatedHandler(taskRepo domain.TaskRepository, dlqWriter *kafka.Writer) *TaskCreatedHandler {
	return &TaskCreatedHandler{
		taskRepo:  taskRepo,
		dlqWriter: dlqWriter,
	}
}

func (h *TaskCreatedHandler) Handle(ctx context.Context, msg kafka.Message) {
	// Satisfies KafkaEventHandler interface, but worker uses RunLoop (KafkaLoopHandler)
}

func (h *TaskCreatedHandler) RunLoop(ctx context.Context, reader *kafka.Reader) {
	topic := reader.Config().Topic
	batchSize := 1000
	flushInterval := 2 * time.Second

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

		var batch []domain.Task
		var msgs []kafka.Message

		var t domain.Task
		if err := json.Unmarshal(msg.Value, &t); err == nil {
			batch = append(batch, t)
			msgs = append(msgs, msg)
		} else {
			log.Printf("Kafka Unmarshal Error [task-created]: %v", err)
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

			var nextT domain.Task
			if err := json.Unmarshal(nextMsg.Value, &nextT); err == nil {
				batch = append(batch, nextT)
				msgs = append(msgs, nextMsg)
			} else {
				log.Printf("Kafka Unmarshal Error [task-created]: %v", err)
				_ = reader.CommitMessages(ctx, nextMsg)
			}
		}

		if len(batch) > 0 {
			h.flushBatch(ctx, reader, batch, msgs)
		}
	}
}

func (h *TaskCreatedHandler) flushBatch(ctx context.Context, reader *kafka.Reader, batch []domain.Task, msgs []kafka.Message) {
	log.Printf("Flushing %d tasks to database in batch...", len(batch))
	
	dbCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	failedTasks, err := h.taskRepo.BulkCreate(dbCtx, batch)
	if err != nil {
		log.Printf("Failed to bulk create tasks in database due to connection error: %v. Offsets will not be committed.", err)
		return
	}

	if len(failedTasks) > 0 {
		log.Printf("Isolated %d poisonous tasks to DLQ.", len(failedTasks))
		for _, ft := range failedTasks {
			h.handleDLQ(ctx, ft)
		}
	}

	log.Printf("Successfully processed %d tasks (succeeded: %d, failed/DLQ: %d)!", len(batch), len(batch)-len(failedTasks), len(failedTasks))
	if err := reader.CommitMessages(ctx, msgs...); err != nil {
		log.Printf("Failed to commit Kafka offsets for task-created batch: %v", err)
	}
}

func (h *TaskCreatedHandler) handleDLQ(ctx context.Context, t domain.Task) {
	taskJSON, err := json.Marshal(t)
	if err != nil {
		log.Printf("[DLQ ERROR] Failed to marshal poisonous task creation: %v", err)
		return
	}

	// Publish to Kafka DLQ writer (non-blocking for main loop, handled gracefully if fails)
	dlqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = h.dlqWriter.WriteMessages(dlqCtx, kafka.Message{
		Key:   []byte(t.ID),
		Value: taskJSON,
	})
	if err != nil {
		log.Printf("[DLQ ERROR] Failed to write poisonous task %s to DLQ topic: %v", t.ID, err)
	} else {
		log.Printf("[DLQ] Poisonous task creation event isolated to DLQ topic: %s", t.ID)
	}
}

type TaskUpdatedHandler struct {
	taskRepo  domain.TaskRepository
	dlqWriter *kafka.Writer
}

func NewTaskUpdatedHandler(taskRepo domain.TaskRepository, dlqWriter *kafka.Writer) *TaskUpdatedHandler {
	return &TaskUpdatedHandler{
		taskRepo:  taskRepo,
		dlqWriter: dlqWriter,
	}
}

func (h *TaskUpdatedHandler) Handle(ctx context.Context, msg kafka.Message) {
	// Satisfies KafkaEventHandler interface, but worker uses RunLoop (KafkaLoopHandler)
}

func (h *TaskUpdatedHandler) RunLoop(ctx context.Context, reader *kafka.Reader) {
	topic := reader.Config().Topic
	batchSize := 1000
	flushInterval := 2 * time.Second

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

		var batch []domain.Task
		var msgs []kafka.Message

		var t domain.Task
		if err := json.Unmarshal(msg.Value, &t); err == nil {
			batch = append(batch, t)
			msgs = append(msgs, msg)
		} else {
			log.Printf("Kafka Unmarshal Error [task-updated]: %v", err)
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

			var nextT domain.Task
			if err := json.Unmarshal(nextMsg.Value, &nextT); err == nil {
				batch = append(batch, nextT)
				msgs = append(msgs, nextMsg)
			} else {
				log.Printf("Kafka Unmarshal Error [task-updated]: %v", err)
				_ = reader.CommitMessages(ctx, nextMsg)
			}
		}

		if len(batch) > 0 {
			h.flushBatch(ctx, reader, batch, msgs)
		}
	}
}

func (h *TaskUpdatedHandler) flushBatch(ctx context.Context, reader *kafka.Reader, batch []domain.Task, msgs []kafka.Message) {
	log.Printf("Flushing %d updated tasks to database in batch...", len(batch))
	
	dbCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	failedTasks, err := h.taskRepo.BulkUpdate(dbCtx, batch)
	if err != nil {
		log.Printf("Failed to bulk update tasks in database due to connection error: %v. Offsets will not be committed.", err)
		return
	}

	if len(failedTasks) > 0 {
		log.Printf("Isolated %d poisonous task updates to DLQ.", len(failedTasks))
		for _, ft := range failedTasks {
			h.handleDLQ(ctx, ft)
		}
	}

	log.Printf("Successfully processed %d updated tasks (succeeded: %d, failed/DLQ: %d)!", len(batch), len(batch)-len(failedTasks), len(failedTasks))
	if err := reader.CommitMessages(ctx, msgs...); err != nil {
		log.Printf("Failed to commit Kafka offsets for task-updated batch: %v", err)
	}
}

func (h *TaskUpdatedHandler) handleDLQ(ctx context.Context, t domain.Task) {
	taskJSON, err := json.Marshal(t)
	if err != nil {
		log.Printf("[DLQ ERROR] Failed to marshal poisonous task update: %v", err)
		return
	}

	dlqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = h.dlqWriter.WriteMessages(dlqCtx, kafka.Message{
		Key:   []byte(t.ID),
		Value: taskJSON,
	})
	if err != nil {
		log.Printf("[DLQ ERROR] Failed to write poisonous task update %s to DLQ topic: %v", t.ID, err)
	} else {
		log.Printf("[DLQ] Poisonous task update event isolated to DLQ topic: %s", t.ID)
	}
}

type TaskStatusUpdatedHandler struct {
	taskRepo  domain.TaskRepository
	dlqWriter *kafka.Writer
}

func NewTaskStatusUpdatedHandler(taskRepo domain.TaskRepository, dlqWriter *kafka.Writer) *TaskStatusUpdatedHandler {
	return &TaskStatusUpdatedHandler{
		taskRepo:  taskRepo,
		dlqWriter: dlqWriter,
	}
}

func (h *TaskStatusUpdatedHandler) Handle(ctx context.Context, msg kafka.Message) {
	// Satisfies KafkaEventHandler interface, but worker uses RunLoop (KafkaLoopHandler)
}

func (h *TaskStatusUpdatedHandler) RunLoop(ctx context.Context, reader *kafka.Reader) {
	topic := reader.Config().Topic
	batchSize := 1000
	flushInterval := 2 * time.Second

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

		var batch []domain.TaskStatusUpdate
		var msgs []kafka.Message

		var u domain.TaskStatusUpdate
		if err := json.Unmarshal(msg.Value, &u); err == nil {
			batch = append(batch, u)
			msgs = append(msgs, msg)
		} else {
			log.Printf("Kafka Unmarshal Error [task-status-updated]: %v", err)
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

			var nextU domain.TaskStatusUpdate
			if err := json.Unmarshal(nextMsg.Value, &nextU); err == nil {
				batch = append(batch, nextU)
				msgs = append(msgs, nextMsg)
			} else {
				log.Printf("Kafka Unmarshal Error [task-status-updated]: %v", err)
				_ = reader.CommitMessages(ctx, nextMsg)
			}
		}

		if len(batch) > 0 {
			h.flushBatch(ctx, reader, batch, msgs)
		}
	}
}

func (h *TaskStatusUpdatedHandler) flushBatch(ctx context.Context, reader *kafka.Reader, batch []domain.TaskStatusUpdate, msgs []kafka.Message) {
	log.Printf("Flushing %d task status updates to database in batch...", len(batch))
	
	dbCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	failedUpdates, err := h.taskRepo.BulkUpdateStatus(dbCtx, batch)
	if err != nil {
		log.Printf("Failed to bulk update task status in database due to connection error: %v. Offsets will not be committed.", err)
		return
	}

	if len(failedUpdates) > 0 {
		log.Printf("Isolated %d poisonous task status updates to DLQ.", len(failedUpdates))
		for _, fu := range failedUpdates {
			h.handleDLQ(ctx, fu)
		}
	}

	log.Printf("Successfully processed %d task status updates (succeeded: %d, failed/DLQ: %d)!", len(batch), len(batch)-len(failedUpdates), len(failedUpdates))
	if err := reader.CommitMessages(ctx, msgs...); err != nil {
		log.Printf("Failed to commit Kafka offsets for task-status-updated batch: %v", err)
	}
}

func (h *TaskStatusUpdatedHandler) handleDLQ(ctx context.Context, u domain.TaskStatusUpdate) {
	updateJSON, err := json.Marshal(u)
	if err != nil {
		log.Printf("[DLQ ERROR] Failed to marshal poisonous task status update: %v", err)
		return
	}

	dlqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = h.dlqWriter.WriteMessages(dlqCtx, kafka.Message{
		Key:   []byte(u.TaskID),
		Value: updateJSON,
	})
	if err != nil {
		log.Printf("[DLQ ERROR] Failed to write poisonous task status update %s to DLQ topic: %v", u.TaskID, err)
	} else {
		log.Printf("[DLQ] Poisonous task status update event isolated to DLQ topic: %s", u.TaskID)
	}
}
