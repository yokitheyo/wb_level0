package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewProducer(brokers []string, topic string, logger *zap.Logger) *Producer {
	writer := kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Completion:   nil,
		Compression:  kafka.Snappy,
		Logger:       kafka.LoggerFunc(logger.Sugar().Infof),
		ErrorLogger:  kafka.LoggerFunc(logger.Sugar().Errorf),
		Transport: &kafka.Transport{
			Dial: (&kafka.Dialer{
				Timeout:   10 * time.Second,
				DualStack: true,
			}).DialFunc,
		},
	}
	return &Producer{
		writer: &writer,
		logger: logger,
	}
}

func (p *Producer) SendMessage(ctx context.Context, message []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})
	if err != nil {
		p.logger.Error("failed to send message to kafka", zap.Error(err))
		return err
	}
	p.logger.Info("message sent to kafka successfully")
	return nil
}

func (p *Producer) SendMessageWithKey(ctx context.Context, key, message []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: message,
	})
	if err != nil {
		p.logger.Error("failed to send message with key to kafka",
			zap.Error(err),
			zap.String("key", string(key)))
		return err
	}
	p.logger.Info("message with key sent to kafka successfully", zap.String("key", string(key)))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
