// Filename: cmd/api/comments.go
package main

import (
	"fmt"
	"net/http"
	// import the data package which contains the definition for Comment
	"github.com/mickali02/qod/internal/data"
	// "github.com/mickali02/qod/internal/validator" 
)

func (a *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	// create a struct to hold the incoming JSON data
	var incomingData struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}

	// perform the decoding
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Create a new data.Comment struct and populate it with the data from the request.
	// We pass a pointer to this struct to the Insert() method.
	comment := &data.Comment{
		Content: incomingData.Content,
		Author:  incomingData.Author,
	}

	// Add the comment to the database table.
	// The Insert() method will update the 'comment' variable with the ID, CreatedAt, and Version.
	err = a.commentModel.Insert(comment)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Set a Location header. The path to the newly created comment.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/comments/%d", comment.ID))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"comment": comment,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}