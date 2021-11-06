package types

import "sort"

// User holds an user
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type (
	User struct {
		// Name of user (not changable)
		Name string `json:"name" binding:"required,min=4"`

		// Display name
		DisplayName string `json:"displayName"`

		// Password
		Password string `json:"password,omitempty"`

		// Status of this user
		Status string `json:"status"`

		// Role of this user
		Roles StringSlice `json:"roles"`

		// Created at timestamp in epoch milliseconds
		CreatedAt int64 `json:"createdAt"`

		// Name of user who created this user
		CreatedBy string `json:"createdBy"`

		// Last modified at timestamp in epoch milliseconds
		LastmodifiedAt int64 `json:"lastmodifiedAt"`

		// Name of user who last updated this user
		LastmodifiedBy string `json:"lastmodifiedBy"`
	}

	// Users holds one or more users
	Users []User
)

var (
	// NullUser is an empty user type
	NullUser = User{}

	// NullUsers is an empty user slice
	NullUsers = Users{}
)

// Sort a slice of users
func (users Users) Sort() {
	// Sort users by name
	sort.SliceStable(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})
}
