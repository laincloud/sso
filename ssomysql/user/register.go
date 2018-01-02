package user

import (
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"strings"

	"github.com/mijia/sweb/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

const (
	// Use ./cmd/bcryptcost tool to find approriate cost
	BCRYPT_COST = 11
)

const createUserRegisterTableSQL = `
CREATE TABLE IF NOT EXISTS user_register (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(32) NOT NULL,
	fullname VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	email VARCHAR(64) NULL DEFAULT NULL,
	password VARBINARY(60) NOT NULL,
	mobile VARCHAR(11) NULL DEFAULT NULL,
	activation_code VARBINARY(22) NOT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY (activation_code(8))
) DEFAULT CHARSET=latin1`

var (
	compareHashAndPassword   = bcrypt.CompareHashAndPassword
	generateHashFromPassword = bcrypt.GenerateFromPassword

	ErrUserExists   = errors.New("User already exists")
	ErrCodeNotFound = errors.New("Code not found")
)

func sendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

	host, _, _ := net.SplitHostPort(addr)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	//create smtp client
	conn, err := tls.Dial("tcp", addr, tlsconfig)

	if err != nil {
		log.Warnf("tls.Dail: %s", err.Error())
		return err
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Warnf("create SMTP client: %s", err.Error())
		return err
	}

	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Warnf("Error during AUTH: %s", err.Error())
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

func GenericSendMail(ctx *models.Context, to []string, msg []byte) (err error) {
	var auth smtp.Auth
	host, _, _ := net.SplitHostPort(ctx.SMTPAddr)

	if ctx.EmailFromPassword != "" {
		auth = smtp.PlainAuth("", ctx.EmailFrom, ctx.EmailFromPassword, host)
	}
	if ctx.EmailTLS == false {
		err = smtp.SendMail(ctx.SMTPAddr, auth, ctx.EmailFrom, to, msg)
	} else {
		err = sendMailUsingTLS(ctx.SMTPAddr, auth, ctx.EmailFrom, to, msg)
	}
	return err
}

type UserRegistration struct {
	Id             int
	Name           string
	FullName       string
	Email          sql.NullString
	Password       string
	Mobile         sql.NullString
	ActivationCode string `db:"activation_code"`
	Created        string
}

func RegisterUser(ctx *models.Context, reg UserRegistration, userback iuser.UserBackend) (code string, err error) {
	if err = validateUserRegistration(ctx, reg, userback); err != nil {
		return "", err
	}

	fullname := reg.FullName
	if fullname == "" {
		fullname = reg.Name
	}

	passwordHash, err := generateHashFromPassword([]byte(reg.Password), BCRYPT_COST)
	if err != nil {
		err = fmt.Errorf("Generate password hash failed: %s", err)
		return
	}

	code = generateActivationCode()
	log.Debugf("Generated activation code: %s", code)

	tx := ctx.DB.MustBegin()
	tx.MustExec("INSERT INTO user_register (name, fullname, email, password, mobile, activation_code) VALUES (?, ?, ?, ?, ?, ?)",
		reg.Name, fullname, reg.Email, passwordHash, reg.Mobile, code)
	if err = tx.Commit(); err != nil {
		return
	}

	err = sendActivationEmail(ctx, reg.Name, reg.Email.String, code)
	return
}

func ActivateUser(ctx *models.Context, code string, userback iuser.UserBackend) (*User, error) {
	reg := UserRegistration{}
	err := ctx.DB.Get(&reg,
		"SELECT * FROM user_register WHERE activation_code=?", code)
	if err == sql.ErrNoRows {
		return nil, ErrCodeNotFound
	} else if err != nil {
		return nil, err
	}

	user := &User{
		Name:         reg.Name,
		FullName:     reg.FullName,
		Email:        reg.Email,
		PasswordHash: []byte(reg.Password),
		Mobile:       reg.Mobile,
	}
	err = userback.CreateUser(user, true)
	if err != nil {
		return nil, err
	}

	tx := ctx.DB.MustBegin()
	tx.Exec("DELETE FROM user_register WHERE activation_code=?", code)
	tx.Commit()
	userIntf, err := userback.GetUserByName(reg.Name)
	user = &User{
		Name:         userIntf.GetName(),
		FullName:     userIntf.(*User).GetFullName(),
		Email:        string2sqlString(userIntf.(*User).GetEmail()),
		PasswordHash: userIntf.(*User).GetPasswordHash(),
		Mobile:       string2sqlString(userIntf.GetMobile()),
	}
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.New("User unexpected disappeard")
	}

	return user, nil
}

func validateUserRegistration(ctx *models.Context, reg UserRegistration, userback iuser.UserBackend) error {
	_, err := userback.GetUserByName(reg.Name)
	if err != iuser.ErrUserNotFound {
		if err == nil {
			return ErrUserExists
		} else {
			return err
		}
	}

	_, err = userback.(*UserBack).GetUserByEmail(reg.Email.String)
	if err != iuser.ErrUserNotFound {
		if err == nil {
			return ErrUserExists
		} else {
			return err
		}
	}

	return nil
}

func generateActivationCode() string {
	code := make([]byte, 16)
	_, err := rand.Read(code)
	if err != nil {
		panic(err)
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(code), "=")
}

func sendActivationEmail(ctx *models.Context, username, email string, code string) error {
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: SSO registration\r\n\r\n"+
		"You are registering user %s at %s.  Please click on "+
		"the following link to activate your account:\n"+
		"    %s/api/activateuser?code=%s",
		email,
		username, ctx.SSOSiteURL, ctx.SSOSiteURL,
		url.QueryEscape(code)))

	if ctx.SMTPAddr == "" || ctx.SMTPAddr == "mail.example.com:25" {
		log.Warnf("No smtp server configured.  Skip sending email %s", msg)
		return nil
	}
	err := GenericSendMail(ctx, []string{email}, msg)
	return err
}

func string2sqlString(s string) sql.NullString {
	var b bool
	if s == "" {
		b = false
	} else {
		b = true
	}
	return sql.NullString{
		String: s,
		Valid:  b,
	}
}
