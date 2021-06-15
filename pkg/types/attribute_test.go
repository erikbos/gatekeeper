package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	NonExistingAttribute string = "NonExistingAttributeKey"

	AttributeWithBadValue string = "AttributeWithNonParseableKey"

	testAttributes Attributes = Attributes{
		{
			Name:  AttributeSNIHostName,
			Value: "www.example.com",
		},
		{
			Name:  AttributeTimeout,
			Value: "42ms",
		},
		{
			Name:  AttributeExtAuthzRequestBodySize,
			Value: "4096",
		},
		{
			Name:  AttributeWithBadValue,
			Value: "-AA",
		},
		{
			Name:  AttributeTLSMinimumVersion,
			Value: AttributeValueTLSVersion12,
		},
		{
			Name:  "Country",
			Value: "CA",
		},
	}
)

func Test_Attribute_Get(t *testing.T) {

	tests := []struct {
		name          string
		attributes    Attributes
		attributeName string
		expectedValue string
		expectedError Error
	}{
		{
			name:          "1",
			attributes:    testAttributes,
			attributeName: AttributeTLSMinimumVersion,
			expectedValue: AttributeValueTLSVersion12,
			expectedError: nil,
		},
		{
			name:          "2",
			attributes:    testAttributes,
			attributeName: AttributeSNIHostName,
			expectedValue: "www.example.com",
			expectedError: nil,
		},
		{
			name:          "3",
			attributes:    testAttributes,
			attributeName: NonExistingAttribute,
			expectedValue: "",
			expectedError: errAttributeNotFound,
		},
	}
	for _, test := range tests {
		value, err := test.attributes.Get(test.attributeName)
		require.Equalf(t, test.expectedValue, value, test.name)
		require.Equalf(t, test.expectedError, err, test.name)
	}
}

func Test_Attribute_GetAsString(t *testing.T) {

	tests := []struct {
		name          string
		attributes    Attributes
		attributeName string
		defaultValue  string
		expectedValue string
	}{
		{
			name:          "1",
			attributes:    testAttributes,
			attributeName: AttributeSNIHostName,
			defaultValue:  "",
			expectedValue: "www.example.com",
		},
		{
			name:          "2",
			attributes:    testAttributes,
			attributeName: AttributeSNIHostName,
			defaultValue:  "123",
			expectedValue: "www.example.com",
		},
		{
			name:          "4",
			attributes:    testAttributes,
			attributeName: NonExistingAttribute,
			defaultValue:  "www.test.com",
			expectedValue: "www.test.com",
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expectedValue,
			test.attributes.GetAsString(test.attributeName, test.defaultValue), test.name)
	}
}

func Test_Attribute_GetAsUInt32(t *testing.T) {

	tests := []struct {
		name          string
		attributes    Attributes
		attributeName string
		defaultValue  uint32
		expectedValue uint32
	}{
		{
			name:          "1",
			attributes:    testAttributes,
			attributeName: AttributeExtAuthzRequestBodySize,
			defaultValue:  0,
			expectedValue: 4096,
		},
		{
			name:          "2",
			attributes:    testAttributes,
			attributeName: AttributeExtAuthzRequestBodySize,
			defaultValue:  69,
			expectedValue: 4096,
		},
		{
			name:          "3 - bad uint32 value",
			attributes:    testAttributes,
			attributeName: AttributeWithBadValue,
			defaultValue:  9999,
			expectedValue: 9999,
		},
		{
			name:          "3",
			attributes:    testAttributes,
			attributeName: NonExistingAttribute,
			defaultValue:  123,
			expectedValue: 123,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expectedValue,
			test.attributes.GetAsUInt32(test.attributeName, test.defaultValue), test.name)
	}
}

