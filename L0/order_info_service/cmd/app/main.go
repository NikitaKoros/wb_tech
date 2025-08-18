package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/cache"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/controller"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/handler"
	kafka "github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/kafka_consumer"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/logger"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/repository"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
	"github.com/caarlos0/env/v6"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type Config struct {
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"postgres"`
	DBName     string `env:"DB_NAME" envDefault:"orders"`

	KafkaBootstrapServers  string        `env:"KAFKA_BOOTSTRAP_SERVERS" envDefault:"localhost:9092"`
	KafkaGroupID           string        `env:"KAFKA_GROUP_ID" envDefault:"order-info-service"`
	KafkaTopic             string        `env:"KAFKA_TOPIC" envDefault:"orders"`
	KafkaAutoOffset        string        `env:"KAFKA_AUTO_OFFSET" envDefault:"earliest"`
	KafkaSessionTimeout    time.Duration `env:"KAFKA_SESSION_TIMEOUT" envDefault:"10s"`
	KafkaRebalanceTimeout  time.Duration `env:"KAFKA_REBALANCE_TIMEOUT" envDefault:"15s"`
	KafkaHeartbeatInterval time.Duration `env:"KAFKA_HEARTBEAT_INTERVAL" envDefault:"3s"`
	KafkaMaxPollInterval   time.Duration `env:"KAFKA_MAX_POLL_INTERVAL" envDefault:"5m"`
	KafkaMaxRetries        int           `env:"KAFKA_MAX_RETRIES" envDefault:"3"`
	KafkaCleanupInterval   time.Duration `env:"KAFKA_CLEANUP_INTERVAL" envDefault:"5m"`
	KafkaMaxAge            time.Duration `env:"KAFKA_MAX_AGE" envDefault:"30m"`

	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logg, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := initDB(cfg, logg)
	if err != nil {
		logg.Error("failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	schemaPath := "./db/postgres.sql"
	if err := applyDBSchema(db, schemaPath, logg); err != nil {
		logg.Error("failed to apply database schema", zap.Error(err))
		os.Exit(1)
	}

	repo := repository.NewOrderRepository(db)

	cache := cache.NewLocalCache()

	ctrl := controller.NewController(repo, cache, logg)

	httpHandler := handler.NewHandler(ctrl, logg)

	kafkaConfig := kafka.KafkaConfig{
		BootstrapServers: cfg.KafkaBootstrapServers,
		GroupID:          cfg.KafkaGroupID,
		Topic:            cfg.KafkaTopic,
		AutoOffset:       cfg.KafkaAutoOffset,
		SessionTimeout:   cfg.KafkaSessionTimeout,
		HeartbeatInterval: cfg.KafkaHeartbeatInterval,
		MaxPollInterval:   cfg.KafkaMaxPollInterval,
		MaxRetries:        cfg.KafkaMaxRetries,
		CleanupInterval:   cfg.KafkaCleanupInterval,
		MaxAge:            cfg.KafkaMaxAge,
	}
	kafkaConsumer, err := kafka.NewKafkaConsumer(kafkaConfig, logg)
	if err != nil {
		logg.Error("failed to create kafka consumer", zap.Error(err))
		os.Exit(1)
	}

	var cachedAmount int
	if cachedAmount, err = controller.WarmUpCache(context.Background(), repo, cache, 100); err != nil {
		logg.Error("failed to warm up cache", zap.Error(err))
	}
	
	logg.Info("orders added to cache", zap.Int("amount", cachedAmount))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: httpHandler,
	}
	go func() {
		logg.Info("starting HTTP server",
			zap.String("addr", server.Addr),
			zap.String("port", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error("HTTP server error", zap.Error(err))
		}
	}()

	go func() {
		if err := kafkaConsumer.Consume(ctx, func(ctx context.Context, order *model.Order) error {
			_, err := repo.UpsertOrder(ctx, order)
			return err
		}); err != nil {
			logg.Error("kafka consumer error", zap.Error(err))
		}
	}()

	logg.Info("application started",
		zap.String("db_host", cfg.DBHost),
		zap.String("db_port", cfg.DBPort),
		zap.String("db_name", cfg.DBName),
		zap.String("kafka_brokers", cfg.KafkaBootstrapServers),
		zap.String("kafka_topic", cfg.KafkaTopic),
		zap.String("server_port", cfg.ServerPort))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logg.Info("shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logg.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	if err := kafkaConsumer.Close(); err != nil {
		logg.Error("failed to close kafka consumer", zap.Error(err))
	}

	logg.Info("application shutdown complete")
}

func initDB(cfg Config, log logger.Logger) (*sql.DB, error) {
	psqlInfo := "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"
	psqlInfo = fmt.Sprintf(psqlInfo,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", srvcerrors.ErrDatabase, err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%w: %v", srvcerrors.ErrDatabase, err)
	}

	log.Info("database successfully connected",
		zap.String("host", cfg.DBHost),
		zap.String("port", cfg.DBPort),
		zap.String("dbname", cfg.DBName))

	return db, nil
}

func applyDBSchema(db *sql.DB, schemaPath string, log logger.Logger) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Error("failed to read db schema file", zap.String("path", schemaPath), zap.Error(err))
		return fmt.Errorf("read schema: %w", err)
	}

	if len(bytes.TrimSpace(schema)) == 0 {
		log.Info("db schema file is empty, skip applying", zap.String("path", schemaPath))
		return nil
	}

	if _, err := db.Exec(string(schema)); err != nil {
		log.Error("failed to execute db schema", zap.String("path", schemaPath), zap.Error(err))
		return fmt.Errorf("exec schema: %w", err)
	}

	log.Info("database schema applied", zap.String("path", schemaPath))
	return nil
}
