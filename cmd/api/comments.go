// Filename: cmd/api/comments.go
package main

import (
	//"encoding/json"
	"fmt"
	"net/http"
	// import the data package which contains the definition for Comment
	//"github.com/mickali02/qod/internal/data" // Now this will be used!
)

func (a *application)createCommentHandler(w http.ResponseWriter,r *http.Request) { 
	// create a struct to hold a comment
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Content  string  `json:"content"`
		Author   string  `json:"author"`
	} 
	
	// perform the decoding
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	
	// for now display the result
	fmt.Fprintf(w, "%+v\n", incomingData)
	}