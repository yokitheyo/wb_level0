package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yokitheyo/wb_level0/internal/services"
	"go.uber.org/zap"
)

type Consumer struct {
	reader  *kafka.Reader
	service services.OrderService
	logger  *zap.Logger
}

func NewConsumer(brokers []string,
	topic, groupID string,
	service services.OrderService,
	logger *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 1 * time.Second,
	})
	return &Consumer{
		reader:  reader,
		service: service,
		logger:  logger,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("starting kafka consumer")

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("stopping kafka consumer")
				if err := c.reader.Close(); err != nil {
					c.logger.Error("failed to close kafka reader", zap.Error(err))
				}
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					c.logger.Error("failed to read message from kafka", zap.Error(err))
					continue
				}

				c.logger.Debug("received message from kafka",
					zap.String("topic", msg.Topic),
					zap.Int("partition", msg.Partition),
					zap.Int64("offset", msg.Offset),
				)
				if err := c.service.ProcessOrder(ctx, msg.Value); err != nil {
					c.logger.Error("failed to process order", zap.Error(err))
					continue
				}
			}
		}
	}()
}
