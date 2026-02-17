package domain

import "time"

type Provider struct {
	ID        string
	Type      string
	Name      string
	AuthJSON  *string
	CreatedAt time.Time
}

// ProviderClient defines the provider boundary for future external integrations.
type ProviderClient interface {
	Type() string
	Name() string
}
