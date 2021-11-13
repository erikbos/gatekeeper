package types

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// errUpdateFailure indicates item cannot be updated
	errUpdateFailure = errors.New("cannot update item")

	// errBadRequest indicates we do not understand the requested action (400)
	errBadRequest = errors.New("bad request")

	// errUnauthorized indicates we are not allowed to do the requested action (401)
	errUnauthorized = errors.New("unauthorized request")

	// errForbidden indicates the requested action was understood but forbidden (403)
	errForbidden = errors.New("forbidden action")

	// errItemNotFound indicates the item cannot be found (404)
	errItemNotFound = errors.New("entity cannot be found")

	// errDatabaseIssue indicates a database error
	errDatabaseIssue = errors.New("database issue")
)

// Error is our error type providing additional (internal error detail
type Error interface {
	Error() string
	ErrorDetails() string
	Type() error
	TypeString() string
}

type errDetails struct {
	errtype error
	details string
}

// Error returns the default printable error
func (e *errDetails) Error() string {
	return fmt.Sprintf("%v: %s", e.errtype, e.details)
}

// ErrorDetails returns the (internal) details of the occured error
func (e *errDetails) ErrorDetails() string {
	return e.details
}

// ErrorDetails returns the type of error
func (e *errDetails) Type() error {
	return e.errtype
}

// ErrorDetails returns the printable error
func (e *errDetails) TypeString() string {
	return e.errtype.Error()
}

func newError(err error, details error) Error {

	if details != nil {
		return &errDetails{
			errtype: err,
			details: details.Error(),
		}
	}
	return &errDetails{errtype: err}
}

// NewUpdateFailureError returns a item not found error
func NewUpdateFailureError(details error) Error {
	return newError(errUpdateFailure, details)
}

// NewBadRequestError returns a bad request error
func NewBadRequestError(details error) Error {
	return newError(errBadRequest, details)
}

// NewUnauthorizedError returns an unauthorized error
func NewUnauthorizedError(details error) Error {
	return newError(errUnauthorized, details)
}

// NewForbiddenError returns a forbidden action error
func NewForbiddenError(details error) Error {
	return newError(errForbidden, details)
}

// NewItemNotFoundError returns a item not found error
func NewItemNotFoundError(details error) Error {
	return newError(errItemNotFound, details)
}

// NewDatabaseError returns a database error
func NewDatabaseError(details error) Error {
	return newError(errDatabaseIssue, details)
}

// HTTPStatusCode returns HTTP status code for Error type
func HTTPStatusCode(e Error) int {

	switch e.Type() {
	case errUpdateFailure:
		return http.StatusServiceUnavailable

	case errBadRequest:
		return http.StatusBadRequest

	case errUnauthorized:
		return http.StatusForbidden

	case errForbidden:
		return http.StatusForbidden

	case errItemNotFound:
		return http.StatusNotFound

	case errDatabaseIssue:
		return http.StatusServiceUnavailable

	default:
		return http.StatusServiceUnavailable
	}
}