func Test_Attribute_GetAsDuration(t *testing.T) {

	tests := []struct {
		name          string
		attributes    Attributes
		attributeName string
		defaultValue  time.Duration
		expectedValue time.Duration
	}{
		{
			name:          "1 - get an attribute as duration",
			attributes:    testAttributes,
			attributeName: AttributeTimeout,
			defaultValue:  0,
			expectedValue: 42 * time.Millisecond,
		},
		{
			name:          "2 - get an attribute as duration",
			attributes:    testAttributes,
			attributeName: AttributeTimeout,
			defaultValue:  69 * time.Second,
			expectedValue: 42 * time.Millisecond,
		},
		{
			name:          "3 - non parseable duration",
			attributes:    testAttributes,
			attributeName: AttributeWithBadValue,
			defaultValue:  12 * time.Hour,
			expectedValue: 12 * time.Hour,
		},
		{
			name:          "3 - non existing attribute",
			attributes:    testAttributes,
			attributeName: NonExistingAttribute,
			defaultValue:  123 * time.Second,
			expectedValue: 123 * time.Second,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expectedValue,
			test.attributes.GetAsDuration(test.attributeName, test.defaultValue), test.name)
	}
}

func Test_Attribute_Set(t *testing.T) {

	tests := []struct {
		name            string
		attributes      Attributes
		attributeToSet  Attribute
		attributesAfter Attributes
		err             Error
	}{
		{
			name:       "1 - add one attribute to empty set",
			attributes: NullAttributes,
			attributeToSet: Attribute{
				Name:  AttributeSNIHostName,
				Value: "www.test.com",
			},
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
			},
			err: nil,
		},
		{
			name: "2 - add second attribute to set with one attribute",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
			},
			attributeToSet: Attribute{
				Name:  AttributeHTTPProtocol,
				Value: AttributeValueHTTPProtocol2,
			},
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			err: nil,
		},
		{
			name: "3 - Overwrite existing attribute in set with one atttribute",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			attributeToSet: Attribute{
				Name:  AttributeSNIHostName,
				Value: "www.example.com",
			},
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.example.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			err: nil,
		},
	}
	for _, test := range tests {
		err := test.attributes.Set(test.attributeToSet)
		require.Equalf(t, test.err, err, test.name)
		require.Equalf(t, test.attributesAfter, test.attributes, test.name)
	}

	// Try filling up attribute bag with more than MaximumNumberofAttributesAllowed entries
	testName := "MaximumNumberofAttributesAllowed"
	a := Attributes{}
	for i := 0; i <= MaximumNumberofAttributesAllowed; i++ {
		require.Equalf(t, nil, a.Set(Attribute{
			Name:  fmt.Sprintf("attribute%d", i),
			Value: fmt.Sprintf("value%d", i),
		}), testName)
	}
	require.Equalf(t, errTooManyAttributes, a.Set(Attribute{
		Name:  "ShouldNotFitAnymore",
		Value: "WeAreFull",
	}), testName)
}

func Test_Attribute_SetMultiple(t *testing.T) {

	tests := []struct {
		name            string
		attributes      Attributes
		attributeToSet  Attributes
		attributesAfter Attributes
		err             Error
	}{
		{
			name:       "1 - two attributes to add to empty set",
			attributes: NullAttributes,
			attributeToSet: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			err: nil,
		},
		{
			name: "2 - two attributes to add to set with one attribute",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
			},
			attributeToSet: Attributes{
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
				Attribute{
					Name:  AttributeTLSMinimumVersion,
					Value: AttributeValueTLSVersion13,
				},
			},
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
				Attribute{
					Name:  AttributeTLSMinimumVersion,
					Value: AttributeValueTLSVersion13,
				},
			},
			err: nil,
		},
		{
			name: "3 - two attributes to set, add one & overwrite one",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			attributeToSet: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.example.com",
				},
			},
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.example.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			err: nil,
		},
	}
	for _, test := range tests {
		err := test.attributes.SetMultiple(test.attributeToSet)
		require.Equalf(t, test.err, err, test.name)
		require.Equalf(t, test.attributesAfter, test.attributes, test.name)
	}

	// Try filling up attribute bag with more than MaximumNumberofAttributesAllowed entries
	testName := "MaximumNumberofAttributesAllowed"
	a := Attributes{}
	for i := 0; i <= MaximumNumberofAttributesAllowed; i++ {
		require.Equalf(t, nil, a.Set(Attribute{
			Name:  fmt.Sprintf("attribute%d", i),
			Value: fmt.Sprintf("value%d", i),
		}), testName)
	}
	require.Equalf(t, errTooManyAttributes, a.SetMultiple(
		Attributes{
			Attribute{
				Name:  "ShouldNotFitAnymore",
				Value: "WeAreFull",
			},
		}), testName)

}

