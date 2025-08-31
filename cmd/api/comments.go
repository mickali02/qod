// Filename: cmd/api/comments.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	// import the data package which contains the definition for Comment
	"github.com/mickali02/qod/internal/data" // Now this will be used!
)

func (a *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	// create a temporary struct to hold the incoming data
	var incomingData struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}

	// perform the decoding
	err := json.NewDecoder(r.Body).Decode(&incomingData)
	if err != nil {
		// You will create this badRequestResponse function in a later step
		// a.badRequestResponse(w, r, err) 
		// For now, this is fine:
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Create an instance of your Comment struct from the 'data' package.
	// By writing 'data.Comment', you are now using the imported package.
	comment := &data.Comment{
		Content: incomingData.Content,
		Author:  incomingData.Author,
	}

	// for now display the result (now using the 'comment' variable)
	fmt.Fprintf(w, "%+v\n", comment)
}