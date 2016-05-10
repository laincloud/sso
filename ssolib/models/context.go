package models

import (
	"net/url"

	"github.com/jmoiron/sqlx"

	"github.com/laincloud/sso/ssolib/lock"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

type Context struct {
	DB         *sqlx.DB
	SSOSiteURL *url.URL
	SMTPAddr   string
	EmailFrom  string

	// 支持多后端要改为数组？
	Back iuser.UserBackend

	Lock *lock.DistributedLock
}
