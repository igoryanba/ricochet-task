package youtrack

import (
	"context"
	"fmt"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// YouTrackBoardProvider implements BoardProvider interface for YouTrack
type YouTrackBoardProvider struct {
	client *YouTrackClient
	config *providers.ProviderConfig
}

// NewYouTrackBoardProvider creates a new YouTrack board provider
func NewYouTrackBoardProvider(client *YouTrackClient, config *providers.ProviderConfig) *YouTrackBoardProvider {
	return &YouTrackBoardProvider{
		client: client,
		config: config,
	}
}

// GetBoard retrieves a specific board by ID
func (bp *YouTrackBoardProvider) GetBoard(ctx context.Context, id string) (*providers.UniversalBoard, error) {
	// YouTrack uses agile boards - get board info
	boardInfo, err := bp.client.GetAgileBoard(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	return &providers.UniversalBoard{
		ID:          boardInfo.ID,
		Name:        boardInfo.Name,
		ProjectID:   boardInfo.Projects[0].ID, // YouTrack boards can have multiple projects, take first
		Description: "YouTrack Agile Board",
		Type:        providers.BoardTypeScrum,
		ProviderName: bp.config.Name,
		CreatedAt:   time.Unix(boardInfo.CreatedAt, 0),
		UpdatedAt:   time.Unix(boardInfo.UpdatedAt, 0),
	}, nil
}

// ListBoards retrieves all boards for a project
func (bp *YouTrackBoardProvider) ListBoards(ctx context.Context, projectID string) ([]*providers.UniversalBoard, error) {
	// Get all agile boards from YouTrack
	boards, err := bp.client.ListAgileBoards(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list boards: %w", err)
	}

	var universalBoards []*providers.UniversalBoard
	for _, board := range boards {
		universalBoard := &providers.UniversalBoard{
			ID:           board.ID,
			Name:         board.Name,
			ProjectID:    projectID,
			Description:  "YouTrack Agile Board",
			Type:         providers.BoardTypeScrum,
			ProviderName: bp.config.Name,
			CreatedAt:    time.Unix(board.CreatedAt, 0),
			UpdatedAt:    time.Unix(board.UpdatedAt, 0),
		}
		universalBoards = append(universalBoards, universalBoard)
	}

	return universalBoards, nil
}

// CreateBoard creates a new board
func (bp *YouTrackBoardProvider) CreateBoard(ctx context.Context, board *providers.UniversalBoard) (*providers.UniversalBoard, error) {
	// YouTrack board creation
	createRequest := &YouTrackCreateBoardRequest{
		Name:      board.Name,
		ProjectID: board.ProjectID,
	}

	createdBoard, err := bp.client.CreateAgileBoard(ctx, createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create board: %w", err)
	}

	return &providers.UniversalBoard{
		ID:           createdBoard.ID,
		Name:         createdBoard.Name,
		ProjectID:    board.ProjectID,
		Description:  board.Description,
		Type:         providers.BoardTypeScrum,
		ProviderName: bp.config.Name,
		CreatedAt:    time.Unix(createdBoard.CreatedAt, 0),
		UpdatedAt:    time.Unix(createdBoard.UpdatedAt, 0),
	}, nil
}

// UpdateBoard updates an existing board
func (bp *YouTrackBoardProvider) UpdateBoard(ctx context.Context, id string, updates *providers.BoardUpdate) error {
	updateRequest := &YouTrackUpdateBoardRequest{}
	
	if updates.Name != nil {
		updateRequest.Name = *updates.Name
	}
	
	if updates.Description != nil {
		updateRequest.Description = *updates.Description
	}

	err := bp.client.UpdateAgileBoard(ctx, id, updateRequest)
	if err != nil {
		return fmt.Errorf("failed to update board: %w", err)
	}

	return nil
}

// DeleteBoard deletes a board
func (bp *YouTrackBoardProvider) DeleteBoard(ctx context.Context, id string) error {
	err := bp.client.DeleteAgileBoard(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete board: %w", err)
	}
	return nil
}

// GetBoardColumns retrieves columns for a board
func (bp *YouTrackBoardProvider) GetBoardColumns(ctx context.Context, boardID string) ([]*providers.BoardColumn, error) {
	columns, err := bp.client.GetBoardColumns(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get board columns: %w", err)
	}

	var universalColumns []*providers.BoardColumn
	for _, col := range columns {
		universalColumn := &providers.BoardColumn{
			ID:          col.ID,
			Name:        col.Name,
			Order:       col.Position,
			Description: "",
			Status: providers.TaskStatus{
				ID:   col.ID,
				Name: col.Name,
			},
		}
		universalColumns = append(universalColumns, universalColumn)
	}

	return universalColumns, nil
}

// MoveBetweenColumns moves a task between columns
func (bp *YouTrackBoardProvider) MoveBetweenColumns(ctx context.Context, taskID, fromColumn, toColumn string) error {
	err := bp.client.MoveTaskBetweenColumns(ctx, taskID, fromColumn, toColumn)
	if err != nil {
		return fmt.Errorf("failed to move task between columns: %w", err)
	}
	return nil
}

// GetWorkflowRules retrieves workflow rules for a board
func (bp *YouTrackBoardProvider) GetWorkflowRules(ctx context.Context, boardID string) ([]*providers.WorkflowRule, error) {
	// YouTrack workflow rules - simplified implementation
	return []*providers.WorkflowRule{}, nil
}

// CreateWorkflowRule creates a new workflow rule
func (bp *YouTrackBoardProvider) CreateWorkflowRule(ctx context.Context, rule *providers.WorkflowRule) error {
	// YouTrack workflow rule creation - simplified implementation
	return fmt.Errorf("workflow rule creation not yet implemented for YouTrack")
}

// YouTrack-specific types for API communication
type YouTrackCreateBoardRequest struct {
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
}

type YouTrackUpdateBoardRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type YouTrackBoardInfo struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	Projects  []YouTrackProjectInfo    `json:"projects"`
	CreatedAt int64                    `json:"created"`
	UpdatedAt int64                    `json:"updated"`
}

type YouTrackProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type YouTrackColumnInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}