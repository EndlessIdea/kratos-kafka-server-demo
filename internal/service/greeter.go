package service

import (
	"fmt"

	"kratos-kafka-server-demo/internal/biz"
	"kratos-kafka-server-demo/pkg/kafka"

	"github.com/Shopify/sarama"
	"github.com/go-kratos/kratos/v2/log"
)

var _ kafka.Handler = (*GreeterService)(nil)

const (
	greeterTopic = "greeter"
)

// GreeterService is a greeter service.
type GreeterService struct {
	topic string
	uc    *biz.GreeterUsecase
	log   *log.Helper
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase, logger log.Logger) *GreeterService {
	return &GreeterService{topic: greeterTopic, uc: uc, log: log.NewHelper(logger)}
}

func (s *GreeterService) Topic() string {
	return s.topic
}

func (s *GreeterService) Handle(message *sarama.ConsumerMessage) error {
	fmt.Printf("receive greeter message key %s value %s\n", string(message.Key), string(message.Value))

	return nil
}
