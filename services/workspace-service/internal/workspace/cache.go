package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewWorkspaceCache(rdb *redis.Client) domain.WorkspaceCache {
	return &redisCache{
		rdb: rdb,
		ttl: 24 * time.Hour,
	}
}

func (c *redisCache) userOwnedKey(userID string) string {
	return fmt.Sprintf("user:%s:workspaces:owned", userID)
}

func (c *redisCache) userJoinedKey(userID string) string {
	return fmt.Sprintf("user:%s:workspaces:joined", userID)
}

func (c *redisCache) workspaceMetaKey(workspaceID string) string {
	return fmt.Sprintf("workspace:%s:meta", workspaceID)
}

func (c *redisCache) workspaceRolesKey(workspaceID string) string {
	return fmt.Sprintf("workspace:%s:roles", workspaceID)
}

func (c *redisCache) workspaceMembersKey(workspaceID string) string {
	return fmt.Sprintf("workspace:%s:members", workspaceID)
}

func (c *redisCache) AddOwnedWorkspaceID(ctx context.Context, userID string, workspaceID string) error {
	key := c.userOwnedKey(userID)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		pipe := c.rdb.Pipeline()
		pipe.ZRem(ctx, key, "__empty__")
		pipe.ZAdd(ctx, key, redis.Z{Score: 0, Member: workspaceID})
		_, err = pipe.Exec(ctx)
		return err
	}
	return nil
}

func (c *redisCache) AddJoinedWorkspaceID(ctx context.Context, userID string, workspaceID string) error {
	key := c.userJoinedKey(userID)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		pipe := c.rdb.Pipeline()
		pipe.ZRem(ctx, key, "__empty__")
		pipe.ZAdd(ctx, key, redis.Z{Score: 0, Member: workspaceID})
		_, err = pipe.Exec(ctx)
		return err
	}
	return nil
}

func (c *redisCache) SetWorkspaceMeta(ctx context.Context, ws *domain.Workspace) error {
	key := c.workspaceMetaKey(ws.ID)
	data := map[string]interface{}{
		"id":          ws.ID,
		"name":        ws.Name,
		"slug":        ws.Slug,
		"owner_id":    ws.OwnerID,
		"description": ws.Description,
		"created_at":  ws.CreatedAt.Format(time.RFC3339Nano),
	}

	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, key, data)
	pipe.Expire(ctx, key, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) GetWorkspaceMeta(ctx context.Context, workspaceID string) (*domain.Workspace, error) {
	key := c.workspaceMetaKey(workspaceID)
	res, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	createdAt, err := time.Parse(time.RFC3339Nano, res["created_at"])
	if err != nil {
		createdAt, _ = time.Parse(time.RFC3339, res["created_at"])
	}

	return &domain.Workspace{
		ID:          res["id"],
		Name:        res["name"],
		Slug:        res["slug"],
		OwnerID:     res["owner_id"],
		Description: res["description"],
		CreatedAt:   createdAt,
	}, nil
}

