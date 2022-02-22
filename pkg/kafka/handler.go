package kafka

import (
	"github.com/Shopify/sarama"
)

// Handler is used to handle the received message
type Handler interface {
	Topic() string
	Handle(message *sarama.ConsumerMessage) error
}
