package domain

type Workspace struct {
	ID         string
	ProviderID string
	RemoteID   *string
	Name       string
}

type Board struct {
	ID          string
	WorkspaceID string
	RemoteID    *string
	Name        string
	ViewDefault string
}
