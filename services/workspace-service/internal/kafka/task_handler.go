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

type TaskUpdatedHandler struct {
	taskRepo domain.TaskRepository
	taskChan chan domain.Task
	mu       sync.Mutex
	batch    []domain.Task
}

func NewTaskUpdatedHandler(taskRepo domain.TaskRepository) *TaskUpdatedHandler {
	h := &TaskUpdatedHandler{
		taskRepo: taskRepo,
		taskChan: make(chan domain.Task, 20000), // Buffer size to absorb spikes
		batch:    make([]domain.Task, 0),
	}
	go h.startBatchWorker()
	return h
}

func (h *TaskUpdatedHandler) Handle(ctx context.Context, msg kafka.Message) {
	var t domain.Task
	if err := json.Unmarshal(msg.Value, &t); err != nil {
		log.Printf("Kafka Unmarshal Error [task-updated]: %v", err)
		return
	}
	h.taskChan <- t
}

func (h *TaskUpdatedHandler) startBatchWorker() {
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

func (h *TaskUpdatedHandler) flush() {
	h.mu.Lock()
	if len(h.batch) == 0 {
		h.mu.Unlock()
		return
	}
	currentBatch := h.batch
	h.batch = make([]domain.Task, 0)
	h.mu.Unlock()

	log.Printf("Flushing %d updated tasks to database in batch...", len(currentBatch))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.taskRepo.BulkUpdate(ctx, currentBatch); err != nil {
		log.Printf("Failed to bulk update tasks in database: %v", err)
	} else {
		log.Printf("Successfully flushed %d updated tasks in batch to database!", len(currentBatch))
	}
}

type TaskStatusUpdatedHandler struct {
	taskRepo domain.TaskRepository
	chanUpd  chan domain.TaskStatusUpdate
	mu       sync.Mutex
	batch    []domain.TaskStatusUpdate
}

func NewTaskStatusUpdatedHandler(taskRepo domain.TaskRepository) *TaskStatusUpdatedHandler {
	h := &TaskStatusUpdatedHandler{
		taskRepo: taskRepo,
		chanUpd:  make(chan domain.TaskStatusUpdate, 20000), // Buffer size to absorb spikes
		batch:    make([]domain.TaskStatusUpdate, 0),
	}
	go h.startBatchWorker()
	return h
}

func (h *TaskStatusUpdatedHandler) Handle(ctx context.Context, msg kafka.Message) {
	var u domain.TaskStatusUpdate
	if err := json.Unmarshal(msg.Value, &u); err != nil {
		log.Printf("Kafka Unmarshal Error [task-status-updated]: %v", err)
		return
	}
	h.chanUpd <- u
}

func (h *TaskStatusUpdatedHandler) startBatchWorker() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case u, ok := <-h.chanUpd:
			if !ok {
				h.flush()
				return
			}
			h.mu.Lock()
			h.batch = append(h.batch, u)
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

func (h *TaskStatusUpdatedHandler) flush() {
	h.mu.Lock()
	if len(h.batch) == 0 {
		h.mu.Unlock()
		return
	}
	currentBatch := h.batch
	h.batch = make([]domain.TaskStatusUpdate, 0)
	h.mu.Unlock()

	log.Printf("Flushing %d task status updates to database in batch...", len(currentBatch))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.taskRepo.BulkUpdateStatus(ctx, currentBatch); err != nil {
		log.Printf("Failed to bulk update task status in database: %v", err)
	} else {
		log.Printf("Successfully flushed %d task status updates in batch to database!", len(currentBatch))
	}
}
