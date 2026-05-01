//go:build wireinject
// +build wireinject

package main

import (
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/app"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/provider"
	"github.com/google/wire"
)

func InitializeApp() (*app.Application, error) {
	wire.Build(provider.Set, app.NewApplication)
	return nil, nil
}
