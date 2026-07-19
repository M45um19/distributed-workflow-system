package project

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sort"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/google/uuid"
)

type service struct {
	projectRepo domain.ProjectRepository
	wsRepo      domain.WorkspaceRepository
	cache       domain.ProjectCache
	wsCache     domain.WorkspaceCache
}

func NewService(
	projectRepo domain.ProjectRepository,
	wsRepo domain.WorkspaceRepository,
	cache domain.ProjectCache,
	wsCache domain.WorkspaceCache,
) domain.ProjectService {
	return &service{
		projectRepo: projectRepo,
		wsRepo:      wsRepo,
		cache:       cache,
		wsCache:     wsCache,
	}
}

func (s *service) getWorkspace(ctx context.Context, workspaceID string) (*domain.Workspace, error) {
	ws, err := s.wsCache.GetWorkspaceMeta(ctx, workspaceID)
	if err != nil {
		log.Printf("[Cache Error] GetWorkspaceMeta failed: %v", err)
	}
	if ws != nil {
		return ws, nil
	}

	ws, err = s.wsRepo.FindByID(ctx, workspaceID, workspaceID)
	if err != nil {
		return nil, err
	}

	if ws != nil {
		if cacheErr := s.wsCache.SetWorkspaceMeta(ctx, ws); cacheErr != nil {
			log.Printf("[Cache Error] SetWorkspaceMeta failed: %v", cacheErr)
		}
	}

	return ws, nil
}

func (s *service) getMemberRole(ctx context.Context, workspaceID string, userID string, ownerID string) (string, error) {
	if ownerID == userID {
		return "OWNER", nil
	}

	role, cached, err := s.wsCache.GetMemberRole(ctx, workspaceID, userID)
	if err != nil {
		log.Printf("[Cache Error] GetMemberRole failed: %v", err)
	}

	if !cached {
		role, err = s.wsRepo.GetMemberRole(ctx, workspaceID, userID)
		if err != nil {
			return "", err
		}
		if cacheErr := s.wsCache.SetMemberRole(ctx, workspaceID, userID, role); cacheErr != nil {
			log.Printf("[Cache Error] SetMemberRole failed: %v", cacheErr)
		}
	}

	return role, nil
}

func (s *service) CreateProject(ctx context.Context, workspaceID string, input domain.ProjectCreateInput, userID string) (*domain.Project, error) {
	ws, err := s.getWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	if ws.OwnerID != userID {
		role, err := s.getMemberRole(ctx, workspaceID, userID, ws.OwnerID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, apperror.Forbidden("You are not a member of this workspace")
			}
			return nil, err
		}
		if role != "ADMIN" {
			return nil, apperror.Forbidden("Only workspace owners and admins can create projects")
		}
	}

	projID, err := uuid.NewV7()
	if err != nil {
		return nil, apperror.InternalServer("failed to generate project ID: " + err.Error())
	}

	p := &domain.Project{
		ID:          projID.String(),
		WorkspaceID: workspaceID,
		Name:        input.Name,
		Description: input.Description,
		Status:      "ACTIVE",
		CreatedBy:   userID,
	}

	if err := s.projectRepo.Create(ctx, workspaceID, p); err != nil {
		return nil, err
	}

	// Update cache
	if cacheErr := s.cache.SetProjectMeta(ctx, p); cacheErr != nil {
		log.Printf("[Cache Error] CreateProject SetProjectMeta failed: %v", cacheErr)
	}
	if cacheErr := s.cache.AddProjectID(ctx, workspaceID, p.ID); cacheErr != nil {
		log.Printf("[Cache Error] CreateProject AddProjectID failed: %v", cacheErr)
	}

	return p, nil
}

func (s *service) GetProjectsByWorkspace(ctx context.Context, workspaceID string, userID string, limit int, cursor string) ([]domain.Project, string, error) {
	ws, err := s.getWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, "", err
	}
	if ws == nil {
		return nil, "", apperror.NotFound("Workspace not found")
	}

	if ws.OwnerID != userID {
		role, err := s.getMemberRole(ctx, workspaceID, userID, ws.OwnerID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, "", apperror.Forbidden("You do not have access to this workspace")
			}
			return nil, "", err
		}
		if role == "" {
			return nil, "", apperror.Forbidden("You do not have access to this workspace")
		}
	}

	// Query Project IDs from cache
	ids, cacheExists, err := s.cache.GetProjectIDs(ctx, workspaceID, limit, cursor)
	if err != nil {
		log.Printf("[Cache Error] GetProjectIDs failed: %v", err)
	}

	var projects []domain.Project

	if !cacheExists {
		// Cache Miss: Query DB for projects (fetch a bulk of 1000 items to populate cache)
		allDBProjects, err := s.projectRepo.GetByWorkspaceID(ctx, workspaceID, 1000, "")
		if err != nil {
			return nil, "", err
		}

		var dbIDs []string
		for _, p := range allDBProjects {
			dbIDs = append(dbIDs, p.ID)
			if cacheErr := s.cache.SetProjectMeta(ctx, &p); cacheErr != nil {
				log.Printf("[Cache Error] SetProjectMeta failed: %v", cacheErr)
			}
		}

		if cacheErr := s.cache.SetProjectIDs(ctx, workspaceID, dbIDs); cacheErr != nil {
			log.Printf("[Cache Error] SetProjectIDs failed: %v", cacheErr)
		}

		projects = allDBProjects
		// Slice in-memory
		projects = paginateProjectsInGo(projects, cursor, limit)
	} else {
		if len(ids) == 0 {
			return []domain.Project{}, "", nil
		}

		// Cache Hit: fetch meta using Redis pipelining
		var missingIDs []string
		projects, missingIDs, err = s.cache.GetProjectMetas(ctx, ids)
		if err != nil {
			log.Printf("[Cache Error] GetProjectMetas failed: %v", err)
		}

		// Handle any evicted metadata hashes
		if len(missingIDs) > 0 {
			for _, id := range missingIDs {
				p, err := s.projectRepo.GetByID(ctx, id)
				if err != nil {
					return nil, "", err
				}
				if p != nil {
					projects = append(projects, *p)
					if cacheErr := s.cache.SetProjectMeta(ctx, p); cacheErr != nil {
						log.Printf("[Cache Error] SetProjectMeta failed: %v", cacheErr)
					}
				}
			}
		}

		// Sort projects descending
		sortProjectsDesc(projects)
	}

	nextCursor := ""
	if len(projects) > 0 {
		nextCursor = projects[len(projects)-1].ID
	}

	return projects, nextCursor, nil
}

func sortProjectsDesc(projects []domain.Project) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ID > projects[j].ID
	})
}

func paginateProjectsInGo(projects []domain.Project, cursor string, limit int) []domain.Project {
	var filtered []domain.Project
	for _, p := range projects {
		if cursor == "" || p.ID < cursor {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) > limit {
		return filtered[:limit]
	}
	return filtered
}
