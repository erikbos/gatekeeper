package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all users
// (GET /v1/users)
func (h *Handler) GetV1Users(c *gin.Context) {

	users, err := h.service.User.GetAll()
	if err != nil {
		responseError(c, err)
		return
	}
	removePasswords(users)
	h.responseUsers(c, users)
}

// removePasswords set passwords of all users in slice to empty string
func removePasswords(users types.Users) {

	for index := range users {
		users[index].Password = ""
	}
}

// createUser creates an user
// (POST /v1/users)
func (h *Handler) PostV1Users(c *gin.Context) {

	var receivedUser User
	if err := c.ShouldBindJSON(&receivedUser); err != nil {
		responseError(c, types.NewBadRequestError(err))
		return
	}
	newUser := fromUser(receivedUser)
	storedUser, err := h.service.User.Create(newUser, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	// Remove password so we do not show in response
	storedUser.Password = ""
	h.responseUserCreated(c, storedUser)
}

// returns a user
// (GET /v1/users/{user_name})
func (h *Handler) GetV1UsersUserName(c *gin.Context, userName UserName) {

	user, err := h.service.User.Get(string(userName))
	if err != nil {
		responseError(c, err)
		return
	}
	user.Password = ""
	h.responseUser(c, user)
}

// updates an existing user
// (POST /v1/users/{user_name})
func (h *Handler) PostV1UsersUserName(c *gin.Context, userName UserName) {

	_, err := h.service.User.Get(string(userName))
	if err != nil {
		responseError(c, err)
		return
	}
	var receivedUser User
	if err := c.ShouldBindJSON(&receivedUser); err != nil {
		responseError(c, types.NewBadRequestError(err))
		return
	}
	updatedUser := fromUser(receivedUser)
	if updatedUser.Name != string(userName) {
		responseErrorNameValueMisMatch(c)
		return
	}
	storedUser, err := h.service.User.Update(updatedUser, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	// Remove password so we do not show in response
	storedUser.Password = ""
	h.responseUserUpdated(c, storedUser)
}

// deletes an user
// (DELETE /v1/users/{user_name})
func (h *Handler) DeleteV1UsersUserName(c *gin.Context, userName UserName) {

	deletedUser, err := h.service.User.Delete(string(userName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	// Remove password so we do not show in response
	deletedUser.Password = ""
	h.responseUser(c, deletedUser)
}

// API responses

func (h *Handler) responseUsers(c *gin.Context, users types.Users) {

	all_users := make([]User, len(users))
	for i := range users {
		all_users[i] = h.ToUserResponse(&users[i])
	}
	c.IndentedJSON(http.StatusOK, Users{
		User: &all_users,
	})
}

func (h *Handler) responseUser(c *gin.Context, user *types.User) {

	c.IndentedJSON(http.StatusOK, h.ToUserResponse(user))
}

func (h *Handler) responseUserCreated(c *gin.Context, user *types.User) {

	c.IndentedJSON(http.StatusCreated, h.ToUserResponse(user))
}

func (h *Handler) responseUserUpdated(c *gin.Context, user *types.User) {

	c.IndentedJSON(http.StatusOK, h.ToUserResponse(user))
}

// type conversion

func (h *Handler) ToUserResponse(l *types.User) User {

	emptyPassword := ""
	user := User{
		CreatedAt:      &l.CreatedAt,
		CreatedBy:      &l.CreatedBy,
		DisplayName:    &l.DisplayName,
		LastModifiedBy: &l.LastModifiedBy,
		LastModifiedAt: &l.LastModifiedAt,
		Name:           l.Name,
		Password:       &emptyPassword,
		Status:         &l.Status,
	}
	if l.Roles != nil {
		user.Roles = &l.Roles
	} else {
		user.Roles = &[]string{}
	}
	return user
}

func fromUser(u User) types.User {

	user := types.User{}
	if u.CreatedAt != nil {
		user.CreatedAt = *u.CreatedAt
	}
	if u.CreatedBy != nil {
		user.CreatedBy = *u.CreatedBy
	}
	if u.DisplayName != nil {
		user.DisplayName = *u.DisplayName
	}
	if u.LastModifiedBy != nil {
		user.LastModifiedBy = *u.LastModifiedBy
	}
	if u.LastModifiedAt != nil {
		user.LastModifiedAt = *u.LastModifiedAt
	}
	if u.Name != "" {
		user.Name = u.Name
	}
	if u.Password != nil {
		user.Password = *u.Password
	}
	if u.Roles != nil {
		user.Roles = *u.Roles
	} else {
		user.Roles = []string{}
	}
	return user
}
