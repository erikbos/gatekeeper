package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// UserService is
type UserService struct {
	db        *db.Database
	changelog *Changelog
}

// NewUserService returns a new user instance
func NewUserService(database *db.Database, c *Changelog) *UserService {

	return &UserService{db: database, changelog: c}
}

// GetAll returns all users
func (us *UserService) GetAll() (users types.Users, err types.Error) {

	return us.db.User.GetAll()
}

// Get returns details of an user
func (us *UserService) Get(userName string) (user *types.User, err types.Error) {

	return us.db.User.Get(userName)
}

// Create creates an user
func (us *UserService) Create(newUser types.User, who Requester) (*types.User, types.Error) {

	existingUser, err := us.db.User.Get(newUser.Name)
	if err == nil {
		return nil, types.NewBadRequestError(
			fmt.Errorf("User '%s' already exists", existingUser.Name))
	}
	// Automatically set default fields
	newUser.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newUser.CreatedBy = who.Username

	// encrypt password
	encryptedPassword, encryptError := cryptPassword(newUser.Password)
	if encryptError != nil {
		return nil, types.NewUpdateFailureError(encryptError)
	}
	newUser.Password = encryptedPassword

	if err = us.updateUser(&newUser, who); err != nil {
		return nil, err
	}
	us.changelog.Create(newUser, who)
	return &newUser, nil
}

// Update updates an existing user
func (us *UserService) Update(updatedUser types.User,
	who Requester) (*types.User, types.Error) {

	currentUser, err := us.db.User.Get(updatedUser.Name)
	if err != nil {
		return nil, types.NewItemNotFoundError(err)
	}
	// Populate fields which are not updateable
	updatedUser.Name = currentUser.Name
	updatedUser.CreatedAt = currentUser.CreatedAt
	updatedUser.CreatedBy = currentUser.CreatedBy

	// Update password only when a (new) password has been provided
	if updatedUser.Password != "" {
		encryptedPasswd, encryptError := cryptPassword(updatedUser.Password)
		if encryptError != nil {
			return nil, types.NewUpdateFailureError(encryptError)
		}
		updatedUser.Password = encryptedPasswd
	}

	if err = us.updateUser(&updatedUser, who); err != nil {
		return nil, err
	}
	us.changelog.Update(currentUser, updatedUser, who)
	return &updatedUser, nil
}

// updateUser updates last-modified field(s) and updates user in database
func (us *UserService) updateUser(updatedUser *types.User, who Requester) types.Error {

	updatedUser.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedUser.LastmodifiedBy = who.Username
	return us.db.User.Update(updatedUser)
}

// Delete deletes an user
func (us *UserService) Delete(userName string, who Requester) (deletedUser *types.User, e types.Error) {

	user, err := us.db.User.Get(userName)
	if err != nil {
		return nil, err
	}
	if err = us.db.User.Delete(userName); err != nil {
		return nil, err
	}
	us.changelog.Delete(user, who)
	return user, nil
}

// cryptPassword returns a bcrypt()ed password
func cryptPassword(password string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPasswordHash validates password hash and returns true in case of a match
func CheckPasswordHash(password, hash string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
