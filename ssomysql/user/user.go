package user

import (
	"database/sql"
	"fmt"

	"github.com/mijia/sweb/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/laincloud/sso/ssolib/models/iuser"
)

type User struct {
	Id           int
	Name         string
	FullName     string
	Email        sql.NullString
	PasswordHash []byte `db:"password"`
	Mobile       sql.NullString
	Created      string
	Updated      string

	backend iuser.UserBackend
}

type UserProfile struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
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

func (u *User) GetPasswordHash() []byte {
	return u.PasswordHash
}

func (u *User) VerifyPassword(password []byte) bool {
	valid := compareHashAndPassword(u.PasswordHash, password) == nil
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

func (u *User) GetEmail() string {
	if !u.Email.Valid {
		return ""
	}
	return u.Email.String
}

func (u *User) GetMobile() string {
	if !u.Mobile.Valid {
		return ""
	}
	return u.Mobile.String
}

func (u *User) updatePassword(password []byte) error {
	passwordhash, err := generateHashFromPassword(password, BCRYPT_COST)
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

func (p *UserProfile) GetName() string {
	return p.Name
}

func (p *UserProfile) GetEmail() string {
	return p.Email
}

func (p *UserProfile) GetMobile() string {
	return p.Mobile
}
