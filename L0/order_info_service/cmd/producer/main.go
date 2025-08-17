package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Config struct {
	BootstrapServers string
	Topic            string
	ValidCount       int
	InvalidCount     int
	DelayMs          int
	Seed             int64
}

type OrderGenerator struct {
	rand *rand.Rand
}

func NewOrderGenerator(seed int64) *OrderGenerator {
	return &OrderGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (g *OrderGenerator) GenerateValidOrder() *model.Order {
	orderUID := g.generateOrderUID()
	trackNumber := fmt.Sprintf("WBILMTESTTRACK%d", g.rand.Intn(10000))

	return &model.Order{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: model.Delivery{
			OrderUID: orderUID,
			Name:     g.generateName(),
			Phone:    "+9720000000",
			Zip:      "2639809",
			City:     "Kiryat Mozkin",
			Address:  "Ploshad Mira 15",
			Region:   "Kraiot",
			Email:    "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []*model.Item{
			{
				OrderUID:    orderUID,
				ChrtID:      9934930,
				TrackNumber: trackNumber,
				Price:       453,
				RID:         g.generateRID(),
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}

func (g *OrderGenerator) GenerateInvalidOrder() *model.Order {
	orderUID := g.generateOrderUID()
	trackNumber := fmt.Sprintf("WBILMTESTTRACK%d", g.rand.Intn(10000))

	invalidType := g.rand.Intn(5)

	order := &model.Order{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: model.Delivery{
			OrderUID: orderUID,
			Name:     g.generateName(),
			Zip:      "2639809",
			City:     "Kiryat Mozkin",
			Address:  "Ploshad Mira 15",
			Region:   "Kraiot",
			Email:    "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []*model.Item{
			{
				OrderUID:    orderUID,
				ChrtID:      9934930,
				TrackNumber: trackNumber,
				Price:       453,
				RID:         g.generateRID(),
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	switch invalidType {
	case 0:
		order.OrderUID = "b563feb7b2b84b6!@#test"
	case 1:
		order.Delivery.Phone = "1234567890"
	case 2:
		order.Locale = "es"
	case 3:
		order.Items = []*model.Item{}
	case 4:
		order.DateCreated = time.Now().Add(24 * time.Hour)
	}

	return order
}

func (g *OrderGenerator) generateOrderUID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 20)
	for i := range b {
		b[i] = charset[g.rand.Intn(len(charset))]
	}
	return string(b)
}

func (g *OrderGenerator) generateName() string {
	names := []string{"John Doe", "Jane Smith", "Alice Johnson", "Bob Wilson", "Test Testov"}
	return names[g.rand.Intn(len(names))]
}

func (g *OrderGenerator) generateRID() string {
	return fmt.Sprintf("ab4219087a764ae0b%d", g.rand.Intn(10000))
}

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaProducer(config Config) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServers,
		"acks":              "all",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &KafkaProducer{
		producer: p,
		topic:    config.Topic,
	}, nil
}

func (kp *KafkaProducer) SendOrder(order *model.Order) error {
	value, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	err = kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kp.topic,
			Partition: kafka.PartitionAny,
		},
		Value: value,
	}, deliveryChan)

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
	}

	return nil
}

func (kp *KafkaProducer) Close() {
	kp.producer.Flush(5000)
	kp.producer.Close()
}

func main() {
	defaultBootstrap := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if defaultBootstrap == "" {
		defaultBootstrap = "localhost:19092"
	}

	var (
		bootstrapServers = flag.String("bootstrap-servers", defaultBootstrap, "Kafka bootstrap servers")
		topic            = flag.String("topic", "orders", "Kafka topic")
		validCount       = flag.Int("valid-count", 10, "Number of valid orders to generate")
		invalidCount     = flag.Int("invalid-count", 5, "Number of invalid orders to generate")
		delayMs          = flag.Int("delay-ms", 100, "Delay between messages in milliseconds")
		seed             = flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	)
	flag.Parse()

	config := Config{
		BootstrapServers: *bootstrapServers,
		Topic:            *topic,
		ValidCount:       *validCount,
		InvalidCount:     *invalidCount,
		DelayMs:          *delayMs,
		Seed:             *seed,
	}

	generator := NewOrderGenerator(config.Seed)

	producer, err := NewKafkaProducer(config)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	fmt.Printf("Starting order generator with config:\n")
	fmt.Printf("- Bootstrap servers: %s\n", config.BootstrapServers)
	fmt.Printf("- Topic: %s\n", config.Topic)
	fmt.Printf("- Valid orders: %d\n", config.ValidCount)
	fmt.Printf("- Invalid orders: %d\n", config.InvalidCount)
	fmt.Printf("- Delay: %dms\n", config.DelayMs)
	fmt.Printf("- Random seed: %d\n\n", config.Seed)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		fmt.Println("\nReceived termination signal, shutting down...")
		producer.Close()
		os.Exit(0)
	}()

	fmt.Printf("Generating %d valid orders...\n", config.ValidCount)
	for i := 0; i < config.ValidCount; i++ {
		order := generator.GenerateValidOrder()
		if err := producer.SendOrder(order); err != nil {
			log.Printf("Failed to send valid order %d: %v", i+1, err)
		} else {
			fmt.Printf("Sent valid order %d/%d (UID: %s)\n", i+1, config.ValidCount, order.OrderUID)
		}

		if config.DelayMs > 0 && i < config.ValidCount-1 {
			time.Sleep(time.Duration(config.DelayMs) * time.Millisecond)
		}
	}

	fmt.Printf("\nGenerating %d invalid orders...\n", config.InvalidCount)
	for i := 0; i < config.InvalidCount; i++ {
		order := generator.GenerateInvalidOrder()
		if err := producer.SendOrder(order); err != nil {
			log.Printf("Failed to send invalid order %d: %v", i+1, err)
		} else {
			fmt.Printf("Sent invalid order %d/%d (UID: %s)\n", i+1, config.InvalidCount, order.OrderUID)
		}

		if config.DelayMs > 0 && i < config.InvalidCount-1 {
			time.Sleep(time.Duration(config.DelayMs) * time.Millisecond)
		}
	}

	fmt.Println("\nAll orders sent successfully!")
}
