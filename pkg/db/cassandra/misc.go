package cassandra

import (
	"encoding/json"
	"log"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// columnToString returns key in map as string, if key exists
func columnToString(m map[string]interface{}, columnName string) string {

	if m != nil {
		if columnData, ok := m[columnName]; ok {
			switch columnValue := columnData.(type) {
			case string:
				return columnValue
			default:
				fatalWrongColumnValueType(columnData, columnName)
			}
		}
	}
	return ""
}

// columnToInt returns key in map as int, if key exists
func columnToInt(m map[string]interface{}, columnName string) int {

	if m != nil {
		if columnData, ok := m[columnName]; ok {
			switch columnValue := columnData.(type) {
			case int:
				return columnValue
			default:
				fatalWrongColumnValueType(columnData, columnName)
			}
		}
	}
	return -1
}

// columnToInt64 returns key in map as int64, if key exists
func columnToInt64(m map[string]interface{}, columnName string) int64 {

	if m != nil {
		if columnData, ok := m[columnName]; ok {
			switch columnValue := columnData.(type) {
			case int64:
				return columnValue
			default:
				fatalWrongColumnValueType(columnData, columnName)
			}
		}
	}
	return -1
}

// columnToStringSlice returns key in map as []string, if key exists
func columnToStringSlice(m map[string]interface{}, columnName string) []string {

	if m != nil {
		if columnData, ok := m[columnName]; ok {
			switch columnValue := columnData.(type) {
			case []string:
				return columnValue
			default:
				fatalWrongColumnValueType(columnData, columnName)
			}
		}
	}
	return []string{}
}

// columnToMapString returns key in map (encoding as JSON) as map[string]interface, if key exists
func columnToMapString(m map[string]interface{}, columnName string) map[string]interface{} {

	if m != nil {
		if columnData, ok := m[columnName]; ok {
			switch columnValue := columnData.(type) {
			case string:
				var mapStringString map[string]interface{}
				if err := json.Unmarshal([]byte(columnValue), &mapStringString); err != nil {
					log.Fatalf("columnToMapString: cannot unmarshal (%s) in column '%s'", columnValue, columnName)
				}
				return mapStringString
			default:
				fatalWrongColumnValueType(columnData, columnName)
			}
		}
	}
	return map[string]interface{}{}
}

func fatalWrongColumnValueType(columnData interface{}, columnName string) {

	log.Fatalf("columnToMapString: received wrong type (%T) in column '%s'", columnData, columnName)
}

// columnToAttributes converts Cassandra column type "map<text, text>" returned
// by gocql as map[string]string into types.Attributes
func columnToAttributes(m map[string]interface{}, columnName string) types.Attributes {

	if m != nil {
		if columnData, ok := m[columnName]; ok {
			switch columnValue := columnData.(type) {
			case map[string]string:
				attributes := types.Attributes{}
				for name, value := range columnValue {
					attributes = append(attributes, *types.NewAttribute(name, value))
				}
				return attributes
			default:
				log.Fatalf("columnToAttributes: received wrong type (%T) in column '%s'", columnData, columnName)
			}
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
