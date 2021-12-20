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
	h.responseUsers(c, users)
}

// createUser creates an user
// (POST /v1/users)
func (h *Handler) PostV1Users(c *gin.Context) {

	var receivedUser User
	if err := c.ShouldBindJSON(&receivedUser); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newUser := fromUser(receivedUser)
	storedUser, err := h.service.User.Create(newUser, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
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
		responseErrorBadRequest(c, err)
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
	h.responseUserUpdated(c, storedUser)
}

// deletes an user
// (DELETE /v1/users/{user_name})
func (h *Handler) DeleteV1UsersUserName(c *gin.Context, userName UserName) {

	deletedUser, err := h.service.User.Get(string(userName))
	if err != nil {
		responseError(c, err)
		return
	}
	if err := h.service.User.Delete(string(userName), h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseUser(c, deletedUser)
}

// API responses

func (h *Handler) responseUsers(c *gin.Context, users types.Users) {

	allUsers := make([]User, len(users))
	for i := range users {
		allUsers[i] = h.ToUserResponse(&users[i])
	}
	c.IndentedJSON(http.StatusOK, Users{
		User: &allUsers,
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

	user := User{
		CreatedAt:      &l.CreatedAt,
		CreatedBy:      &l.CreatedBy,
		DisplayName:    &l.DisplayName,
		LastModifiedBy: &l.LastModifiedBy,
		LastModifiedAt: &l.LastModifiedAt,
		Name:           l.Name,
		// We never return password
		Password: nil,
		Status:   &l.Status,
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
	if u.Status != nil {
		user.Status = *u.Status
	}
	return user
}
