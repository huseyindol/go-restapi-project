package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type DBResponse struct {
	Version string `json:"version"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Get connection string from environment variable
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgresql://user:password@neon_hostname/dbname?sslmode=require"
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
	response := DBResponse{
		Version: version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
