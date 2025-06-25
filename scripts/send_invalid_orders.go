package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func main() {
	writer := kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	ctx := context.Background()

	invalidMessages := []struct {
		name    string
		message string
	}{
		{
			name:    "Невалидный JSON",
			message: `{"order_uid": "invalid_json", "track_number":}`,
		},
		{
			name:    "Пустой order_uid",
			message: `{"order_uid": "", "track_number": "TEST123"}`,
		},
		{
			name:    "Отсутствует обязательное поле",
			message: `{"order_uid": "missing_fields", "entry": "WBIL"}`,
		},
		{
			name:    "Пустое сообщение",
			message: `{}`,
		},
		{
			name:    "Не JSON",
			message: `это не JSON сообщение`,
		},
	}

	for i, invalid := range invalidMessages {
		err := writer.WriteMessages(ctx, kafka.Message{
			Value: []byte(invalid.message),
		})
		if err != nil {
			log.Printf("Failed to send invalid message %d (%s): %v", i+1, invalid.name, err)
			continue
		}

		log.Printf("Successfully sent invalid message %d: %s", i+1, invalid.name)
	}

	log.Println("All invalid messages sent!")
}
