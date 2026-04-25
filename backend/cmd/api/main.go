package main

import (
	"encoding/json"
	"log"
	"net/http"

	"travelingzip/core"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":         "ok",
			"implementation": "go",
		})
	})
	mux.HandleFunc("/api/simulation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}
		defer r.Body.Close()

		var body json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		config, err := core.ParseConfig(body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, core.BuildSimulationSnapshot(config))
	})

	log.Println("Go API listening on http://127.0.0.1:3001")
	log.Fatal(http.ListenAndServe("127.0.0.1:3001", mux))
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
