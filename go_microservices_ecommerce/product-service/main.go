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

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var db *sql.DB

func main() {
	var err error
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:8706)/ecommerce_product?parseTime=true"
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
	mux.HandleFunc("GET /products", getProducts)
	mux.HandleFunc("POST /products", createProduct)
	mux.HandleFunc("DELETE /products/{id}", deleteProduct)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("Product Service running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(mux)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, price FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}
	if products == nil {
		products = []Product{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := db.Exec("INSERT INTO products (name, price) VALUES (?, ?)", p.Name, p.Price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	p.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}
	_, err := db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
