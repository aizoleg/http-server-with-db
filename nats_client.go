package main

import (
	"database/sql"
	"encoding/json"
	"log"

	stan "github.com/nats-io/stan.go"
)

const (
	// Константы для подключения к NATS
	clusterID = "test-cluster"
	clientID  = "0000"
	natsURL   = "nats://localhost:4222"
)

// Инициализация клиента NATS и подписка на канал для получения сообщений
func StartNATSClient(db *sql.DB) {
	// Создание подключения
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Fatalf("Failed to connect to NATS Streaming server: %v", err)
	}
	defer sc.Close()

	// Подписка на канал и обработка сообщений
	_, err = sc.Subscribe("myChannel", func(m *stan.Msg) {
		log.Printf("Received a message: %s\n", string(m.Data))

		// Десериализация данных
		var orderData Order
		err := json.Unmarshal(m.Data, &orderData)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return
		}

		// Вставка данных заказа в БД
		if err = InsertOrderToDB(db, &orderData); err != nil {
			log.Printf("Failed to insert order: %v", err)
			return
		}

		// Вставка данных доставки в БД
		if err = InsertDeliveryToDB(db, orderData.OrderUID, orderData.Delivery); err != nil {
			log.Printf("Failed to insert delivery: %v", err)
			return
		}
		// Вставка данных оплаты в БД
		if err = InsertPaymentToDB(db, orderData.OrderUID, orderData.Payment); err != nil {
			log.Printf("Failed to insert payment: %v", err)
			return
		}

		// Вставка товаров заказа в БД
		if err = InsertItemsToDB(db, orderData.OrderUID, orderData.Items); err != nil {
			log.Printf("Failed to insert items: %v", err)
			return
		}

		// Вставка данных заказа в кэш
		cacheMutex.Lock()
		cache[orderData.OrderUID] = orderData
		cacheMutex.Unlock()

		log.Printf("Received order with ID: %s", orderData.OrderUID)
		// Выводим лог об успешной записи в БД
		log.Println("Data has been successfully written to the database:", orderData)

	}, stan.DurableName("myDurableName"))

	if err != nil {
		log.Fatalf("Failed to subscribe to channel: %v", err)
	}
}

// Публикация данных в NATS после их сериализации в JSON
func PublishToNATS(sc stan.Conn, channel string, data interface{}) error {
	// Сериализация данных в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Публикация сериализованных данных в NATS
	return sc.Publish(channel, jsonData)
}
