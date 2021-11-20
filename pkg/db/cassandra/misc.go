package cassandra

import (
	"encoding/json"

	"github.com/erikbos/gatekeeper/pkg/types"
)

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

// columnToString returns key in map as string, if key exists
func columnToString(m map[string]interface{}, columnName string) string {

	if m != nil {
		if val, ok := m[columnName]; ok {
			return val.(string)
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

// columnToAttributes changes Cassandra column type "map<text, text>" returned
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
