package types

import "sort"

// User holds an user
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type (
	User struct {
		// Name of user (not changable)
		Name string `binding:"required,min=4"`

		// Display name
		DisplayName string

		// Password
		Password string

		// Status of this user
		Status string

		// Role of this user
		Roles []string

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this user
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this user
		LastModifiedBy string
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
