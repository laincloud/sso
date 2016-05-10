package testbackend

import (
	"github.com/laincloud/sso/ssolib/models/iuser"
)

type TestUser struct {
	Id           int
	Name         string
	PasswordHash []byte
}

func (u *TestUser) GetMobile() (ret string) {
	return
}

func (u *TestUser) GetSub() (ret string) {
	return
}

func (u *TestUser) GetId() (ret int) {
	return u.Id
}

func (u *TestUser) GetProfile() (ret iuser.UserProfile) {
	return
}

func (u *TestUser) SetBackend(back iuser.UserBackend) {
	back.(*TestBackend).Add(u)
	return
}

func (u *TestUser) GetName() (ret string) {
	return u.Name
}
