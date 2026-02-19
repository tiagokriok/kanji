package providers

import "github.com/tiagokriok/kanji/internal/domain"

type LocalProvider struct{}

func NewLocalProvider() domain.ProviderClient {
	return LocalProvider{}
}

func (LocalProvider) Type() string {
	return "local"
}

func (LocalProvider) Name() string {
	return "Local"
}
