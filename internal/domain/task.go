package domain

import "time"

type Task struct {
	ID              string
	ProviderID      string
	WorkspaceID     string
	BoardID         *string
	ColumnID        *string
	RemoteID        *string
	Title           string
	DescriptionMD   string
	Status          *string
	Priority        int
	DueAt           *time.Time
	EstimateMinutes *int
	Assignee        *string
	Labels          []string
	Position        float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TaskPatch struct {
	Title         *string
	DescriptionMD *string
	Status        *string
	Priority      *int
	DueAt         *time.Time
	ColumnID      *string
	Labels        *[]string
}

type TaskFilter struct {
	WorkspaceID string
	BoardID     string
	TitleQuery  string
	ColumnID    string
	Status      string
	DueSoonBy   *time.Time
}

type MoveTaskInput struct {
	TaskID    string
	ColumnID  *string
	Status    *string
	Position  float64
	UpdatedAt time.Time
}
