package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Order struct {
	ID        int     `json:"id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
}

var db *sql.DB

func main() {
	var err error
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:8706)/ecommerce_order?parseTime=true"
	}
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Cannot connect to db yet: %v, retrying...", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders", getOrders)
	mux.HandleFunc("POST /orders", createOrder)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	fmt.Printf("Order Service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(mux)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, product_id, quantity, total FROM orders")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.ProductID, &o.Quantity, &o.Total); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []Order{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	var o Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := db.Exec("INSERT INTO orders (product_id, quantity, total) VALUES (?, ?, ?)", o.ProductID, o.Quantity, o.Total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	o.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}
