// Filename: cmd/api/healthcheck.go
package main

import (
	"encoding/json"
	"net/http"
)

const version = "1.0.0"

func (a *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map to hold healthcheck data.
	data := map[string]string{
		"status":      "available",
		"environment": a.config.env,
		"version":     version, 
	}

	jsResponse, err := json.Marshal(data)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		return
	}

	jsResponse = append(jsResponse, '\n')

	w.Header().Set("Content-Type", "application/json")

	w.Write(jsResponse)
}