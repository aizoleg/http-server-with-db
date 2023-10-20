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

var cacheMutex sync.RWMutex
var cache = make(map[string]Order)

func InitializeNATSConnection() (stan.Conn, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	// Создаем соединение с NATS Streaming. "test-cluster" — это имя вашего кластера.
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

	sc, err := InitializeNATSConnection()
	if err != nil {
		log.Fatalf("Failed to initialize NATS connection: %v", err)
	}
	defer sc.Close()
	DemonstratePublishToNATS(sc)

	// Запуск NATS-клиента
	StartNATSClient(db)

	// После запуска NATS-клиента загрузите кэш из БД
	if err := LoadCacheFromDB(db); err != nil {
		log.Fatalf("Failed to load cache from database: %v", err)
	}

	// Запускаем HTTP-сервер в отдельной горутине
	go StartHTTPServer()

	// Ждем завершения сигнала
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Exiting...")
}
