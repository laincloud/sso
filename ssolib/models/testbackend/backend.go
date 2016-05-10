package testbackend

import (
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/iuser"
)

type TestBackend struct {
	users []*TestUser
}

func (b *TestBackend) SupportedVerificationMethods() (ret []string) {
	return
}

func (b *TestBackend) Name() (ret string) {
	return
}

func (b *TestBackend) UserIdToSub(id int) (ret string) {
	return
}

func (b *TestBackend) UserSubToId(sub string) (ret int, err error) {
	return
}

func (b *TestBackend) GetUserByFeature(string) (ret iuser.User, err error) {
	return
}

func (b *TestBackend) GetUser(id int) (ret iuser.User, err error) {
	log.Debug(b)
	for _, v := range b.users {
		if v.Id == id {
			return v, nil
		}
	}
	return nil, iuser.ErrUserNotFound
}

func (b *TestBackend) AuthPassword(sub, passwd string) (ret bool, err error) {
	return
}

func (b *TestBackend) ListUsers(ctx context.Context) (ret []iuser.User, err error) {
	return
}

func (b *TestBackend) GetUserByName(name string) (ret iuser.User, err error) {
	for _, v := range b.users {
		if v.Name == name {
			return v, nil
		}
	}
	return nil, iuser.ErrUserNotFound

}

func (b *TestBackend) CreateUser(user iuser.User, passwordHashed bool) (err error) {
	user.SetBackend(b)
	return
}

func (b *TestBackend) DeleteUser(u iuser.User) (err error) {
	return
}

func (b *TestBackend) InitModel(model interface{}) {
	return
}

func (b *TestBackend) Add(u *TestUser) {
	u.Id = len(b.users)
	log.Debug(u)
	b.users = append(b.users, u)
}
