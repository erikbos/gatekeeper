package types

import (
	"sort"

	"github.com/go-playground/validator/v10"
)

// User holds an user
type (
	User struct {
		// Name of user (not changable)
		Name string `validate:"required,min=1,max=100"`

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

// Validate checks if field values are set correct and are allowed
func (u *User) Validate() error {

	validate := validator.New()
	return validate.Struct(u)
}

// Sort a slice of users
func (users Users) Sort() {
	// Sort users by name
	sort.SliceStable(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})
}
