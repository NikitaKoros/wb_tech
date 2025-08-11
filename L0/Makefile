start-postgres:
	@docker-compose -f ./test/docker-compose.db_test.yml up -d

wait-postgres:
	@echo "Waiting for PostgreSQL to start..."
	@docker-compose -f ./test/docker-compose.db_test.yml up -d --wait
	@echo "PostgreSQL is ready"

stop-postgres:
	@docker-compose -f ./test/docker-compose.db_test.yml down
	
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
test-handler:
