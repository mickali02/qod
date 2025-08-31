// Filename: cmd/api/helpers.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

//create an envelope
type envelope map[string]any

func (a *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	jsResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	jsResponse = append(jsResponse, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(jsResponse)
	if err != nil {
		return err
	}

	return nil
}

func (a *application) readJSON(w http.ResponseWriter, r *http.Request, destination any) error {
    // Use http.MaxBytesReader to limit the size of the request body to 250KB.
    maxBytes := 256_000
    r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

    // Initialize the JSON decoder and call the DisallowUnknownFields() method on it
    // before decoding. This means that if the JSON from the client includes any
    // return an error instead of just ignoring the field.
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()

    // Decode the request body into the destination.
    err := dec.Decode(destination)
    if err != nil {
        // If there is an error during decoding, start the triage...
        var syntaxError *json.SyntaxError
        var unmarshalTypeError *json.UnmarshalTypeError
        var invalidUnmarshalError *json.InvalidUnmarshalError
        var maxBytesError *http.MaxBytesError

        switch {
        // Use the errors.As() function to check whether the error has the type
        // *json.SyntaxError. If it does, then return a plain-english error message
        // which includes the location of the problem.
        case errors.As(err, &syntaxError):
            return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

        // In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
        // for syntax errors in the JSON.
        case errors.Is(err, io.ErrUnexpectedEOF):
            return errors.New("body contains badly-formed JSON")

        // Catch any *json.UnmarshalTypeError errors. These occur when the JSON
        // value is the wrong type for the target destination. If the error relates
        // to a specific field, then we include that in our error message to make
        // it easier for the client to debug.
        case errors.As(err, &unmarshalTypeError):
            if unmarshalTypeError.Field != "" {
                return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
            }
            return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

        // An io.EOF error will be returned by Decode() if the request body is empty.
        case errors.Is(err, io.EOF):
            return errors.New("body must not be empty")
            
        // If the JSON contains a field which cannot be mapped to the target destination
        // then Decode() will now return an error message in the format "json: unknown
        // field "<name>"". We check for this, extract the field name, and interpolate
        // it into our error message.
        case strings.HasPrefix(err.Error(), "json: unknown field "):
            fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
            return fmt.Errorf("body contains unknown key %s", fieldName)

        // Use the errors.As() function to check whether the error has the type
        // *http.MaxBytesError. If it does, then it means the request body exceeded our
        // size limit of 250KB and we return a clear error message.
        case errors.As(err, &maxBytesError):
             return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

        // A json.InvalidUnmarshalError error will be returned if we pass something
        // that is not a non-nil pointer to Decode(). We catch this and panic,
        // rather than returning it to the client.
        case errors.As(err, &invalidUnmarshalError):
            panic(err)

        // For anything else, return the error message as-is.
        default:
            return err
        }
    }

    // Call Decode() again, using a pointer to an empty anonymous struct as the
    // destination. If the request body only contained a single JSON value this will
    // return an io.EOF error. So if we get anything else, we know that there is
    // additional data in the request body, and we return our own custom error message.
    err = dec.Decode(&struct{}{})
    if !errors.Is(err, io.EOF) {
        return errors.New("body must only contain a single JSON value")
    }

    return nil
}