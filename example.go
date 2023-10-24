package main

import (
	"encoding/json"
	"log"

	stan "github.com/nats-io/stan.go"
)

// Публикация данных заказа в NATS
func DemonstratePublishToNATS(sc stan.Conn) {
	sampleData := Order{
		OrderUID:    "345566",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: &Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com"},
		Payment: &Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "test1",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0},
		Items: []Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
			{
				ChrtID:      523489234,
				TrackNumber: "FRMEKFKE",
				Price:       357,
				RID:         "ac2222222a764ae0btest",
				Name:        "Mqwerjkhw",
				Sale:        21,
				Size:        "1",
				TotalPrice:  552,
				NmID:        938287,
				Brand:       "Qwerty Uiop",
				Status:      200,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       "2021-11-26T06:22:19Z",
		OofShard:          "1"}

	// Попытка публикации sampleData в NATS
	err := PublishToNATS(sc, "myChannel", sampleData)
	if err != nil {
		log.Printf("Error publishing to NATS: %v", err)
	}

	// Сериализация sampleData в JSON
	jsonData, err := json.Marshal(sampleData)
	if err != nil {
		log.Fatalf("Error marshaling data to JSON: %v", err)
		return
	}

	// Публикация JSON в NATS Streaming канал
	err = sc.Publish("myChannel", jsonData) // используется метод Publish из stan.Conn
	if err != nil {
		log.Printf("Error publishing to NATS: %v", err)
	}
}
