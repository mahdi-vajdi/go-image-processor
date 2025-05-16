package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

func ResponseJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("error while enconding response json %v", err)
		}
	}
}

func ErrorJSON(w http.ResponseWriter, status int, message string) {
	ResponseJSON(w, status, map[string]string{"message": message})
}
