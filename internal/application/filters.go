package application

import "time"

type ListTaskFilters struct {
	WorkspaceID string
	BoardID     string
	TitleQuery  string
	ColumnID    string
	Status      string
	DueSoonDays int
}

func (f ListTaskFilters) DueSoonBy(now time.Time) *time.Time {
	if f.DueSoonDays <= 0 {
		return nil
	}
	v := now.AddDate(0, 0, f.DueSoonDays)
	return &v
}
