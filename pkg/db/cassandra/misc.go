package cassandra

import (
	"encoding/json"

	"github.com/gocql/gocql"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// columnToString returns key in map as string, if key exists
func columnToString(m map[string]interface{}, columnName string) string {

	if m != nil {
		if val, ok := m[columnName]; ok {
			switch columnValue := val.(type) {
			case string:
				return columnValue
			case gocql.UUID:
				return columnValue.String()
			}
		}
	}
	return ""
}

// columnToInt returns key in map as int, if key exists
func columnToInt(m map[string]interface{}, columnName string) int {

	if m != nil {
		if val, ok := m[columnName]; ok {
			return val.(int)
		}
	}
	return -1
}

// columnToInt64 returns key in map as int64, if key exists
func columnToInt64(m map[string]interface{}, columnName string) int64 {

	if m != nil {
		if val, ok := m[columnName]; ok {
			return val.(int64)
		}
	}
	return -1
}

// columnToStringSlice returns key in map as []string, if key exists
func columnToStringSlice(m map[string]interface{}, columnName string) []string {

	if m != nil {
		if val, ok := m[columnName]; ok {
			return val.([]string)
		}
	}
	return []string{}
}

// columnToMapString returns key in map (encoding as JSON) as map[string]interface, if key exists
func columnToMapString(m map[string]interface{}, columnName string) map[string]interface{} {

	if m != nil {
		if val, ok := m[columnName]; ok {
			var v map[string]interface{}
			_ = json.Unmarshal([]byte(val.(string)), &v)
			return v
		}
	}
	return map[string]interface{}{}
}

// columnToAttributes converts Cassandra column type "map<text, text>" returned
// by gocql as map[string]string into types.Attributes
func columnToAttributes(m map[string]interface{}, columnName string) types.Attributes {

	if m != nil {
		if val, ok := m[columnName]; ok {
			attributes := types.Attributes{}
			for name, value := range val.(map[string]string) {
				attributes = append(attributes, *types.NewAttribute(name, value))
			}
			return attributes
		}
	}
	return types.NullAttributes
}

// attributesToColumn converts attributes to map[string]string
// so it can be stored in a map<text, text> column
func attributesToColumn(attributes types.Attributes) map[string]string {

	a := make(map[string]string, len(attributes))
	for i := range attributes {
		a[attributes[i].Name] = attributes[i].Value
	}
	return a
}

func valueToJSON(s interface{}) string {

	jsonEncoded, err := json.Marshal(s)
	if err == nil {
		return string(jsonEncoded)
	}
	return ""
}
