package task

import (
	"context"
	"fmt"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewTaskCache(rdb *redis.Client) domain.TaskCache {
	return &redisCache{
		rdb: rdb,
		ttl: 3 * 24 * time.Hour, // 3 days TTL
	}
}

func (c *redisCache) zsetKey(projectID string, columnName string) string {
	return fmt.Sprintf("project:%s:col:%s", projectID, columnName)
}

func (c *redisCache) hashKey(taskID string) string {
	return fmt.Sprintf("task:%s:data", taskID)
}

func (c *redisCache) getUUIDv7Score(taskID string) (float64, error) {
	u, err := uuid.Parse(taskID)
	if err != nil {
		return 0, err
	}
	if u.Version() != 7 {
		return 0, fmt.Errorf("task ID %s is not a UUIDv7", taskID)
	}
	// Extract 48-bit millisecond timestamp from bytes [0..5]
	score := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 | int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
	return float64(score), nil
}

func (c *redisCache) serializeTask(t *domain.Task) map[string]interface{} {
	assigneeID := ""
	if t.AssigneeID != "" {
		assigneeID = t.AssigneeID
	}
	assigneeName := ""
	if t.AssigneeName != nil {
		assigneeName = *t.AssigneeName
	}

	return map[string]interface{}{
		"id":            t.ID,
		"workspace_id":  t.WorkspaceID,
		"project_id":    t.ProjectID,
		"title":         t.Title,
		"description":   t.Description,
		"status":        t.Status,
		"priority":      t.Priority,
		"assignee_id":   assigneeID,
		"assignee_name": assigneeName,
		"deadline":      t.Deadline.Format(time.RFC3339Nano),
		"created_at":    t.CreatedAt.Format(time.RFC3339Nano),
	}
}

func (c *redisCache) deserializeTask(res map[string]string) domain.Task {
	var deadline time.Time
	if res["deadline"] != "" {
		var err error
		deadline, err = time.Parse(time.RFC3339Nano, res["deadline"])
		if err != nil {
			deadline, _ = time.Parse(time.RFC3339, res["deadline"])
		}
	}
	var createdAt time.Time
	if res["created_at"] != "" {
		var err error
		createdAt, err = time.Parse(time.RFC3339Nano, res["created_at"])
		if err != nil {
			createdAt, _ = time.Parse(time.RFC3339, res["created_at"])
		}
	}

	var assigneeName *string
	if name, exists := res["assignee_name"]; exists && name != "" {
		assigneeName = &name
	}

	return domain.Task{
		ID:           res["id"],
		WorkspaceID:  res["workspace_id"],
		ProjectID:    res["project_id"],
		Title:        res["title"],
		Description:  res["description"],
		Status:       res["status"],
		Priority:     res["priority"],
		AssigneeID:   res["assignee_id"],
		AssigneeName: assigneeName,
		Deadline:     deadline,
		CreatedAt:    createdAt,
	}
}

func (c *redisCache) AddTask(ctx context.Context, t *domain.Task) error {
	score, err := c.getUUIDv7Score(t.ID)
	if err != nil {
		return err
	}

	zkey := c.zsetKey(t.ProjectID, t.Status)
	hkey := c.hashKey(t.ID)

	exists, err := c.rdb.Exists(ctx, zkey).Result()
	if err != nil {
		return err
	}

	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, hkey, c.serializeTask(t))
	pipe.Expire(ctx, hkey, c.ttl)

	if exists > 0 {
		pipe.ZRem(ctx, zkey, "__empty__")
		pipe.ZAdd(ctx, zkey, redis.Z{Score: score, Member: t.ID})
		pipe.ZRemRangeByRank(ctx, zkey, 0, -101) // Keep only latest 100 items
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (c *redisCache) GetTaskIDs(ctx context.Context, projectID string, status string, limit int, cursor string) ([]string, []float64, bool, error) {
	zkey := c.zsetKey(projectID, status)
	exists, err := c.rdb.Exists(ctx, zkey).Result()
	if err != nil {
		return nil, nil, false, err
	}
	if exists == 0 {
		return nil, nil, false, nil
	}

	maxScore := "+inf"
	if cursor != "" {
		score, err := c.rdb.ZScore(ctx, zkey, cursor).Result()
		if err != nil {
			if err == redis.Nil {
				s, extErr := c.getUUIDv7Score(cursor)
				if extErr == nil {
					maxScore = fmt.Sprintf("(%f", s)
				}
			} else {
				return nil, nil, false, err
			}
		} else {
			maxScore = fmt.Sprintf("(%f", score)
		}
	}

	opt := redis.ZRangeBy{
		Min:    "-inf",
		Max:    maxScore,
		Offset: 0,
		Count:  int64(limit),
	}
	zList, err := c.rdb.ZRevRangeByScoreWithScores(ctx, zkey, &opt).Result()
	if err != nil {
		return nil, nil, false, err
	}

	taskIDs := make([]string, 0, len(zList))
	scores := make([]float64, 0, len(zList))
	for _, item := range zList {
		id := item.Member.(string)
		if id == "__empty__" {
			continue
		}
		taskIDs = append(taskIDs, id)
		scores = append(scores, item.Score)
	}

	return taskIDs, scores, true, nil
}

func (c *redisCache) GetTaskMetas(ctx context.Context, taskIDs []string) ([]domain.Task, []string, error) {
	if len(taskIDs) == 0 {
		return nil, nil, nil
	}

	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(taskIDs))
	for i, id := range taskIDs {
		cmds[i] = pipe.HGetAll(ctx, c.hashKey(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, nil, err
	}

	var tasks []domain.Task
	var missingIDs []string

	for i, cmd := range cmds {
		res, err := cmd.Result()
		if err != nil || len(res) == 0 || res["id"] == "" {
			missingIDs = append(missingIDs, taskIDs[i])
			continue
		}

		tasks = append(tasks, c.deserializeTask(res))
	}

	return tasks, missingIDs, nil
}

func (c *redisCache) SetTaskMeta(ctx context.Context, t *domain.Task) error {
	hkey := c.hashKey(t.ID)
	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, hkey, c.serializeTask(t))
	pipe.Expire(ctx, hkey, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) SetColumnCache(ctx context.Context, projectID string, status string, tasks []domain.Task) error {
	zkey := c.zsetKey(projectID, status)

	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, zkey)

	if len(tasks) > 0 {
		zMembers := make([]redis.Z, 0, len(tasks))
		for _, t := range tasks {
			score, err := c.getUUIDv7Score(t.ID)
			if err != nil {
				continue
			}
			zMembers = append(zMembers, redis.Z{
				Score:  score,
				Member: t.ID,
			})

			hkey := c.hashKey(t.ID)
			pipe.HSet(ctx, hkey, c.serializeTask(&t))
			pipe.Expire(ctx, hkey, c.ttl)
		}

		if len(zMembers) > 0 {
			pipe.ZAdd(ctx, zkey, zMembers...)
			pipe.Expire(ctx, zkey, c.ttl)
		}
	} else {
		pipe.ZAdd(ctx, zkey, redis.Z{Score: -1, Member: "__empty__"})
		pipe.Expire(ctx, zkey, c.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) UpdateTaskMeta(ctx context.Context, t *domain.Task) error {
	hkey := c.hashKey(t.ID)
	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, hkey, c.serializeTask(t))
	pipe.Expire(ctx, hkey, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) UpdateTaskStatus(ctx context.Context, projectID string, taskID string, oldStatus string, newStatus string) error {
	score, err := c.getUUIDv7Score(taskID)
	if err != nil {
		return err
	}

	oldZKey := c.zsetKey(projectID, oldStatus)
	newZKey := c.zsetKey(projectID, newStatus)
	hkey := c.hashKey(taskID)

	pipeCheck := c.rdb.Pipeline()
	newExistsCmd := pipeCheck.Exists(ctx, newZKey)
	_, _ = pipeCheck.Exec(ctx)
	newExists := newExistsCmd.Val() > 0

	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, hkey, "status", newStatus)
	pipe.Expire(ctx, hkey, c.ttl)

	pipe.ZRem(ctx, oldZKey, taskID)

	if newExists {
		pipe.ZRem(ctx, newZKey, "__empty__")
		pipe.ZAdd(ctx, newZKey, redis.Z{Score: score, Member: taskID})
		pipe.ZRemRangeByRank(ctx, newZKey, 0, -101)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (c *redisCache) InvalidateTasks(ctx context.Context, projectID string) error {
	statuses := []string{"TODO", "IN_PROGRESS", "REVIEW", "DONE"}
	var keys []string
	for _, st := range statuses {
		keys = append(keys, c.zsetKey(projectID, st))
	}
	return c.rdb.Del(ctx, keys...).Err()
}
