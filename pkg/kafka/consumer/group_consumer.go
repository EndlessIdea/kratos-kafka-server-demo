package consumer

import (
	"context"
	"sync"
	"time"

	"kratos-kafka-server-demo/pkg/kafka"

	"github.com/Shopify/sarama"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

const (
	defaultVersion     = "2.1.0"
	defaultMaxProcTime = 250 * time.Millisecond
)

type GroupConsumer struct {
	// config setting
	brokers         []string
	topics          []string
	group           string
	version         string
	balanceStrategy sarama.BalanceStrategy
	initialOffset   int64
	maxProcTime     time.Duration

	ready chan struct{}

	sarama.Client
	consumer sarama.ConsumerGroup
	handlers map[string]kafka.Handler
	logger   *log.Helper
}

// ConsumerOption is a GroupConsumer option.
type GroupConsumerOption func(*GroupConsumer)

// Version with Kafka cluster version.
func Version(version string) GroupConsumerOption {
	return func(c *GroupConsumer) {
		c.version = version
	}
}

// BalanceStrategy with the rebalance strategy
func BalanceStrategy(strategy sarama.BalanceStrategy) GroupConsumerOption {
	return func(c *GroupConsumer) {
		c.balanceStrategy = strategy
	}
}

// InitialOffset with the consumer initial offset
func InitialOffset(offset int64) GroupConsumerOption {
	return func(c *GroupConsumer) {
		c.initialOffset = offset
	}
}

// MaxProcessingTime with the maximum amount of time the consumer expects a message takes
func MaxProcessingTime(maxTime time.Duration) GroupConsumerOption {
	return func(c *GroupConsumer) {
		c.maxProcTime = maxTime
	}
}

// Logger with the specify logger
func Logger(logger log.Logger) GroupConsumerOption {
	return func(c *GroupConsumer) {
		c.logger = log.NewHelper(logger)
	}
}

// NewGroupConsumer inits a consumer group consumer
func NewGroupConsumer(brokers, topics []string, group string, opts ...GroupConsumerOption) (*GroupConsumer, error) {
	// parse config setting
	result := &GroupConsumer{
		brokers:         brokers,
		topics:          topics,
		group:           group,
		version:         defaultVersion,
		balanceStrategy: sarama.BalanceStrategySticky,
		initialOffset:   sarama.OffsetNewest,
		maxProcTime:     defaultMaxProcTime,
		ready:           make(chan struct{}),
		handlers:        make(map[string]kafka.Handler),
		logger:          log.NewHelper(log.DefaultLogger),
	}

	for _, o := range opts {
		o(result)
	}

	// check version
	kafkaVersion, err := sarama.ParseKafkaVersion(result.version)
	if err != nil {
		return nil, errors.Wrapf(err, "parse kafka version %s error", result.version)
	}

	// init config
	config := sarama.NewConfig()
	config.Version = kafkaVersion
	config.Consumer.MaxProcessingTime = result.maxProcTime
	config.Consumer.Group.Rebalance.Strategy = result.balanceStrategy
	config.Consumer.Offsets.Initial = result.initialOffset
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "init sarama kafka client error")
	}
	consumer, err := sarama.NewConsumerGroupFromClient(group, client)
	if err != nil {
		return nil, errors.Wrap(err, "init sarama kafka group consumer error")
	}
	result.Client = client
	result.consumer = consumer

	return result, nil
}

// Topics returns all the topics this consumer subscribes
func (c *GroupConsumer) Topics() []string {
	return c.topics
}

// RegisterHandler registers a handler to handle the messages of a specific topic
func (c *GroupConsumer) RegisterHandler(handler kafka.Handler) {
	c.handlers[handler.Topic()] = handler
}

// RegisterHandler checks whether this consumer has a handler for the specific topic
func (c *GroupConsumer) HasHandler(topic string) bool {
	_, ok := c.handlers[topic]
	return ok
}

// Consume starts the consumer to receive and handle the messages
func (c *GroupConsumer) Consume(ctx context.Context) error {
	// check handlers before consuming
	for _, topic := range c.topics {
		if _, ok := c.handlers[topic]; !ok {
			return errors.Errorf("no handler for topic %s", topic)
		}
	}

	// start consuming
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			c.logger.Infof("consumer %+v session starts", c)

			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.consumer.Consume(ctx, c.topics, c); err != nil {
				c.logger.Errorf("consumer %+v consumes error %+v", c, err)
				return
			}

			c.logger.Infof("consumer %+v session exits", c)

			// check if context was cancelled, signaling that the consumer should stop
			if err := ctx.Err(); err != nil {
				c.logger.Errorf("consumer %+v exits due to context is canceled %+v", c, err)
				return
			}

			c.ready = make(chan struct{})
		}
	}()

	<-c.ready // Await till the consumer has been set up

	wg.Wait()
	if err := c.consumer.Close(); err != nil {
		return errors.Errorf("close kafka consumer %+v error %+v", c.consumer, err)
	}

	return errors.Errorf("consumer %+v exited", c.consumer)
}

func (c *GroupConsumer) Close() error {
	return c.consumer.Close()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *GroupConsumer) Setup(session sarama.ConsumerGroupSession) error {
	c.logger.Infof("consumer of group %s setup status %+v", c.group, session.Claims())
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *GroupConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	c.logger.Infof("consumer of group %s exit status %+v", c.group, session.Claims())
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *GroupConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		handler, ok := c.handlers[message.Topic]
		if !ok {
			c.logger.Errorf("no consumer handler for topic %s", message.Topic)
			continue
		}
		if err := handler.Handle(message); err != nil {
			// make sure you have a way to record or retry the error message
			c.logger.Errorf("consume message %s of topic %s partition %d error %+v", string(message.Value), message.Topic, message.Partition, err)
			continue
		}
		session.MarkMessage(message, "")
	}

	return nil
}
