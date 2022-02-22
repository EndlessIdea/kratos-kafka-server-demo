// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"kratos-kafka-server-demo/internal/biz"
	"kratos-kafka-server-demo/internal/conf"
	"kratos-kafka-server-demo/internal/data"
	"kratos-kafka-server-demo/internal/server"
	"kratos-kafka-server-demo/internal/service"
)

// Injectors from wire.go:

// initApp init kratos application.
func initApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	dataData, cleanup, err := data.NewData(confData, logger)
	if err != nil {
		return nil, nil, err
	}
	greeterRepo := data.NewGreeterRepo(dataData, logger)
	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
	v := service.NewServiceHandlers(greeterUsecase, logger)
	kafkaServer, err := server.NewKafkaServer(confServer, v)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	app := newApp(logger, kafkaServer)
	return app, func() {
		cleanup()
	}, nil
}