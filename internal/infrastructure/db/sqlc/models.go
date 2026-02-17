package sqlc

import "database/sql"

type Provider struct {
	ID        string
	Type      string
	Name      string
	AuthJSON  sql.NullString
	CreatedAt string
}

type Workspace struct {
	ID         string
	ProviderID string
	RemoteID   sql.NullString
	Name       string
}

type Board struct {
	ID          string
	WorkspaceID string
	RemoteID    sql.NullString
	Name        string
	ViewDefault string
}

type Column struct {
	ID       string
	BoardID  string
	RemoteID sql.NullString
	Name     string
	Position int64
	WipLimit sql.NullInt64
}

type Task struct {
	ID              string
	ProviderID      string
	WorkspaceID     string
	BoardID         sql.NullString
	ColumnID        sql.NullString
	RemoteID        sql.NullString
	Title           string
	DescriptionMd   string
	Status          sql.NullString
	Priority        int64
	DueAt           sql.NullString
	EstimateMinutes sql.NullInt64
	Assignee        sql.NullString
	LabelsJSON      string
	Position        float64
	CreatedAt       string
	UpdatedAt       string
}

type Comment struct {
	ID         string
	TaskID     string
	ProviderID string
	RemoteID   sql.NullString
	BodyMd     string
	Author     sql.NullString
	CreatedAt  string
}
