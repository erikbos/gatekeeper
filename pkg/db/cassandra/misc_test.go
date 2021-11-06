package cassandra

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func Test_Attribute_Unmarshall(t *testing.T) {

	tests := []struct {
		name           string
		attributesJson string
		attributes     types.Attributes
	}{
		{
			name:           "1 - attributes in json",
			attributesJson: "[{\"name\":\"SNIHostName\",\"value\":\"www.example.com\"},{\"name\":\"HTTPProtocol\",\"value\":\"HTTP/2\"}]",
			attributes: types.Attributes{
				types.Attribute{
					Name:  types.AttributeSNIHostName,
					Value: "www.example.com",
				},
				types.Attribute{
					Name:  types.AttributeHTTPProtocol,
					Value: types.AttributeValueHTTPProtocol2,
				},
			},
		},
		{
			name:           "1 - attributes in json",
			attributesJson: "",
			attributes:     types.NullAttributes,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.attributes, AttributesUnmarshal(test.attributesJson), test.name)
	}
}

func Test_Attribute_Marshall(t *testing.T) {

	tests := []struct {
		name           string
		attributesJson string
		attributes     types.Attributes
	}{
		{
			name: "1 - two attributes",
			attributes: types.Attributes{
				types.Attribute{
					Name:  types.AttributeSNIHostName,
					Value: "www.test.com",
				},
				types.Attribute{
					Name:  types.AttributeHTTPProtocol,
					Value: types.AttributeValueHTTPProtocol2,
				},
			},
			attributesJson: "[{\"name\":\"SNIHostName\",\"value\":\"www.test.com\"},{\"name\":\"HTTPProtocol\",\"value\":\"HTTP/2\"}]",
		},
		{
			name:           "2 - no attributes",
			attributes:     types.NullAttributes,
			attributesJson: "[]",
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.attributesJson, AttributesMarshal(test.attributes), test.name)
	}
}
