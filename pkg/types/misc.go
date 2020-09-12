package types

import (
	"encoding/json"
)

// StringSlice holds a number of strings
type StringSlice []string

// Unmarshal unpacks JSON to slice of strings
// e.g. [\"PetStore5\",\"PizzaShop1\"] to []string
//
func (s StringSlice) Unmarshal(jsonArrayOfStrings string) StringSlice {

	if jsonArrayOfStrings != "" {
		var StringValues []string
		err := json.Unmarshal([]byte(jsonArrayOfStrings), &StringValues)
		if err == nil {
			return StringValues
		}
	}
	return StringSlice{}
}

// Marshal packs slice of strings into JSON
// e.g. []string to [\"PetStore5\",\"PizzaShop1\"]
//
func (s StringSlice) Marshal() string {

	if len(s) > 0 {
		ArrayOfStringsInJSON, err := json.Marshal(s)
		if err == nil {
			return string(ArrayOfStringsInJSON)
		}
	}
	return "[]"
}
