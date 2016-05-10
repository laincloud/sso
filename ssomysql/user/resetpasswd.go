package user

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
)

const createResetPasswordTableSQL = `
CREATE TABLE IF NOT EXISTS reset_password (
	id INT NOT NULL AUTO_INCREMENT,
	user_id INT NOT NULL,
	activation_code VARBINARY(22) NOT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY (activation_code(8))
) DEFAULT CHARSET=latin1`

type ResetPasswordRequest struct {
	Id             int
	UserId         int    `db:"user_id"`
	ActivationCode string `db:"activation_code"`
	Created        string
}

func RequestResetPassword(ctx *models.Context, user *User) error {
	code := generateActivationCode()

	tx := ctx.DB.MustBegin()
	tx.MustExec("INSERT INTO reset_password (user_id, activation_code) VALUES (?, ?)",
		user.Id, code)
	if err := tx.Commit(); err != nil {
		return err
	}

	err := sendResetPasswordEmail(ctx, user, code)
	return err
}

func sendResetPasswordEmail(ctx *models.Context, user *User, code string) error {
	msg := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: SSO reset password\r\n\r\n"+
			"%s, you requested to change your password at %s.  "+
			"If the request is issued by you, click the following link.  "+
			"Ignore this mail otherwise.\n"+
			"    %s/spa/user/password/reset/%s/%s\n"+
			"This link will expire in 24 hours.",
		user.Email.String,
		user.FullName, ctx.SSOSiteURL, ctx.SSOSiteURL,
		url.QueryEscape(user.Name), url.QueryEscape(code)))

	if ctx.SMTPAddr == "" || ctx.SMTPAddr == "mail.example.com:25" {
		log.Warnf("No smtp server configured.  Skip sending email %s", msg)
		return nil
	}

	err := SendMail(ctx.SMTPAddr, nil, ctx.EmailFrom, []string{user.Email.String}, msg)
	return err
}

func ResetPassword(ctx *models.Context, user *User, code string, password string) error {
	req := ResetPasswordRequest{}
	err := ctx.DB.Get(&req,
		"SELECT * FROM reset_password WHERE activation_code=?", code)
	if err == sql.ErrNoRows {
		return ErrCodeNotFound
	} else if err != nil {
		return err
	}

	if user.Id != req.UserId {
		return ErrCodeNotFound
	}

	if err := user.updatePassword([]byte(password)); err != nil {
		return err
	}

	tx := ctx.DB.MustBegin()
	tx.Exec("DELETE FROM reset_password WHERE activation_code=?", code)
	tx.Commit()

	return nil
}
