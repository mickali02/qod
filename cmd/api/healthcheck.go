// Filename: cmd/api/healthcheck.go
package main

import (
	"net/http"
)

const version = "1.0.0"

func (a *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map to hold healthcheck data.
	data := envelope{
		"status":      "available",
		"system_info": map[string]string{
		"environment": a.config.env,
		"version":     version, 
		},
	}

	err := a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}


	w.Header().Set("Content-Type", "application/json")

}