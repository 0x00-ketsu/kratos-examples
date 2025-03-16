//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"
	"usermanage/gen/proto/conf"
	"usermanage/internal/biz"
	"usermanage/internal/data"
	"usermanage/internal/pkg/db"
	"usermanage/internal/server"
	"usermanage/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireData init data.
func wireData(*conf.Data, log.Logger) (*data.Data, error) {
	panic(wire.Build(db.ProviderSet, data.NewData))
}

// wireApp init kratos application.
func wireApp(context.Context, *conf.Server, *conf.Data, log.Logger) (*kratos.App, error) {
	panic(
		wire.Build(
			server.ProviderSet,
			data.ProviderSet,
			biz.ProviderSet,
			service.ProviderSet,
			newApp,
		),
	)
}
