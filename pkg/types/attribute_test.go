package types

import (
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
			name:          "1",
			attributes:    testAttributes,
			attributeName: AttributeTimeout,
			defaultValue:  0,
			expectedValue: 42 * time.Millisecond,
		},
		{
			name:          "2",
			attributes:    testAttributes,
			attributeName: AttributeTimeout,
			defaultValue:  69 * time.Second,
			expectedValue: 42 * time.Millisecond,
		},
		{
			name:          "3 - bad time value",
			attributes:    testAttributes,
			attributeName: AttributeWithBadValue,
			defaultValue:  12 * time.Hour,
			expectedValue: 12 * time.Hour,
		},
		{
			name:          "3",
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
