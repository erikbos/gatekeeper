package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Attribute is an array with attributes
//
// Field validation settings (binding) are validated with
// https://godoc.org/github.com/go-playground/validator
type Attribute struct {
	// Attribute name, minimum required length is 4
	Name string `json:"name" binding:"required,min=4"`

	// Attribute value, minimum required length is 1 as we do not want empty values
	Value string `json:"value" binding:"required,min=1"`
}

// Attributes holds one or more attributes
type Attributes []Attribute

var (
	// NullAttribute is an empty attribute type
	NullAttribute = Attribute{}

	// NullAttributes is an empty attributes slice
	NullAttributes = Attributes{}
)

// AttributeValue is the single attribute type we receive from API
type AttributeValue struct {
	// Attribute value, minimum required length is 1
	Value string `json:"value" binding:"required,min=1"`
}

// Get return one named attribute from attributes
func (attributes *Attributes) Get(name string) (string, Error) {

	for _, element := range *attributes {
		if element.Name == name {
			return element.Value, nil
		}
	}
	return "", NewItemNotFoundError(fmt.Errorf("Cannot find attribute '%s'", name))
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
func (attributes *Attributes) Set(attributeValue Attribute) (oldValue string, oldValuePresent bool) {

	updatedAttributes := Attributes{}

	for _, oldAttribute := range *attributes {
		// In case attribute exists overwrite it
		if oldAttribute.Name == attributeValue.Name {
			oldValuePresent = true
			oldValue = oldAttribute.Value

			updatedAttributes = append(updatedAttributes, attributeValue)
		} else {
			updatedAttributes = append(updatedAttributes, oldAttribute)
		}
	}
	// In case it is a new attribute append it
	if !oldValuePresent {
		updatedAttributes = append(updatedAttributes, attributeValue)
	}

	// Overwrite existing slice with new slice
	*attributes = updatedAttributes

	if oldValuePresent {
		return oldValue, true
	}
	return "", false
}

// SetMultiple updates or adds multiple attribute. Returns error in case of isses
func (attributes *Attributes) SetMultiple(attributeValues Attributes) Error {

	for _, attribute := range attributeValues {
		_, _ = attributes.Set(attribute)
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
			// Record this element as an encountered element.
			encountered[attribute.Name] = true
			// Append to result slice.
			updatedAttributes = append(updatedAttributes, attribute)
		}
	}

	// Sort slice by attribute name
	sort.SliceStable(updatedAttributes, func(i, j int) bool {
		return updatedAttributes[i].Name < updatedAttributes[j].Name
	})

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
	return "", NewItemNotFoundError(fmt.Errorf("Cannot delete attribute '%s'", name))
}

// Unmarshal unpacks JSON array of attributes
// Example input: [{"name":"Shoesize","value":"42"}, {"name":"Destination","value":"Mars"}]
func (attributes Attributes) Unmarshal(jsonArrayOfAttributes string) Attributes {

	if jsonArrayOfAttributes != "" {
		var attributes = make(Attributes, 0)
		if err := json.Unmarshal([]byte(jsonArrayOfAttributes), &attributes); err == nil {
			return attributes
		}
	}
	return NullAttributes
}

// Marshal packs slice of attributes into JSON
// Example output: [{"name":"Shoesize","value":"42"}, {"name":"Destination","value":"Mars"}]
func (attributes Attributes) Marshal() string {

	if len(attributes) > 0 {
		json, err := json.Marshal(attributes)
		if err == nil {
			return string(json)
		}
	}
	return "[]"
}
