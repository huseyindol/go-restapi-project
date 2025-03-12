package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type UserModel struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	// Vercel'den PostgreSQL bağlantı bilgilerini al
	connStr := os.Getenv("POSTGRES_URL")
	if connStr == "" {
		http.Error(w, "Veritabanı bağlantı bilgisi bulunamadı", http.StatusInternalServerError)
		return
	}

	// Veritabanına bağlan
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "Veritabanı bağlantı hatası", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// HTTP metodu kontrolü
	switch r.Method {
	case "GET":
		// Kullanıcıları getir
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			http.Error(w, "Sorgu hatası", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var k User
			if err := rows.Scan(&k.ID, &k.Name, &k.Email); err != nil {
				http.Error(w, "Veri okuma hatası", http.StatusInternalServerError)
				return
			}
			users = append(users, k)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)

	default:
		http.Error(w, "Desteklenmeyen metod", http.StatusMethodNotAllowed)
	}
}
