package main

import (
	"errors"
	"net/http"
)

type apiError struct {
	Err    string
	Status int
}

func (e *apiError) Error() string {
	return e.Err
}

var (
	ErrExecTmpl         = &apiError{Err: "failed to exec tmpl", Status: http.StatusInternalServerError}
	ErrMethodNotAllowed = &apiError{Err: "method not allowed", Status: http.StatusMethodNotAllowed}

	ErrGetKey                = errors.New("server: failed to get the key")
	ErrConnStr               = errors.New("server: invalid connection string")
	ErrFetchingConfiguration = errors.New("server: error fetching configuration")
	ErrCreateKey             = errors.New("server: failed to create a key on the fly")
	ErrUnexpectedBody        = errors.New("server: unexpected configuration response body")
)
