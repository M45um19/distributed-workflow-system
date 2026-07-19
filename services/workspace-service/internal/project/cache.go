package project

import (
	"context"
	"fmt"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewProjectCache(rdb *redis.Client) domain.ProjectCache {
	return &redisCache{
		rdb: rdb,
		ttl: 24 * time.Hour,
	}
}

func (c *redisCache) projectListKey(workspaceID string) string {
	return fmt.Sprintf("workspace:%s:projects", workspaceID)
}

func (c *redisCache) projectMetaKey(projectID string) string {
	return fmt.Sprintf("project:%s:meta", projectID)
}

func (c *redisCache) AddProjectID(ctx context.Context, workspaceID string, projectID string) error {
	key := c.projectListKey(workspaceID)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return c.rdb.ZAdd(ctx, key, redis.Z{Score: 0, Member: projectID}).Err()
	}
	return nil
}

func (c *redisCache) SetProjectMeta(ctx context.Context, p *domain.Project) error {
	key := c.projectMetaKey(p.ID)
	data := map[string]interface{}{
		"id":           p.ID,
		"workspace_id": p.WorkspaceID,
		"name":         p.Name,
		"description":  p.Description,
		"status":       p.Status,
		"created_by":   p.CreatedBy,
		"created_at":   p.CreatedAt.Format(time.RFC3339Nano),
	}

	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, key, data)
	pipe.Expire(ctx, key, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) GetProjectMetas(ctx context.Context, projectIDs []string) ([]domain.Project, []string, error) {
	if len(projectIDs) == 0 {
		return nil, nil, nil
	}

	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(projectIDs))
	for i, id := range projectIDs {
		cmds[i] = pipe.HGetAll(ctx, c.projectMetaKey(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, nil, err
	}

	var projects []domain.Project
	var missingIDs []string

	for i, cmd := range cmds {
		res, err := cmd.Result()
		if err != nil || len(res) == 0 {
			missingIDs = append(missingIDs, projectIDs[i])
			continue
		}

		createdAt, parseErr := time.Parse(time.RFC3339Nano, res["created_at"])
		if parseErr != nil {
			createdAt, _ = time.Parse(time.RFC3339, res["created_at"])
		}

		projects = append(projects, domain.Project{
			ID:          res["id"],
			WorkspaceID: res["workspace_id"],
			Name:        res["name"],
			Description: res["description"],
			Status:      res["status"],
			CreatedBy:   res["created_by"],
			CreatedAt:   createdAt,
		})
	}

	return projects, missingIDs, nil
}

func (c *redisCache) GetProjectIDs(ctx context.Context, workspaceID string, limit int, cursor string) ([]string, bool, error) {
	key := c.projectListKey(workspaceID)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}

	max := "+"
	if cursor != "" {
		max = "(" + cursor
	}

	opt := &redis.ZRangeBy{
		Min:    "-",
		Max:    max,
		Offset: 0,
		Count:  int64(limit),
	}

	ids, err := c.rdb.ZRevRangeByLex(ctx, key, opt).Result()
	if err != nil {
		return nil, false, err
	}

	return ids, true, nil
}

func (c *redisCache) SetProjectIDs(ctx context.Context, workspaceID string, ids []string) error {
	key := c.projectListKey(workspaceID)
	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, key)
	if len(ids) > 0 {
		zMembers := make([]redis.Z, len(ids))
		for i, id := range ids {
			zMembers[i] = redis.Z{
				Score:  0,
				Member: id,
			}
		}
		pipe.ZAdd(ctx, key, zMembers...)
		pipe.Expire(ctx, key, c.ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) InvalidateProjects(ctx context.Context, workspaceID string) error {
	return c.rdb.Del(ctx, c.projectListKey(workspaceID)).Err()
}
