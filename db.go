package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

func ConnectToDB() (*sql.DB, error) {
	connStr := "user=admin dbname=postgres sslmode=disable password=root host=127.0.0.1 port=5432"
	return sql.Open("postgres", connStr)
}

// Функция для вставки данных в таблицу orders
func InsertOrderToDB(db *sql.DB, orderData *Order) error {
	query := `
        INSERT INTO orders (OrderUID, TrackNumber, Entry, Locale, InternalSignature, CustomerID, DeliveryService, Shardkey, SmID, DateCreated, OofShard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

	_, err := db.Exec(query, orderData.OrderUID, orderData.TrackNumber, orderData.Entry, orderData.Locale, orderData.InternalSignature, orderData.CustomerID, orderData.DeliveryService, orderData.Shardkey, orderData.SmID, orderData.DateCreated, orderData.OofShard)
	return err
}

// Функция для вставки данных в таблицу delivery
func InsertDeliveryToDB(db *sql.DB, orderUID string, deliveryData *Delivery) error {
	query := `
        INSERT INTO delivery (order_id, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := db.Exec(query, orderUID, deliveryData.Name, deliveryData.Phone, deliveryData.Zip, deliveryData.City, deliveryData.Address, deliveryData.Region, deliveryData.Email)
	return err
}

// Функция для вставки данных в таблицу payment
func InsertPaymentToDB(db *sql.DB, orderUID string, paymentData *Payment) error {
	query := `
    INSERT INTO payment (order_id, transaction, requestid, currency, provider, amount, paymentdt, bank, deliverycost, goodstotal, customfee) 
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`

	_, err := db.Exec(query, orderUID, paymentData.Transaction, paymentData.RequestID, paymentData.Currency, paymentData.Provider, paymentData.Amount, paymentData.PaymentDT, paymentData.Bank, paymentData.DeliveryCost, paymentData.GoodsTotal, paymentData.CustomFee)
	return err
}

// Функция для вставки данных в таблицу item
func InsertItemsToDB(db *sql.DB, orderUID string, itemsData []Item) error {
	query := `
        INSERT INTO item (order_id, chrtID, trackNumber, price, rid, name, sale, size, totalPrice, nmID, brand, status) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

	// Так как может быть несколько товаров, используем цикл
	for _, item := range itemsData {
		_, err := db.Exec(query, orderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

// Функция, извлекающая все заказы из БД
func GetAllOrdersFromDB(db *sql.DB) ([]Order, error) {
	ordersQuery := `
    SELECT ID, OrderUID, TrackNumber, Entry, Locale, InternalSignature, CustomerID, 
	DeliveryService, Shardkey, SmID, DateCreated, OofShard
    FROM orders
    `
	rows, err := db.Query(ordersQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order

	for rows.Next() {
		var order Order
		err = rows.Scan(
			&order.ID, &order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
		if err != nil {
			return nil, err
		}

		// Запрос информации о доставке для этого заказа
		deliveryQuery := `
        SELECT name, phone, zip, city, address, region, email 
        FROM delivery 
        WHERE order_id = $1
        `
		deliveryRow := db.QueryRow(deliveryQuery, order.OrderUID)
		var delivery Delivery
		err = deliveryRow.Scan(&delivery.Name, &delivery.Phone, &delivery.Zip,
			&delivery.City, &delivery.Address, &delivery.Region, &delivery.Email)
		if err == nil {
			order.Delivery = &delivery
		}

		// Запрос информации о платеже для этого заказа
		paymentQuery := `
        SELECT transaction, requestID, currency, provider, amount, paymentDt, 
        bank, deliveryCost, goodsTotal, customFee
        FROM payment 
        WHERE order_id = $1
        `
		paymentRow := db.QueryRow(paymentQuery, order.OrderUID)
		var payment Payment
		err = paymentRow.Scan(&payment.Transaction, &payment.RequestID, &payment.Currency,
			&payment.Provider, &payment.Amount, &payment.PaymentDT, &payment.Bank,
			&payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
		if err != nil {
			log.Printf("Error scanning payment data: %v", err)
		} else {
			order.Payment = &payment
		}

		// Запрос товаров для этого заказа
		itemsQuery := `
        SELECT ChrtID, TrackNumber, Price, Rid, Name, Sale, Size, 
        TotalPrice, NmID, Brand, Status
        FROM item 
        WHERE order_id = $1
        `
		itemsRows, err := db.Query(itemsQuery, order.OrderUID)
		if err != nil {
			return nil, err
		}
		defer itemsRows.Close() // Добавляем закрытие rows после использования

		var items []Item
		for itemsRows.Next() {
			var item Item
			err = itemsRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}

		// Проверка на ошибки после цикла
		if err = itemsRows.Err(); err != nil {
			return nil, err
		}

		order.Items = items
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// Функция, заполняющая кэш
func LoadCacheFromDB(db *sql.DB) error {
	orders, err := GetAllOrdersFromDB(db)
	if err != nil {
		return err
	}

	cacheMutex.Lock()
	for _, order := range orders {
		cache[order.OrderUID] = order
	}
	cacheMutex.Unlock()

	return nil
}
