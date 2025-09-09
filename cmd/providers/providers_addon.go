package providers

import "github.com/grik-ai/ricochet-task/pkg/providers"

// GetRegistry возвращает текущий реестр провайдеров
func GetRegistry() *providers.ProviderRegistry {
	return registry
}