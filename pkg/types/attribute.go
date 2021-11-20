package types

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Attribute is an array with attributes
//
// Field validation settings (binding) are validated with
// https://godoc.org/github.com/go-playground/validator
type (
	Attribute struct {
		// Attribute name, minimum required length is 1
		Name string `validate:"required,min=1"`

		// Attribute value
		Value string `validate:"required"`
	}

	// Attributes holds one or more attributes
	Attributes []Attribute
)

var (
	// NullAttribute is an empty attribute type
	NullAttribute = Attribute{}

	// NullAttributes is an empty attributes slice
	NullAttributes = Attributes{}

	// Maximum number of attributes allowed in set
	MaximumNumberofAttributesAllowed = 100

	errAttributeNotFound = NewItemNotFoundError(fmt.Errorf("cannot find attribute"))

	errTooManyAttributes = NewUpdateFailureError(errors.New("cannot add more than 100 attributes"))
)

// NewAttribute creates a new attribute
func NewAttribute(name, value string) *Attribute {

	return &Attribute{Name: name, Value: value}
}

// Get return one named attribute from attributes
func (attributes *Attributes) Get(name string) (string, Error) {

	for _, element := range *attributes {
		if element.Name == name {
			return element.Value, nil
		}
	}
	return "", errAttributeNotFound
}

// GetAsString returns attribute value (or provided default if not found) as type string
func (attributes *Attributes) GetAsString(name, defaultValue string) string {

	value, err := attributes.Get(name)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetAsUInt32 returns attribute value (or provided default) as type integer
func (attributes *Attributes) GetAsUInt32(name string, defaultValue uint32) uint32 {

	value, err := attributes.Get(name)
	if err == nil {
		integer, err := strconv.ParseUint(value, 10, 32)
		if err == nil {
			return uint32(integer)
		}
	}
	return defaultValue
}

// GetAsDuration returns attribute value (or provided default) as type time.Duration
func (attributes *Attributes) GetAsDuration(name string, defaultDuration time.Duration) time.Duration {

	value, err := attributes.Get(name)
	if err == nil {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultDuration
}

// Set updates or adds attribute in slice. Returns old value if attribute already existed.
func (attributes *Attributes) Set(attributeValue *Attribute) Error {

	if len(*attributes) > MaximumNumberofAttributesAllowed {
		return errTooManyAttributes
	}

	updatedAttributes := Attributes{}
	attributePresent := false

	for _, oldAttribute := range *attributes {
		// In case attribute exists append new value
		if oldAttribute.Name == attributeValue.Name {
			attributePresent = true
			updatedAttributes = append(updatedAttributes, *attributeValue)
		} else {
			updatedAttributes = append(updatedAttributes, oldAttribute)
		}
	}
	// In case it is a new attribute append it
	if !attributePresent {
		updatedAttributes = append(updatedAttributes, *attributeValue)
	}
	// Overwrite existing slice with new slice
	*attributes = updatedAttributes

	return nil
}

// SetMultiple updates or adds multiple attribute. Returns error in case of isses
func (attributes *Attributes) SetMultiple(attributeValues Attributes) Error {

	for _, attribute := range attributeValues {
		if err := attributes.Set(&attribute); err != nil {
			return err
		}
	}
	return nil
}

// Tidy removes duplicate, trims all names & values and sorts attribute by name
func (attributes *Attributes) Tidy() {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	updatedAttributes := Attributes{}

	// Remove dups & trim
	for _, attribute := range *attributes {
		if encountered[strings.TrimSpace(attribute.Name)] {
			// Do not add duplicate.
		} else {
			// Trim whitespace we like tidy
			attribute.Name = strings.TrimSpace(attribute.Name)
			attribute.Value = strings.TrimSpace(attribute.Value)
			// Record this attribute as an encountered attribute.
			encountered[attribute.Name] = true
			// Append to result slice.
			updatedAttributes = append(updatedAttributes, attribute)
		}
	}
	updatedAttributes.Sort()
	*attributes = updatedAttributes
}

// Delete removes attribute from slice. Returns delete status and deleted attribute's value
func (attributes *Attributes) Delete(name string) (valueOfDeletedAttribute string, e Error) {

	updatedAttributes := Attributes{}

	var attributeDeleted bool

	for _, attribute := range *attributes {
		if attribute.Name == name {
			valueOfDeletedAttribute = attribute.Value
			attributeDeleted = true
		} else {
			// No match, keep attribute
			updatedAttributes = append(updatedAttributes, attribute)
		}
	}
	*attributes = updatedAttributes

	if attributeDeleted {
		return valueOfDeletedAttribute, nil
	}
	return "", errAttributeNotFound
}

// Sort slice by attribute name
func (attributes Attributes) Sort() {

	// Sort slice by attribute name
	sort.SliceStable(attributes, func(i, j int) bool {
		return attributes[i].Name < attributes[j].Name
	})
}

// Validate checks if field values are set correct and are allowed
func (a *Attribute) Validate() error {

	validate := validator.New()
	return validate.Struct(a)
}

// Validate checks if field values are set correct and are allowed
func (a *Attributes) Validate() error {

	validate := validator.New()
	return validate.Struct(a)
}
