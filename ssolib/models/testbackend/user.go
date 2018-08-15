package testbackend

import (
	"github.com/laincloud/sso/ssolib/models/iuser"
)

type TestUser struct {
	Id           int
	Name         string
	PasswordHash []byte
	Email        string
}

type UserProfile struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
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

func (u *TestUser) GetProfile() iuser.UserProfile {
	return &UserProfile{
		Name: u.Name,
		FullName: u.Name,
		Email: u.Email,
	}
}

func (u *TestUser) GetPublicProfile() (ret iuser.UserProfile) {
	return
}

func (u *TestUser) SetBackend(back iuser.UserBackend) {
	back.(*TestBackend).Add(u)
	return
}

func (u *TestUser) GetName() (ret string) {
	return u.Name
}

func (p *UserProfile) GetName() string {
	return p.Name
}

func (p *UserProfile) GetEmail() string {
	return p.Email
}

func (p *UserProfile) GetMobile() string {
	return p.Mobile
}
