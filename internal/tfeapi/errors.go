package tfeapi

import (
	"errors"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/tofutf/tofutf/internal"
)

func lookupHTTPCode(err error) int {
	if errors.Is(err, internal.ErrResourceNotFound) {
		return http.StatusNotFound
	} else if errors.Is(err, internal.ErrAccessNotPermitted) {
		return http.StatusForbidden
	} else if errors.Is(err, internal.ErrInvalidTerraformVersion) {
		return http.StatusUnprocessableEntity
	} else if errors.Is(err, internal.ErrResourceAlreadyExists) {
		return http.StatusConflict
	} else if errors.Is(err, internal.ErrConflict) {
		return http.StatusConflict
	}

	return http.StatusInternalServerError
}

// Error writes an HTTP response with a JSON-API encoded error.
func Error(w http.ResponseWriter, err error) {
	var (
		httpError *internal.HTTPError
		missing   *internal.MissingParameterError
		code      int
	)
	// If error is type internal.HTTPError then extract its status code
	if errors.As(err, &httpError) {
		code = httpError.Code
	} else if errors.As(err, &missing) {
		// report missing parameter errors as a 422
		code = http.StatusUnprocessableEntity
	} else {
		code = lookupHTTPCode(err)
	}
	b, err := jsonapi.Marshal(&jsonapi.Error{
		Status: &code,
		Title:  http.StatusText(code),
		Detail: err.Error(),
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(code)
	w.Write(b) //nolint:errcheck
}
