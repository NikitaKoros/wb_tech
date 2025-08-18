package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/logger"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type KafkaConsumerInterface interface {
	Consume(ctx context.Context, handler func(context.Context, *model.Order) error) error
	Close() error
}

type KafkaConfig struct {
	BootstrapServers  string        `env:"KAFKA_BOOTSTRAP_SERVERS" env-required:"true"`
	GroupID           string        `env:"KAFKA_GROUP_ID" env-required:"true"`
	Topic             string        `env:"KAFKA_TOPIC" env-required:"true"`
	AutoOffset        string        `env:"KAFKA_AUTO_OFFSET" default:"earliest"`
	SessionTimeout    time.Duration `env:"KAFKA_SESSION_TIMEOUT" default:"10s"`
	HeartbeatInterval time.Duration `env:"KAFKA_HEARTBEAT_INTERVAL" default:"3s"`
	MaxPollInterval   time.Duration `env:"KAFKA_MAX_POLL_INTERVAL" default:"5m"`
	MaxRetries        int           `env:"KAFKA_MAX_RETRIES" default:"3"`
	CleanupInterval   time.Duration `env:"KAFKA_CLEANUP_INTERVAL" default:"5m"`
	MaxAge            time.Duration `env:"KAFKA_MAX_AGE" default:"30m"`
}

type KafkaConsumer struct {
	consumer       *kafka.Consumer
	topic          string
	logger         logger.Logger
	config         KafkaConfig
	validator      *validator.Validate
	processed      map[string]time.Time
	processedMutex sync.RWMutex
	cleanupTicker  *time.Ticker
	cleanupDone    chan struct{}
}

func NewKafkaConsumer(config KafkaConfig, logger logger.Logger) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     config.BootstrapServers,
		"group.id":              config.GroupID,
		"auto.offset.reset":     config.AutoOffset,
		"enable.auto.commit":    false,
		"session.timeout.ms":    int(config.SessionTimeout.Milliseconds()),
		"heartbeat.interval.ms": int(config.HeartbeatInterval.Milliseconds()),
		"max.poll.interval.ms":  int(config.MaxPollInterval.Milliseconds()),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", srvcerrors.ErrKafka, err)
	}

	if err := c.SubscribeTopics([]string{config.Topic}, nil); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("%w: %v", srvcerrors.ErrKafka, err)
	}

	kc := &KafkaConsumer{
		consumer:    c,
		topic:       config.Topic,
		logger:      logger,
		config:      config,
		validator:   validator.New(),
		processed:   make(map[string]time.Time),
		cleanupDone: make(chan struct{}),
	}

	kc.cleanupTicker = time.NewTicker(kc.config.CleanupInterval)
	go kc.startCleanupRoutine()

	return kc, nil
}

func (k *KafkaConsumer) startCleanupRoutine() {
	for {
		select {
		case <-k.cleanupTicker.C:
			k.cleanupProcessed()
		case <-k.cleanupDone:
			return
		}
	}
}

func (k *KafkaConsumer) cleanupProcessed() {
	cutoff := time.Now().Add(-k.config.MaxAge)
	k.processedMutex.Lock()
	defer k.processedMutex.Unlock()

	countBefore := len(k.processed)
	removed := 0

	for key, timestamp := range k.processed {
		if timestamp.Before(cutoff) {
			delete(k.processed, key)
			removed++
		}
	}

	if removed > 0 {
		k.logger.Debug("processed map cleaned",
			zap.Int("removed", removed),
			zap.Int("remaining", countBefore-removed),
			zap.Duration("max_age", k.config.MaxAge))
	}
}

func (k *KafkaConsumer) Consume(ctx context.Context, handler func(context.Context, *model.Order) error) error {
	k.logger.Info("kafka consumer started",
		zap.String("topic", k.topic),
		zap.String("group_id", k.config.GroupID))

	for {
		select {
		case <-ctx.Done():
			k.logger.Info("kafka consumer shutting down")
			return nil
		default:
			msg, err := k.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if isTimeoutError(err) {
					continue
				}
				k.logger.Error("kafka consume error",
					zap.String("topic", k.topic),
					zap.Error(fmt.Errorf("%w: %v", srvcerrors.ErrKafka, err)))
				continue
			}

			if err := k.processMessageWithRetry(ctx, msg, handler); err != nil {
				k.logger.Error("failed to process message after all retries",
					zap.String("topic", k.topic),
					zap.String("key", string(msg.Key)),
					zap.Int32("partition", msg.TopicPartition.Partition),
					zap.Int64("offset", int64(msg.TopicPartition.Offset)),
					zap.Error(err))
			}
		}
	}
}

