DEV_DB_HOST ?= 127.0.0.1
DEV_DB_PORT ?= 5432
DEV_DB_USER ?= postgres
DEV_DB_PASSWORD ?= postgres
DEV_DB_NAME ?= orders

KAFKA_BOOTSTRAP_SERVERS ?= localhost:19092
INTERNAL_KAFKA_BOOTSTRAP ?= kafka:9092

SERVER_PORT ?= 8080

TEST_DB_PORT ?= 5431
TEST_DB_USER ?= testuser
TEST_DB_PASSWORD ?= testpass
TEST_DB_NAME ?= testdb

.PHONY: start-postgres wait-postgres stop-postgres start-dev-db wait-dev-db stop-dev-db \
        start-kafka wait-kafka stop-kafka create-kafka-topics run run-dev producer show-config \
        test-all test-repository test-cache test-controller test-handler start-all stop-all

start-postgres:
	@docker-compose -f ./test/docker-compose.db_test.yml up -d

wait-postgres:
	@echo "Waiting for PostgreSQL to start..."
	@docker-compose -f ./test/docker-compose.db_test.yml up -d --wait
	@echo "PostgreSQL is ready"

stop-postgres:
	@docker-compose -f ./test/docker-compose.db_test.yml down --remove-orphans

start-dev-db:
	@export DEV_DB_HOST=$(DEV_DB_HOST) && \
	export DEV_DB_PORT=$(DEV_DB_PORT) && \
	export DEV_DB_USER=$(DEV_DB_USER) && \
	export DEV_DB_PASSWORD=$(DEV_DB_PASSWORD) && \
	export DEV_DB_NAME=$(DEV_DB_NAME) && \
	docker-compose -f ./docker-compose.dev.yml up -d

wait-dev-db:
	@export DEV_DB_HOST=$(DEV_DB_HOST) && \
	export DEV_DB_PORT=$(DEV_DB_PORT) && \
	export DEV_DB_USER=$(DEV_DB_USER) && \
	export DEV_DB_PASSWORD=$(DEV_DB_PASSWORD) && \
	export DEV_DB_NAME=$(DEV_DB_NAME) && \
	docker-compose -f ./docker-compose.dev.yml up --wait

stop-dev-db:
	@docker-compose -f ./docker-compose.dev.yml down --remove-orphans

start-kafka:
	@docker-compose -f ./docker-compose.kafka.yml up -d

wait-kafka:
	@echo "Waiting for Kafka to be ready..."
	@sh -c '\
	max_attempts=30; \
	attempt=0; \
	while [ $$attempt -lt $$max_attempts ]; do \
		docker-compose -f ./docker-compose.kafka.yml exec -T kafka kafka-broker-api-versions --bootstrap-server $(INTERNAL_KAFKA_BOOTSTRAP) >/dev/null 2>&1 && { echo "Kafka is ready!"; exit 0; }; \
		printf "Kafka not ready yet, retrying in 2 seconds... (attempt %d/%d)\n" $$((attempt+1)) $$max_attempts; \
		sleep 2; \
		attempt=$$((attempt+1)); \
	done; \
	echo "ERROR: Kafka failed to start within the expected time"; \
	exit 1'

stop-kafka:
	@docker-compose -f ./docker-compose.kafka.yml down

create-kafka-topics:
	@echo "Creating Kafka topics..."
	@docker-compose -f ./docker-compose.kafka.yml exec -T kafka \
		kafka-topics --bootstrap-server $(INTERNAL_KAFKA_BOOTSTRAP) --create --if-not-exists --topic orders --partitions 1 --replication-factor 1

run:
	@echo "Starting application..."
	@DB_HOST=$(DEV_DB_HOST) DB_PORT=$(DEV_DB_PORT) DB_USER=$(DEV_DB_USER) \
		DB_PASSWORD=$(DEV_DB_PASSWORD) DB_NAME=$(DEV_DB_NAME) \
		KAFKA_BOOTSTRAP_SERVERS=$(KAFKA_BOOTSTRAP_SERVERS) \
		SERVER_PORT=$(SERVER_PORT) \
		go run ./order_info_service/cmd/app/main.go

run-dev: start-dev-db wait-dev-db start-kafka wait-kafka create-kafka-topics run

producer:
	@echo "Running producer for testing..."
	@KAFKA_BOOTSTRAP_SERVERS=$(KAFKA_BOOTSTRAP_SERVERS) \
		DB_HOST=$(DEV_DB_HOST) DB_PORT=$(DEV_DB_PORT) DB_USER=$(DEV_DB_USER) \
		DB_PASSWORD=$(DEV_DB_PASSWORD) DB_NAME=$(DEV_DB_NAME) \
		go run ./order_info_service/cmd/producer/main.go $(ARGS)

show-config:
	@echo "Development DB configuration:"
	@echo "  DB_HOST=$(DEV_DB_HOST)"
	@echo "  DB_PORT=$(DEV_DB_PORT)"
	@echo "  DB_USER=$(DEV_DB_USER)"
	@echo "  DB_PASSWORD=$(DEV_DB_PASSWORD)"
	@echo "  DB_NAME=$(DEV_DB_NAME)"
	@echo ""
	@echo "Kafka configuration (external):"
	@echo "  KAFKA_BOOTSTRAP_SERVERS=$(KAFKA_BOOTSTRAP_SERVERS)"
	@echo "Kafka configuration (internal for container commands):"
	@echo "  INTERNAL_KAFKA_BOOTSTRAP=$(INTERNAL_KAFKA_BOOTSTRAP)"
	@echo ""
	@echo "Server configuration:"
	@echo "  SERVER_PORT=$(SERVER_PORT)"

test-all:
	$(MAKE) test-repository
	$(MAKE) test-cache
	$(MAKE) test-controller
	$(MAKE) test-handler

test-repository:
	$(MAKE) start-postgres
	$(MAKE) wait-postgres
	@echo "Running repository tests..."
	@richgo test ./order_info_service/internal/repository/... -v; \
	ret=$$?; \
	$(MAKE) stop-postgres; \
	exit $$ret

test-cache:
	@echo "Running cache tests..."
	@richgo test ./order_info_service/internal/cache/... -v

test-controller:
	@echo "Running controller tests..."
	@richgo test ./order_info_service/internal/controller/... -v

test-handler:
	@echo "Running handler tests..."
	@richgo test ./order_info_service/internal/handler/... -v

start-all: start-dev-db wait-dev-db start-kafka wait-kafka create-kafka-topics run

stop-all: stop-dev-db stop-kafka
