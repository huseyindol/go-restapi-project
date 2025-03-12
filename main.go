package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// API endpoint'lerini tanımla
	http.HandleFunc("/api/users", handleUsers)
	http.HandleFunc("/api/db", handleDB)

	// Statik dosyalar için
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Sunucuyu başlat
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	// CORS başlıkları
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Veritabanı bağlantısı
	connStr := os.Getenv("POSTGRES_URL")
	if connStr == "" {
		// Yerel geliştirme için varsayılan bağlantı
		connStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "Veritabanı bağlantı hatası: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// URL'den ID parametresini al
	path := r.URL.Path
	pathParts := strings.Split(path, "/")
	var id int = 0

	if len(pathParts) > 2 && pathParts[2] != "" {
		idStr := pathParts[2]
		idVal, err := strconv.Atoi(idStr)
		if err == nil {
			id = idVal
		}
	}

	// HTTP metodu kontrolü
	switch r.Method {
	case "GET":
		if id > 0 {
			// Belirli bir kullanıcıyı getir
			var user User
			err := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &user.Email)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
				} else {
					http.Error(w, "Sorgu hatası", http.StatusInternalServerError)
				}
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
		} else {
			// Tüm kullanıcıları getir
			rows, err := db.Query("SELECT id, name, email FROM users")
			if err != nil {
				http.Error(w, "Sorgu hatası", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			users := []User{}
			for rows.Next() {
				var user User
				if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
					http.Error(w, "Veri okuma hatası", http.StatusInternalServerError)
					return
				}
				users = append(users, user)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
		}
	}
}

func handleDB(w http.ResponseWriter, r *http.Request) {
	// Get connection string from environment variable
	connStr := os.Getenv("POSTGRES_URL")
	if connStr == "" {
		// Yerel geliştirme için varsayılan bağlantı
		connStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query database
	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}

	// Return response
	response := struct {
		Version string `json:"version"`
	}{
		Version: version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
