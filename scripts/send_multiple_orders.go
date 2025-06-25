package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func main() {
	writer := kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	ctx := context.Background()

	names := []string{"Иван Иванов", "Петр Петров", "Анна Сидорова", "Мария Козлова", "Алексей Смирнов"}
	cities := []string{"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань"}
	products := []string{"Тушь для ресниц", "Помада", "Крем для лица", "Шампунь", "Духи"}
	brands := []string{"Vivienne Sabo", "L'Oreal", "Maybelline", "Nivea", "Chanel"}
	currencies := []string{"USD", "EUR", "RUB"}
	banks := []string{"alpha", "sber", "vtb", "tinkoff", "gazprom"}

	for i := 1; i <= 10; i++ {
		orderUID := fmt.Sprintf("order_%d_%d", i, time.Now().Unix())
		trackNumber := fmt.Sprintf("TRACK%d%d", i, rand.Intn(1000))

		order := Order{
			OrderUID:          orderUID,
			TrackNumber:       trackNumber,
			Entry:             "WBIL",
			Locale:            "ru",
			InternalSignature: "",
			CustomerID:        fmt.Sprintf("customer_%d", i),
			DeliveryService:   "meest",
			ShardKey:          fmt.Sprintf("%d", rand.Intn(10)),
			SmID:              99 + i,
			DateCreated:       time.Now().Add(-time.Duration(i) * time.Hour),
			OofShard:          "1",
			Delivery: Delivery{
				Name:    names[rand.Intn(len(names))],
				Phone:   fmt.Sprintf("+7900%07d", rand.Intn(10000000)),
				Zip:     fmt.Sprintf("%06d", rand.Intn(1000000)),
				City:    cities[rand.Intn(len(cities))],
				Address: fmt.Sprintf("ул. Тестовая, д. %d", rand.Intn(100)),
				Region:  "Центральный",
				Email:   fmt.Sprintf("user%d@example.com", i),
			},
			Payment: Payment{
				Transaction:  orderUID,
				RequestID:    "",
				Currency:     currencies[rand.Intn(len(currencies))],
				Provider:     "wbpay",
				Amount:       rand.Intn(5000) + 500,
				PaymentDt:    time.Now().Unix(),
				Bank:         banks[rand.Intn(len(banks))],
				DeliveryCost: rand.Intn(500) + 200,
				GoodsTotal:   rand.Intn(4000) + 300,
				CustomFee:    0,
			},
			Items: []Item{
				{
					ChrtID:      9934930 + i,
					TrackNumber: trackNumber,
					Price:       rand.Intn(1000) + 100,
					RID:         fmt.Sprintf("rid_%d_%d", i, rand.Intn(1000)),
					Name:        products[rand.Intn(len(products))],
					Sale:        rand.Intn(50),
					Size:        "0",
					TotalPrice:  rand.Intn(800) + 100,
					NmID:        2389212 + i,
					Brand:       brands[rand.Intn(len(brands))],
					Status:      202,
				},
			},
		}

		orderJSON, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order %d: %v", i, err)
			continue
		}

		err = writer.WriteMessages(ctx, kafka.Message{
			Value: orderJSON,
		})
		if err != nil {
			log.Printf("Failed to send order %d: %v", i, err)
			continue
		}

		log.Printf("Successfully sent order %d: %s", i, orderUID)

		time.Sleep(100 * time.Millisecond)
	}

	log.Println("All orders sent successfully!")
}
