package user

import (
	"database/sql"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	reg1 = UserRegistration{
		Name:     "testuser",
		Email:    sql.NullString{String: "testuser@example.com", Valid: true},
		Password: "test",
	}
)

func TestRegisterUserShouldSendActivationCodeWhenSuccess(t *testing.T) {
	th := NewTestHelper(t)

	sendMailCalled := 0
	SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sendMailCalled += 1
		assert.Equal(t, to, []string{"testuser@example.com"})
		return nil
	}

	_, err := RegisterUser(th.Ctx, reg1, testBack)
	assert.Nil(t, err)

	assert.Equal(t, 1, sendMailCalled)
}

func TestRegisterUserShouldFailWhenNameAlreadyToken(t *testing.T) {
	th := NewTestHelper(t)

	err := testBack.CreateUser(&User{Name: "testuser", PasswordHash: []byte("foo")}, true)
	assert.Nil(t, err)

	_, err = RegisterUser(th.Ctx, reg1, testBack)
	assert.Equal(t, ErrUserExists, err)
}

func TestActivateUserShouldInsertIntoUserTableWhenCodeMatch(t *testing.T) {
	th := NewTestHelper(t)

	code, err := RegisterUser(th.Ctx, reg1, testBack)
	assert.Nil(t, err)
	user, err := ActivateUser(th.Ctx, code, testBack)
	assert.Nil(t, err)
	assert.Equal(t, "testuser", user.Name)
}
