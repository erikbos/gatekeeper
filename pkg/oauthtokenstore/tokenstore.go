package oauthtokenstore

import (
	"log"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

// NewTokenStore bla
func NewTokenStore(a string) (oauth2.TokenStore, error) {

	return &TokenStore{
		s: a,
	}, nil
}

// TokenStore bla
type TokenStore struct {
	s string
}

// Create bla
func (ts *TokenStore) Create(info oauth2.TokenInfo) (err error) {

	log.Printf("Create: %+v", info)
	return nil
}

// RemoveByCode bla
func (ts *TokenStore) RemoveByCode(code string) (err error) {

	log.Printf("RemoveByCode: %+v", code)
	return nil
}

// RemoveByAccess bla
func (ts *TokenStore) RemoveByAccess(access string) error {

	log.Printf("RemoveByAccess: %+v", access)
	return nil
}

// RemoveByRefresh bla
func (ts *TokenStore) RemoveByRefresh(refresh string) error {

	log.Printf("RemoveByRefresh: %+v", refresh)
	return nil
}

// GetByCode bla
func (ts *TokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {

	log.Printf("GetByCode: %+v", code)

	var tm models.Token

	return &tm, nil
}

// GetByAccess bla
func (ts *TokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {

	log.Printf("GetByAccess: %+v", access)

	var tm models.Token

	return &tm, nil
}

// GetByRefresh bla
func (ts *TokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {

	log.Printf("GetByRefresh: %+v", refresh)

	var tm models.Token

	return &tm, nil
}
