package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func StartHTTPServer() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/getorder", handleGetOrder)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("orderID")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	cacheMutex.RLock()
	order, ok := cache[orderID]
	log.Println("Cache contents:", cache) // Logging inside the read lock
	cacheMutex.RUnlock()

	if !ok {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	jsonResponse, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Error processing order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
