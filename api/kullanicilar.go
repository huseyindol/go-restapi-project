package handler

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
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

func Handler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Veritabanı bağlantı bilgisi bulunamadı", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, "Veritabanı bağlantı hatası", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// URL'den ID parametresini al
	path := r.URL.Path
	pathParts := strings.Split(path, "/")
	var id int

	// /api/users/123 gibi bir URL için ID'yi al
	if len(pathParts) > 2 {
		idStr := pathParts[2]
		if idStr != "" {
			id, _ = strconv.Atoi(idStr)
		}
	}

	switch r.Method {
	case "GET":
		if id > 0 {
			// Tek kullanıcı getir
			var k User
			err := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&k.ID, &k.Name, &k.Email)
			if err != nil {
				http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(k)
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
				var k User
				if err := rows.Scan(&k.ID, &k.Name, &k.Email); err != nil {
					http.Error(w, "Veri okuma hatası", http.StatusInternalServerError)
					return
				}
				users = append(users, k)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
		}

	case "POST":
		// Yeni kullanıcı ekle
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "İstek gövdesi okunamadı", http.StatusBadRequest)
			return
		}

		var k User
		if err := json.Unmarshal(body, &k); err != nil {
			http.Error(w, "Geçersiz JSON formatı", http.StatusBadRequest)
			return
		}

		var newID int
		err = db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", k.Name, k.Email).Scan(&newID)
		if err != nil {
			http.Error(w, "Kullanıcı eklenemedi", http.StatusInternalServerError)
			return
		}

		k.ID = newID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(k)

	case "PUT":
		if id <= 0 {
			http.Error(w, "Geçersiz kullanıcı ID", http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "İstek gövdesi okunamadı", http.StatusBadRequest)
			return
		}

		var k User
		if err := json.Unmarshal(body, &k); err != nil {
			http.Error(w, "Geçersiz JSON formatı", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", k.Name, k.Email, id)
		if err != nil {
			http.Error(w, "Kullanıcı güncellenemedi", http.StatusInternalServerError)
			return
		}

		k.ID = id
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(k)

	case "DELETE":
		if id <= 0 {
			http.Error(w, "Geçersiz kullanıcı ID", http.StatusBadRequest)
			return
		}

		result, err := db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Kullanıcı silinemedi", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Desteklenmeyen metod", http.StatusMethodNotAllowed)
	}
}
