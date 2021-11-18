package types

import "github.com/go-playground/validator/v10"

// APIProduct type contains everything about an API product
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type (
	APIProduct struct {
		// Name of apiproduct (not changable)
		Name string `validate:"required,min=1"`

		// Routegroup this apiproduct should match to
		RouteGroup string

		// List of paths this apiproduct applies to
		APIResources []string `binding:"required,min=1"`

		// List of scopes that apply to this product
		Scopes []string

		// Approval type of this apiproduct
		ApprovalType string

		// Attributes of this apiproduct
		Attributes Attributes

		// Friendly display name of route
		DisplayName string

		// Full description of this api product
		Description string

		// Comma separated list of policynames, to apply to requests
		Policies string

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this apiproduct
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this apiproduct
		LastModifiedBy string
	}

	// APIProducts holds one or more apiproducts
	APIProducts []APIProduct
)

var (
	// NullAPIProduct is an empty apiproduct type
	NullAPIProduct = APIProduct{}

	// NullAPIProducts is an empty apiproduct slice
	NullAPIProducts = APIProducts{}
)

// Validate checks if field values are set correct and are allowed
func (a *APIProduct) Validate() error {

	validate := validator.New()
	return validate.Struct(a)
}