func Test_Attribute_Tidy(t *testing.T) {

	tests := []struct {
		name            string
		attributes      Attributes
		attributesAfter Attributes
		err             Error
	}{
		{
			name: "1 - naming and values have whitespace, one duplicate entry",
			attributes: Attributes{
				Attribute{
					Name:  "ZZZLastEntry      ",
					Value: "   $1000 ",
				},
				Attribute{
					Name:  " " + AttributeSNIHostName + "   ",
					Value: "www.test.com ",
				},
				Attribute{
					Name:  " " + AttributeSNIHostName + "   ",
					Value: "www.test.com ",
				},
				Attribute{
					Name:  "AAAFirstEntry",
					Value: "123",
				},
			},
			attributesAfter: Attributes{
				Attribute{
					Name:  "AAAFirstEntry",
					Value: "123",
				},
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  "ZZZLastEntry",
					Value: "$1000",
				},
			},
			err: nil,
		},
	}
	for _, test := range tests {
		test.attributes.Tidy()
		require.Equalf(t, test.attributesAfter, test.attributes, test.name)
	}
}

func Test_Attribute_Delete(t *testing.T) {

	tests := []struct {
		name                    string
		attributes              Attributes
		attributeToDelete       string
		attributesAfter         Attributes
		valueOfDeletedAttribute string
		err                     Error
	}{
		{
			name: "1 - delete one existing attribute, one remaining",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			attributeToDelete: AttributeHTTPProtocol,
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
			},
			valueOfDeletedAttribute: AttributeValueHTTPProtocol2,
			err:                     nil,
		},
		{
			name: "2 - delete one existing attribute, none remaining",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
			},
			attributeToDelete:       AttributeSNIHostName,
			attributesAfter:         NullAttributes,
			valueOfDeletedAttribute: "www.test.com",
			err:                     nil,
		},
		{
			name: "3 - delete one non-existing attribute, one remaining",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			attributeToDelete: AttributeTLSMinimumVersion,
			attributesAfter: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			valueOfDeletedAttribute: "",
			err:                     errAttributeNotFound,
		},
	}
	for _, test := range tests {
		valueOfDeletedAttribute, err := test.attributes.Delete(test.attributeToDelete)
		require.Equalf(t, test.err, err, test.name)
		require.Equalf(t, test.valueOfDeletedAttribute, valueOfDeletedAttribute, test.name)
		require.Equalf(t, test.attributesAfter, test.attributes, test.name)
	}
}

func Test_Attribute_Unmarshall(t *testing.T) {

	tests := []struct {
		name           string
		attributesJson string
		attributes     Attributes
	}{
		{
			name:           "1 - attributes in json",
			attributesJson: "[{\"name\":\"SNIHostName\",\"value\":\"www.example.com\"},{\"name\":\"HTTPProtocol\",\"value\":\"HTTP/2\"}]",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.example.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
		},
		{
			name:           "1 - attributes in json",
			attributesJson: "",
			attributes:     NullAttributes,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.attributes, test.attributes.Unmarshal(test.attributesJson), test.name)
	}
}

func Test_Attribute_Marshall(t *testing.T) {

	tests := []struct {
		name           string
		attributesJson string
		attributes     Attributes
	}{
		{
			name: "1 - two attributes",
			attributes: Attributes{
				Attribute{
					Name:  AttributeSNIHostName,
					Value: "www.test.com",
				},
				Attribute{
					Name:  AttributeHTTPProtocol,
					Value: AttributeValueHTTPProtocol2,
				},
			},
			attributesJson: "[{\"name\":\"SNIHostName\",\"value\":\"www.test.com\"},{\"name\":\"HTTPProtocol\",\"value\":\"HTTP/2\"}]",
		},
		{
			name:           "2 - no attributes",
			attributes:     NullAttributes,
			attributesJson: "[]",
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.attributesJson, test.attributes.Marshal(), test.name)
	}
}