func (c *redisCache) GetWorkspaceMetas(ctx context.Context, workspaceIDs []string) ([]domain.Workspace, []string, error) {
	if len(workspaceIDs) == 0 {
		return nil, nil, nil
	}

	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(workspaceIDs))
	for i, id := range workspaceIDs {
		cmds[i] = pipe.HGetAll(ctx, c.workspaceMetaKey(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, nil, err
	}

	var workspaces []domain.Workspace
	var missingIDs []string

	for i, cmd := range cmds {
		res, err := cmd.Result()
		if err != nil || len(res) == 0 {
			missingIDs = append(missingIDs, workspaceIDs[i])
			continue
		}

		createdAt, err := time.Parse(time.RFC3339Nano, res["created_at"])
		if err != nil {
			createdAt, _ = time.Parse(time.RFC3339, res["created_at"])
		}

		workspaces = append(workspaces, domain.Workspace{
			ID:          res["id"],
			Name:        res["name"],
			Slug:        res["slug"],
			OwnerID:     res["owner_id"],
			Description: res["description"],
			CreatedAt:   createdAt,
		})
	}

	return workspaces, missingIDs, nil
}

func (c *redisCache) GetOwnedWorkspaceIDs(ctx context.Context, userID string, limit int, cursor string) ([]string, bool, error) {
	key := c.userOwnedKey(userID)
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

	var filtered []string
	for _, id := range ids {
		if id != "__empty__" {
			filtered = append(filtered, id)
		}
	}
	return filtered, true, nil
}

func (c *redisCache) GetJoinedWorkspaceIDs(ctx context.Context, userID string, limit int, cursor string) ([]string, bool, error) {
	key := c.userJoinedKey(userID)
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

	var filtered []string
	for _, id := range ids {
		if id != "__empty__" {
			filtered = append(filtered, id)
		}
	}
	return filtered, true, nil
}

func (c *redisCache) SetOwnedWorkspaceIDs(ctx context.Context, userID string, ids []string) error {
	key := c.userOwnedKey(userID)
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
	} else {
		pipe.ZAdd(ctx, key, redis.Z{Score: 0, Member: "__empty__"})
	}
	pipe.Expire(ctx, key, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) SetJoinedWorkspaceIDs(ctx context.Context, userID string, ids []string) error {
	key := c.userJoinedKey(userID)
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
	} else {
		pipe.ZAdd(ctx, key, redis.Z{Score: 0, Member: "__empty__"})
	}
	pipe.Expire(ctx, key, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) InvalidateOwnedWorkspaces(ctx context.Context, userID string) error {
	return c.rdb.Del(ctx, c.userOwnedKey(userID)).Err()
}

func (c *redisCache) InvalidateJoinedWorkspaces(ctx context.Context, userID string) error {
	return c.rdb.Del(ctx, c.userJoinedKey(userID)).Err()
}

func (c *redisCache) GetMembers(ctx context.Context, workspaceID string) ([]domain.WorkspaceMemberResponse, bool, error) {
	key := c.workspaceMembersKey(workspaceID)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}

	vals, err := c.rdb.HVals(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}

	var members []domain.WorkspaceMemberResponse
	for _, val := range vals {
		if val == "__empty__" {
			continue
		}
		var m domain.WorkspaceMemberResponse
		if err := json.Unmarshal([]byte(val), &m); err != nil {
			return nil, false, err
		}
		members = append(members, m)
	}

	return members, true, nil
}

func (c *redisCache) SetMembers(ctx context.Context, workspaceID string, members []domain.WorkspaceMemberResponse) error {
	key := c.workspaceMembersKey(workspaceID)
	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, key)

	if len(members) > 0 {
		for _, m := range members {
			data, err := json.Marshal(m)
			if err != nil {
				return err
			}
			pipe.HSet(ctx, key, m.UserID, string(data))
		}
	} else {
		pipe.HSet(ctx, key, "__empty__", "__empty__")
	}
	pipe.Expire(ctx, key, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) AddMember(ctx context.Context, workspaceID string, member domain.WorkspaceMemberResponse) error {
	key := c.workspaceMembersKey(workspaceID)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		data, err := json.Marshal(member)
		if err != nil {
			return err
		}
		pipe := c.rdb.Pipeline()
		pipe.HDel(ctx, key, "__empty__")
		pipe.HSet(ctx, key, member.UserID, string(data))
		_, err = pipe.Exec(ctx)
		return err
	}
	return nil
}

func (c *redisCache) InvalidateMembers(ctx context.Context, workspaceID string) error {
	return c.rdb.Del(ctx, c.workspaceMembersKey(workspaceID)).Err()
}

func (c *redisCache) GetMemberRole(ctx context.Context, workspaceID string, userID string) (string, bool, error) {
	key := c.workspaceRolesKey(workspaceID)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return "", false, err
	}
	if exists == 0 {
		return "", false, nil
	}

	res, err := c.rdb.HGet(ctx, key, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", true, nil
		}
		return "", false, err
	}
	return res, true, nil
}

func (c *redisCache) SetMemberRole(ctx context.Context, workspaceID string, userID string, role string) error {
	key := c.workspaceRolesKey(workspaceID)
	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, key, userID, role)
	pipe.Expire(ctx, key, c.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *redisCache) InvalidateMemberRole(ctx context.Context, workspaceID string, userID string) error {
	key := c.workspaceRolesKey(workspaceID)
	return c.rdb.HDel(ctx, key, userID).Err()
}
