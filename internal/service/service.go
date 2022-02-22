package service

import (
	"kratos-kafka-server-demo/internal/biz"
	"kratos-kafka-server-demo/pkg/kafka"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func NewServiceHandlers(uc *biz.GreeterUsecase, logger log.Logger) []kafka.Handler {
	return []kafka.Handler{
		NewGreeterService(uc, logger),
	}
}

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewServiceHandlers)
