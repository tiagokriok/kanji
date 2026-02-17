package domain

import "time"

type Comment struct {
	ID         string
	TaskID     string
	ProviderID string
	RemoteID   *string
	BodyMD     string
	Author     *string
	CreatedAt  time.Time
}
