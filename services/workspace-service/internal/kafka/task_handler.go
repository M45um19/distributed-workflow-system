package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type TaskCreatedHandler struct {
	taskRepo domain.TaskRepository
	taskChan chan domain.Task
	mu       sync.Mutex
	batch    []domain.Task
}

func NewTaskCreatedHandler(taskRepo domain.TaskRepository) *TaskCreatedHandler {
	h := &TaskCreatedHandler{
		taskRepo: taskRepo,
		taskChan: make(chan domain.Task, 20000), // Buffer size to absorb spikes
		batch:    make([]domain.Task, 0),
	}
	go h.startBatchWorker()
	return h
}

func (h *TaskCreatedHandler) Handle(ctx context.Context, msg kafka.Message) {
	var t domain.Task
	if err := json.Unmarshal(msg.Value, &t); err != nil {
		log.Printf("Kafka Unmarshal Error [task-created]: %v", err)
		return
	}
	h.taskChan <- t
}

func (h *TaskCreatedHandler) startBatchWorker() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t, ok := <-h.taskChan:
			if !ok {
				h.flush()
				return
			}
			h.mu.Lock()
			h.batch = append(h.batch, t)
			batchLen := len(h.batch)
			h.mu.Unlock()

			if batchLen >= 1000 {
				h.flush()
			}

		case <-ticker.C:
			h.flush()
		}
	}
}

func (h *TaskCreatedHandler) flush() {
	h.mu.Lock()
	if len(h.batch) == 0 {
		h.mu.Unlock()
		return
	}
	currentBatch := h.batch
	h.batch = make([]domain.Task, 0)
	h.mu.Unlock()

	log.Printf("Flushing %d tasks to database in batch...", len(currentBatch))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.taskRepo.BulkCreate(ctx, currentBatch); err != nil {
		log.Printf("Failed to bulk create tasks in database: %v", err)
	} else {
		log.Printf("Successfully flushed %d tasks in batch to database!", len(currentBatch))
	}
}
