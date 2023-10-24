package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// StartHTTPServer инициализирует и запускает HTTP-сервер на порту 8080
func StartHTTPServer() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/getorder", handleGetOrder)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleIndex обрабатывает корневой путь ("/") и отдает файл "index.html"
func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// handleGetOrder обрабатывает запросы к "/getorder" и возвращает информацию о заказе в формате JSON
func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что используется метод GET
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID заказа из параметров запроса
	orderID := r.URL.Query().Get("orderID")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	// Проверяем наличие заказа в кэше
	cacheMutex.RLock()
	order, ok := cache[orderID]
	// Логгирование содержимого кэша
	log.Println("Cache contents:", cache)
	cacheMutex.RUnlock()

	// Если заказа нет в кэше, возвращаем ошибку
	if !ok {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Преобразуем информацию о заказе в формат JSON
	jsonResponse, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Error processing order", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок ответа и отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
