package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nats-io/stan.go"

	"github.com/nats-io/nats.go"
)

// Глобальные переменные для кэширования данных
var cacheMutex sync.RWMutex
var cache = make(map[string]Order)

// Инициализация соединения с NATS и возвращение этого соединения
func InitializeNATSConnection() (stan.Conn, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	// Создаем соединение с NATS Streaming
	sc, err := stan.Connect("test-cluster", "your-client-id", stan.NatsConn(nc))
	if err != nil {
		nc.Close()
		return nil, err
	}

	return sc, nil
}

func main() {
	log.Println("Application started successfully!")

	// Подключение к БД
	db, err := ConnectToDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация соединения с NATS
	sc, err := InitializeNATSConnection()
	if err != nil {
		log.Fatalf("Failed to initialize NATS connection: %v", err)
	}
	defer sc.Close()
	// Публикация в NATS
	DemonstratePublishToNATS(sc)

	// Запуск NATS-клиента для прослушивания сообщений
	StartNATSClient(db)

	// После запуска NATS-клиента загрузка кэша из БД
	if err := LoadCacheFromDB(db); err != nil {
		log.Fatalf("Failed to load cache from database: %v", err)
	}

	// Запуск HTTP-сервера в отдельной горутине
	go StartHTTPServer()

	// Ждем завершения сигнала
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Exiting...")
}
