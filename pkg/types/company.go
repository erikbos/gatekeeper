package types

import "github.com/go-playground/validator/v10"

// Company holds details of a company
type (
	Company struct {
		// Name of company (not changable)
		Name string `validate:"required,min=1,max=100"`

		// Friendly display name of company
		DisplayName string

		// Attributes of company
		Attributes Attributes

		// Status of company
		Status string

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this company
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this company
		LastModifiedBy string
	}

	// Companies holds one or more Companies
	Companies []Company
)

var (
	// NullCompany is an empty company type
	NullCompany = Company{}

	// NullCompanies is an empty company slice
	NullCompanies = Companies{}
)

// Validate checks if field values are set correct and are allowed
func (o *Company) Validate() error {

	validate := validator.New()
	return validate.Struct(o)
}

const (
	companyStatusActive   = "active"
	companyInStatusActive = "inactive"
)

// Activate marks a company as active
func (c *Company) Activate() {
	c.Status = companyStatusActive
}

// Deactivate marks a developer as inactive
func (c *Company) Deactivate() {
	c.Status = companyInStatusActive
}

// IsActive returns true in case company's status is active
func (c *Company) IsActive() bool {
	return c.Status == companyStatusActive
}
