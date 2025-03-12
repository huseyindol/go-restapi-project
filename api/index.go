package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Kullanici struct {
	ID    int    `json:"id"`
	Isim  string `json:"isim"`
	Email string `json:"email"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
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
		rows, err := db.Query("SELECT id, isim, email FROM kullanicilar")
		if err != nil {
			http.Error(w, "Sorgu hatası", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		kullanicilar := []Kullanici{}
		for rows.Next() {
			var k Kullanici
			if err := rows.Scan(&k.ID, &k.Isim, &k.Email); err != nil {
				http.Error(w, "Veri okuma hatası", http.StatusInternalServerError)
				return
			}
			kullanicilar = append(kullanicilar, k)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kullanicilar)

	default:
		http.Error(w, "Desteklenmeyen metod", http.StatusMethodNotAllowed)
	}
}

