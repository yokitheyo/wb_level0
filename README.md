# WB Level 0 - Сервис заказов

Микросервис для обработки заказов с использованием Go, PostgreSQL и Apache Kafka.


## Запуск

### Вариант 1: Инфраструктура в Docker, приложение локально
```bash
# Запуск PostgreSQL, Kafka, Zookeeper
docker-compose up -d

# Запуск приложения
go run cmd/main.go
```

### Вариант 2: Все в Docker
```bash
docker-compose -f docker-compose.full.yml up -d
```

## Тестирование

```bash
# Отправка тестового заказа
go run scripts/send_test_order.go

# Проверка API
curl http://localhost:8081/order/b563feb7b2b84b6test

# Веб-интерфейс
open http://localhost:8081
```

## Endpoints

- **Веб-интерфейс**: http://localhost:8081
- **API**: http://localhost:8081/order/{order_uid}
- **Kafka UI**: http://localhost:8080
