package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status    string `json:"Status"`
		Timestamp int64  `json:"Timestamp"`
	}{
		Status:    "UP",
		Timestamp: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
