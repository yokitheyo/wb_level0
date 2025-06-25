package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8081/order/"

	fmt.Println("Получаем список заказов из кеша...")
	resp, err := http.Get("http://localhost:8081/cache/stats")
	if err != nil {
		fmt.Printf("Ошибка получения статистики кеша: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка чтения ответа: %v\n", err)
		return
	}

	fmt.Printf("Статистика кеша: %s\n\n", string(body))

	orderIDs := []string{
		"b563feb7b2b84b6test",
	}

	fmt.Println("=== Тест производительности кеша ===")
	fmt.Println()

	for _, orderID := range orderIDs {
		fmt.Printf("Тестируем заказ: %s\n", orderID)

		fmt.Print("  Первый запрос (из БД): ")
		start := time.Now()
		resp1, err := http.Get(baseURL + orderID)
		duration1 := time.Since(start)

		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}
		resp1.Body.Close()

		if resp1.StatusCode == 404 {
			fmt.Printf("Заказ не найден (404)\n")
			continue
		}

		fmt.Printf("%v\n", duration1)

		time.Sleep(100 * time.Millisecond)

		fmt.Print("  Второй запрос (из кеша): ")
		start = time.Now()
		resp2, err := http.Get(baseURL + orderID)
		duration2 := time.Since(start)

		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}
		resp2.Body.Close()

		fmt.Printf("%v\n", duration2)

		if duration2 > 0 {
			speedup := float64(duration1) / float64(duration2)
			fmt.Printf("  Ускорение: %.2fx\n", speedup)
		}

		fmt.Println()
	}

	fmt.Println("=== Тест множественных запросов к кешу ===")

	orderID := "b563feb7b2b84b6test"
	numRequests := 10

	fmt.Printf("Выполняем %d запросов к заказу %s\n", numRequests, orderID)

	var totalDuration time.Duration
	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := http.Get(baseURL + orderID)
		duration := time.Since(start)
		totalDuration += duration

		if err != nil {
			fmt.Printf("Запрос %d: Ошибка %v\n", i+1, err)
			continue
		}
		resp.Body.Close()

		fmt.Printf("Запрос %d: %v\n", i+1, duration)
	}

	avgDuration := totalDuration / time.Duration(numRequests)
	fmt.Printf("\nСреднее время ответа: %v\n", avgDuration)
}
