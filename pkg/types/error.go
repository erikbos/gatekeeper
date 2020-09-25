package types

import (
	"errors"
	"fmt"
)

var (
	// ErrItemNotFound indicates the item could not be found
	ErrItemNotFound = errors.New("Item could not be found")

	// ErrUpdateFailure indicates item could not be updated
	ErrUpdateFailure = errors.New("Could not update item")

	// ErrBadRequest indicates we do not understand to do the requested action
	ErrBadRequest = errors.New("Bad request")

	// ErrPermissionDenied indicates we are not allowed to do the requested action
	ErrPermissionDenied = errors.New("Permission denied")

	// ErrDatabaseIssue indicates a database error
	ErrDatabaseIssue = errors.New("Database issue")
)

// Error is our  error type providing additional detail
// on what happened
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

// ErrorDetails returns the details of the occured error
func (e *errDetails) ErrorDetails() string {
	return fmt.Sprintf("%s", e.details)
}

// ErrorDetails returns the type of error
func (e *errDetails) Type() error {
	return e.errtype
}

// ErrorDetails returns the printable
func (e *errDetails) TypeString() string {
	return e.errtype.Error()
}

func newError(err error, details error) Error {
	return &errDetails{
		errtype: err,
		details: details.Error(),
	}
}

// NewItemNotFoundError returns a item not found error
func NewItemNotFoundError(details error) Error {
	return newError(ErrItemNotFound, details)
}

// NewUpdateFailureError returns a item not found error
func NewUpdateFailureError(details error) Error {
	return newError(ErrUpdateFailure, details)
}

// NewBadRequestError returns a database error
func NewBadRequestError(details error) Error {
	return newError(ErrBadRequest, details)
}

// NewPermissionDeniedError returns a database error
func NewPermissionDeniedError(details error) Error {
	return newError(ErrPermissionDenied, details)
}

// NewDatabaseError returns a database error
func NewDatabaseError(details error) Error {
	return newError(ErrDatabaseIssue, details)
}
