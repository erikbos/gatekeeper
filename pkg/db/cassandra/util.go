package cassandra

import (
	"encoding/json"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// UnmarshallJSONArrayOfStrings unpacks JSON array of strings
// e.g. [\"PetStore5\",\"PizzaShop1\"] to []string
//
func (d *Database) UnmarshallJSONArrayOfStrings(jsonArrayOfStrings string) []string {
	if jsonArrayOfStrings != "" {
		var StringValues []string
		err := json.Unmarshal([]byte(jsonArrayOfStrings), &StringValues)
		if err == nil {
			return StringValues
		}
	}
	return nil
}

// MarshallArrayOfStringsToJSON packs array of string into JSON
// e.g. []string to [\"PetStore5\",\"PizzaShop1\"]
//
func (d *Database) MarshallArrayOfStringsToJSON(ArrayOfStrings []string) string {
	if len(ArrayOfStrings) > 0 {
		ArrayOfStringsInJSON, err := json.Marshal(ArrayOfStrings)
		if err == nil {
			return string(ArrayOfStringsInJSON)
		}
	}
	return "[]"
}

// UnmarshallJSONArrayOfAttributes unpacks JSON array of attribute bags
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) UnmarshallJSONArrayOfAttributes(jsonArrayOfAttributes string) []shared.AttributeKeyValues {
	if jsonArrayOfAttributes != "" {
		var ResponseAttributes = make([]shared.AttributeKeyValues, 0)
		if err := json.Unmarshal([]byte(jsonArrayOfAttributes), &ResponseAttributes); err == nil {
			return ResponseAttributes
		}
	}
	return nil
}

// MarshallArrayOfAttributesToJSON packs array of attributes into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) MarshallArrayOfAttributesToJSON(ArrayOfAttributes []shared.AttributeKeyValues) string {

	if len(ArrayOfAttributes) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ArrayOfAttributes)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}

// UnmarshallJSONArrayOfProductStatuses unpacks JSON array of attribute bags
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) UnmarshallJSONArrayOfProductStatuses(jsonArrayOfAttributes string) []shared.APIProductStatus {
	if jsonArrayOfAttributes != "" {
		var ResponseAttributes = make([]shared.APIProductStatus, 0)
		if err := json.Unmarshal([]byte(jsonArrayOfAttributes), &ResponseAttributes); err == nil {
			return ResponseAttributes
		}
	}
	return nil
}

// MarshallArrayOfProductStatusesToJSON packs array of attributes into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
//
func (d *Database) MarshallArrayOfProductStatusesToJSON(ArrayOfAttributes []shared.APIProductStatus) string {

	if len(ArrayOfAttributes) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ArrayOfAttributes)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}
