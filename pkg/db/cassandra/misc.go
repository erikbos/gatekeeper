package cassandra

import (
	"encoding/json"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// stringSliceUnmarshal unpacks JSON to slice of strings
// e.g. [\"PetStore5\",\"PizzaShop1\"] to []string
func stringSliceUnmarshal(jsonArrayOfStrings string) []string {

	if jsonArrayOfStrings != "" {
		var StringValues []string
		err := json.Unmarshal([]byte(jsonArrayOfStrings), &StringValues)
		if err == nil {
			return StringValues
		}
	}
	return []string{}
}

// stringSliceMarshal packs slice of strings into JSON
// e.g. []string to [\"PetStore5\",\"PizzaShop1\"]
func stringSliceMarshal(s []string) string {

	if len(s) > 0 {
		SliceStringsInJSON, err := json.Marshal(s)
		if err == nil {
			return string(SliceStringsInJSON)
		}
	}
	return "[]"
}

// Unmarshal unpacks JSON array of attributes
// Example input: [{"name":"Shoesize","value":"42"}, {"name":"Destination","value":"Mars"}]
func AttributesUnmarshal(jsonArrayOfAttributes string) types.Attributes {

	if jsonArrayOfAttributes != "" {
		var attributes = make(types.Attributes, 0)
		if err := json.Unmarshal([]byte(jsonArrayOfAttributes), &attributes); err == nil {
			return attributes
		}
	}
	return types.NullAttributes
}

// Marshal packs slice of attributes into JSON
// Example output: [{"name":"Shoesize","value":"42"}, {"name":"Destination","value":"Mars"}]
func AttributesMarshal(attributes types.Attributes) string {

	if len(attributes) > 0 {
		json, err := json.Marshal(attributes)
		if err == nil {
			return string(json)
		}
	}
	return "[]"
}

// AllowsUnmarshal unpacks JSON-encoded role allows into Allows
func AllowsUnmarshal(roleAllowsAsJSON string) types.Allows {

	if roleAllowsAsJSON != "" {
		var allows types.Allows
		if err := json.Unmarshal([]byte(roleAllowsAsJSON), &allows); err == nil {
			return allows
		}
	}
	return types.NullAllows
}

// AllowsMarshal packs role Allows into JSON
func AllowsMarshal(a types.Allows) string {

	if json, err := json.Marshal(a); err == nil {
		return string(json)
	}
	return "[]"
}

// Unmarshal unpacks a key's product statuses
// Example input: [{"name":"S","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
func KeyAPIProductStatusesUnmarshal(jsonProductStatuses string) types.KeyAPIProductStatuses {

	if jsonProductStatuses != "" {
		var productStatus = make([]types.KeyAPIProductStatus, 0)
		if err := json.Unmarshal([]byte(jsonProductStatuses), &productStatus); err == nil {
			return productStatus
		}
	}
	return types.KeyAPIProductStatuses{}
}

// Marshal packs a key's product statuses into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
func KeyAPIProductStatusesMarshal(ps types.KeyAPIProductStatuses) string {

	if len(ps) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ps)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}
