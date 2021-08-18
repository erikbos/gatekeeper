package service

import (
	"testing"

	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/stretchr/testify/require"
)

func Test_clearSensitiveFields(t *testing.T) {

	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{
			name: "Unmodified Developer",
			value: types.Developer{
				UserName: "joop",
			},
			expected: types.Developer{
				UserName: "joop",
			},
		},
		{
			name: "Modified User",
			value: types.User{
				Name:     "bill",
				Password: "tobedeleted",
			},
			expected: types.User{
				Name:     "bill",
				Password: "",
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected, clearSensitiveFields(test.value), test.name)
	}
}
