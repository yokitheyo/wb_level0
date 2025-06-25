.PHONY: build run test docker-up docker-down send-test-order clean

build:
	go build -o bin/wb_level0 cmd/main.go

run: build
	./bin/wb_level0

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

send-test-order:
	go run scripts/send_test_order.go

send-multiple-orders:
	go run scripts/send_multiple_orders.go

send-invalid-orders:
	go run scripts/send_invalid_orders.go

test-cache:
	go run scripts/test_cache_performance.go

docker-build:
	docker build -t wb-level0 .

docker-run:
	docker run -p 8081:8081 --name wb-level0-app wb-level0

docker-full-up:
	docker-compose -f docker-compose.full.yml up -d

docker-full-down:
	docker-compose -f docker-compose.full.yml down -v

clean:
	rm -rf bin/

stop: docker-down clean

status:
	docker-compose ps

logs:
	docker-compose logs -f

migrate-up:
	docker exec -it orders-postgres psql -U orders_user -d orders_service -f /docker-entrypoint-initdb.d/000001_create_orders_tables.up.sql

db-connect:
	docker exec -it orders-postgres psql -U orders_user -d orders_service

kafka-ui:
	@echo "Kafka UI доступен по адресу: http://localhost:8080"

create-topic:
	docker exec orders-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic orders --partitions 1 --replication-factor 1 --if-not-exists
