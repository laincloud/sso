package user

import (
	"errors"
	"fmt"

	"github.com/mijia/sweb/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

const (
	// Use ./cmd/bcryptcost tool to find approriate cost
	BCRYPT_COST = 11
)

type UserInfo struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Mobile   string `json:"mobile"`
}

func (ur *UserInfo) Validate(ctx context.Context) error {
	if ur.Name == "" {
		return errors.New("Empty name")
	}
	if err := ssolib.ValidateSlug(ur.Name, 32); err != nil {
		return err
	}

	if ur.FullName != "" {
		if err := ssolib.ValidateFullName(ur.FullName); err != nil {
			return err
		}
	}

	if ur.Email == "" {
		return errors.New("Empty email")
	}
	if err := ssolib.ValidateUserEmail(ur.Email, ctx); err != nil {
		return err
	}

	if ur.Password == "" {
		return errors.New("Empty password")
	}
	if len(ur.Password) < 4 {
		return errors.New("Password too short")
	}

	return nil
}

type User struct {
	Id           int
	Name         string
	FullName     string
	Email        string
	PasswordHash []byte `db:"password"`
	Mobile       string
	Created      string
	Updated      string

	dn      string
	backend iuser.UserBackend
}

type UserProfile struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

func (up *UserProfile) GetName() string {
	return up.Name
}

func (up *UserProfile) GetEmail() string {
	return up.Email
}

func (up *UserProfile) GetMobile() string {
	return up.Mobile
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) GetFullName() string {
	return u.FullName
}

func (u *User) GetId() int {
	return u.Id
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetMobile() string {
	log.Debug(u.Mobile)
	return u.Mobile
}

func (u *User) SetBackend(b iuser.UserBackend) {
	u.backend = b
}

func (u *User) GetSub() string {
	return u.backend.(*UserBack).UserIdToSub(u.Id)
}

func (u *User) GetProfile() iuser.UserProfile {
	return &UserProfile{
		Name:     u.GetName(),
		FullName: u.GetFullName(),
		Email:    u.GetEmail(),
		Mobile:   u.GetMobile(),
	}
}

func (u *User) GetPublicProfile() iuser.UserProfile {
	return &UserProfile{
		Name:     u.GetName(),
		FullName: u.GetFullName(),
		Email:    u.GetEmail(),
	}
}

func (u *User) VerifyPassword(password []byte) bool {
	valid := bcrypt.CompareHashAndPassword(u.PasswordHash, password) == nil
	if !valid {
		return false
	}

	// Check bcrypt hash cost.  If it is not equals to BCRYPT_COST, rehash it.
	cost, err := bcrypt.Cost(u.PasswordHash)
	if err != nil {
		log.Warnf("Weird!. Get bcrypt cost failed: %v", err)
		return valid
	}

	if cost != BCRYPT_COST {
		u.updatePassword(password)
	}

	return valid
}

func (u *User) updatePassword(password []byte) error {
	passwordhash, err := bcrypt.GenerateFromPassword(password, BCRYPT_COST)
	if err != nil {
		return fmt.Errorf("Generate password hash failed: %s", err)
	}

	tx := u.backend.(*UserBack).DB.MustBegin()
	_, err1 := tx.Exec(
		"UPDATE user SET password=? WHERE id=?",
		passwordhash, u.Id)

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}

	if err1 != nil {
		return err1
	}

	return nil
}
