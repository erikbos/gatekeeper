package cassandra

import (
	"encoding/json"
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
