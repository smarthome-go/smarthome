package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateUserTokenTable(t *testing.T) {
	assert.NoError(t, createUserTokenTable())
}

func TestInsertIntoTokenTable(t *testing.T) {
	assert.NoError(t, InsertUserToken(
		fmt.Sprint(time.Now().UnixMilli()),
		"admin",
		"For Testing",
	))
}

func TestGetUserTokenByToken(t *testing.T) {
	token := fmt.Sprint(time.Now().UnixMilli())
	assert.NoError(t, InsertUserToken(
		token,
		"admin",
		"For Testing",
	))
	data, found, err := GetUserTokenByToken(fmt.Sprint(token))
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, UserToken{
		User:  "admin",
		Token: fmt.Sprint(token),
		Data: UserTokenData{
			Label: "For Testing",
		},
	}, data)
}
