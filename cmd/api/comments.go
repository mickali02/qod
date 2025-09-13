// Filename: cmd/api/comments.go
package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mickali02/qod/internal/data"
	"github.com/mickali02/qod/internal/validator"
)

func (a *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	comment := &data.Comment{
		Content: incomingData.Content,
		Author:  incomingData.Author,
	}

	v := validator.New()
	data.ValidateComment(v, comment)

	if !v.IsEmpty() {
		// You will need to add this helper function in errors.go later (slide 175)
		// a.failedValidationResponse(w, r, v.Errors)
		// For now, let's just use badRequest
		a.badRequestResponse(w, r, errors.New("validation failed"))
		return
	}

	err = a.commentModel.Insert(comment)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/comments/%d", comment.ID))

	response := envelope{"comment": comment}
	err = a.writeJSON(w, http.StatusCreated, response, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *application) displayCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	comment, err := a.commentModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	response := envelope{"comment": comment}
	err = a.writeJSON(w, http.StatusOK, response, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// This is the code from slides 210-215
func (a *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	comment, err := a.commentModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var incomingData struct {
		Content *string `json:"content"`
		Author  *string `json:"author"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Content != nil {
		comment.Content = *incomingData.Content
	}
	if incomingData.Author != nil {
		comment.Author = *incomingData.Author
	}

	v := validator.New()
	data.ValidateComment(v, comment)
	if !v.IsEmpty() {
		// As before, you will add this helper later.
		// a.failedValidationResponse(w, r, v.Errors)
		a.badRequestResponse(w, r, errors.New("validation failed"))
		return
	}

	err = a.commentModel.Update(comment)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	response := envelope{"comment": comment}
	err = a.writeJSON(w, http.StatusOK, response, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// This is the delete handler from slides 223-226
func (a *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.commentModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"message": "comment successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// Add this handler function to the end of your file
func (a *application) listCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// This struct will hold the query string values.
	var input struct {
		Content string
		Author  string
	}

	// Get the query parameters from the URL
	queryParameters := r.URL.Query()

	// Use your helper to extract the 'content' and 'author' query string values.
	// We are providing default empty strings "" if they are not provided.
	input.Content = a.getSingleQueryParameter(queryParameters, "content", "")
	input.Author = a.getSingleQueryParameter(queryParameters, "author", "")

	// Call the GetAll() method to retrieve the comments.
	comments, err := a.commentModel.GetAll(input.Content, input.Author)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send the JSON response with the list of comments.
	data := envelope{"comments": comments}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}