func (k *KafkaConsumer) processMessageWithRetry(ctx context.Context, msg *kafka.Message, handler func(context.Context, *model.Order) error) error {
	processedKey := fmt.Sprintf("%s_%d_%d", string(msg.Key), msg.TopicPartition.Partition, msg.TopicPartition.Offset)

	k.processedMutex.RLock()
	_, exists := k.processed[processedKey]
	k.processedMutex.RUnlock()

	if exists {
		k.logger.Debug("skipping already processed message",
			zap.String("processed_key", processedKey),
			zap.String("key", string(msg.Key)),
			zap.Int32("partition", msg.TopicPartition.Partition),
			zap.Int64("offset", int64(msg.TopicPartition.Offset)))

		if _, err := k.consumer.CommitMessage(msg); err != nil {
			k.logger.Error("failed to commit duplicate message",
				zap.String("processed_key", processedKey),
				zap.Error(fmt.Errorf("%w: %v", srvcerrors.ErrKafka, err)))
		}
		return nil
	}

	var ord model.Order
	if err := json.Unmarshal(msg.Value, &ord); err != nil {
		k.logger.Error("failed to unmarshal kafka message into model.Order",
			zap.String("topic", k.topic),
			zap.String("key", string(msg.Key)),
			zap.Int32("partition", msg.TopicPartition.Partition),
			zap.Int64("offset", int64(msg.TopicPartition.Offset)),
			zap.Error(err))

		if _, cerr := k.consumer.CommitMessage(msg); cerr != nil {
			k.logger.Error("failed to commit offset after unmarshal error",
				zap.Error(fmt.Errorf("%w: %v", srvcerrors.ErrKafka, cerr)))
			return fmt.Errorf("%w: failed to commit after unmarshal: %v", srvcerrors.ErrKafka, cerr)
		}
		k.logger.Warn("skipped invalid json message after commit", zap.String("key", string(msg.Key)))
		return nil
	}

	if verr := k.validateOrder(&ord); verr != nil {
		k.logger.Warn("order validation failed",
			zap.String("topic", k.topic),
			zap.String("key", string(msg.Key)),
			zap.Int32("partition", msg.TopicPartition.Partition),
			zap.Int64("offset", int64(msg.TopicPartition.Offset)),
			zap.Error(verr))

		if _, cerr := k.consumer.CommitMessage(msg); cerr != nil {
			k.logger.Error("failed to commit offset after validation error",
				zap.Error(fmt.Errorf("%w: %v", srvcerrors.ErrKafka, cerr)))
			return fmt.Errorf("%w: failed to commit after validation error: %v", srvcerrors.ErrKafka, cerr)
		}
		k.logger.Info("committed offset and skipped invalid order", zap.String("key", string(msg.Key)))
		return nil
	}

	var lastErr error
	for i := 0; i <= k.config.MaxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(1<<uint(i)) * time.Second
			k.logger.Debug("retrying message after error",
				zap.String("key", string(msg.Key)),
				zap.Int("attempt", i),
				zap.Duration("backoff", backoff))
			time.Sleep(backoff)
		}

		handlerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := handler(handlerCtx, &ord); err != nil {
			lastErr = err
			k.logger.Warn("handler error",
				zap.String("key", string(msg.Key)),
				zap.Int("attempt", i),
				zap.Int("max_retries", k.config.MaxRetries),
				zap.Error(err))

			if isTemporaryError(err) {
				continue
			}

			k.logger.Error("permanent handler error, committing offset and skipping message",
				zap.String("key", string(msg.Key)),
				zap.Error(err))
			if _, cerr := k.consumer.CommitMessage(msg); cerr != nil {
				k.logger.Error("failed to commit offset after permanent handler error",
					zap.Error(fmt.Errorf("%w: %v", srvcerrors.ErrKafka, cerr)))
				return fmt.Errorf("%w: failed to commit after permanent handler error: %v", srvcerrors.ErrKafka, cerr)
			}

			k.processedMutex.Lock()
			k.processed[processedKey] = time.Now()
			k.processedMutex.Unlock()

			return err
		}

		if _, cerr := k.consumer.CommitMessage(msg); cerr != nil {
			k.logger.Error("failed to commit message offset",
				zap.String("key", string(msg.Key)),
				zap.Error(fmt.Errorf("%w: %v", srvcerrors.ErrKafka, cerr)))
			return fmt.Errorf("%w: %v", srvcerrors.ErrKafka, cerr)
		}

		k.processedMutex.Lock()
		k.processed[processedKey] = time.Now()
		k.processedMutex.Unlock()

		k.logger.Info("message successfully processed",
			zap.String("key", string(msg.Key)),
			zap.String("topic", k.topic),
			zap.Int32("partition", msg.TopicPartition.Partition),
			zap.Int64("offset", int64(msg.TopicPartition.Offset)))
		return nil
	}

	k.logger.Error("all retries exhausted, committing offset and skipping message",
		zap.String("key", string(msg.Key)),
		zap.Error(lastErr))

	if _, cerr := k.consumer.CommitMessage(msg); cerr != nil {
		k.logger.Error("failed to commit offset after exhausting retries",
			zap.Error(fmt.Errorf("%w: %v", srvcerrors.ErrKafka, cerr)))
		return fmt.Errorf("%w: failed to commit after retries: %v", srvcerrors.ErrKafka, cerr)
	}

	k.processedMutex.Lock()
	k.processed[processedKey] = time.Now()
	k.processedMutex.Unlock()

	return lastErr
}

func isTemporaryError(err error) bool {
	return errors.Is(err, srvcerrors.ErrDatabase)
}

func isTimeoutError(err error) bool {
	var kErr kafka.Error
	if errors.As(err, &kErr) {
		return kErr.Code() == kafka.ErrTimedOut
	}
	return false
}

func (k *KafkaConsumer) Close() error {
	k.logger.Info("closing kafka consumer")

	close(k.cleanupDone)
	k.cleanupTicker.Stop()

	k.processedMutex.Lock()
	k.processed = make(map[string]time.Time)
	k.processedMutex.Unlock()

	if err := k.consumer.Close(); err != nil {
		return fmt.Errorf("%w: failed to close Kafka consumer: %v", srvcerrors.ErrKafka, err)
	}

	return nil
}

func (k *KafkaConsumer) validateOrder(o *model.Order) error {
	if o == nil {
		return fmt.Errorf("order is nil")
	}
	if err := k.validator.Struct(o); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			parts := make([]string, 0, len(ve))
			for _, e := range ve {
				parts = append(parts, fmt.Sprintf("%s: %s", e.Field(), e.Tag()))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(parts, "; "))
		}
		return err
	}
	return nil
}
