// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"kratos-kafka-server-demo/internal/biz"
	"kratos-kafka-server-demo/internal/conf"
	"kratos-kafka-server-demo/internal/data"
	"kratos-kafka-server-demo/internal/server"
	"kratos-kafka-server-demo/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// initApp init kratos application.
func initApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
