package domain

type Column struct {
	ID       string
	BoardID  string
	RemoteID *string
	Name     string
	Position int
	WIPLimit *int
}
