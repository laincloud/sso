package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

var (
	ErrAppNotFound = errors.New("App not found")
)

// if the role_id is the group_id, the role tree is belong to this app,
// otherwise the app only uses the role tree which is created by another
// app's admins.
var createAppTableSQL = `
CREATE TABLE IF NOT EXISTS app (
	id INT NOT NULL AUTO_INCREMENT,
	fullname VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	secret VARBINARY(22) NOT NULL,
	redirect_uri VARCHAR(256) NOT NULL,
	admin_group_id INT NULL DEFAULT NULL,
	admin_role_id INT NULL DEFAULT -1,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) DEFAULT CHARSET=latin1`

func InitDatabase(ctx *models.Context) {
	ctx.DB.MustExec(createAppTableSQL)
}

type App struct {
	Id           int
	FullName     string
	Secret       string
	RedirectUri  string `db:"redirect_uri"`
	AdminGroupId int    `db:"admin_group_id"`
	AdminRoleId  int    `db:"admin_role_id"`
	Created      string
	Updated      string
}

func (a *App) SecretString() string {
	return a.Secret
}

// implemenent osin.Client

func (a *App) GetId() string {
	return strconv.Itoa(int(a.Id))
}

func (a *App) GetSecret() string {
	return a.SecretString()
}

func (a *App) GetRedirectUri() string {
	return a.RedirectUri
}

func (a *App) GetUserData() interface{} {
	return nil
}

func UpdateApp(ctx *models.Context, app *App) (*App, error) {
	_, err := ctx.DB.Exec("UPDATE app SET fullname=?, redirect_uri=? WHERE id=?", app.FullName, app.RedirectUri, app.Id)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	return GetApp(ctx, app.Id)
}

func CreateApp(ctx *models.Context, app *App, owner iuser.User) (*App, error) {
	secret := app.Secret
	if len(secret) == 0 {
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		secret = strings.TrimRight(base64.URLEncoding.EncodeToString(buf), "=")
	}
	tx := ctx.DB.MustBegin()
	result, err := tx.Exec("INSERT INTO app (fullname, secret, redirect_uri) "+
		"VALUES (?, ?, ?)", app.FullName, secret, app.RedirectUri)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	groupName := fmt.Sprintf(".app-%d", id)
	groupFullName := fmt.Sprintf("App %d Admin Group", id)
	g, err := group.CreateGroup(ctx, &group.Group{Name: groupName, FullName: groupFullName})
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if _, err = tx.Exec("UPDATE app SET admin_group_id=?, admin_role_id=? WHERE id=?",
		g.Id, g.Id, id); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err = g.AddMember(ctx, owner, group.ADMIN); err != nil {
		tx.Rollback()
		return nil, err
	}

	// todo add first role for this app zhiwang-test-sys
	if _, err = tx.Exec("INSERT INTO role (id, name, fullname, app_id) VALUES (?, ?, ?, ?)",
		g.Id, app.FullName+"-sys", app.FullName+"-sys", id); err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	app, err = GetApp(ctx, int(id))
	if err != nil {
		return nil, err
	}
	return app, nil
}

func ListAppsByAdminGroupIds(ctx *models.Context, groupIds []int) ([]App, error) {
	query, args, err := sqlx.In("SELECT * FROM app WHERE admin_group_id IN(?)", groupIds)
	if err != nil {
		return nil, err
	}
	apps := []App{}
	err = ctx.DB.Select(&apps, query, args...)
	return apps, err
}

func GetApp(ctx *models.Context, id int) (*App, error) {
	log.Debugf("GetAPP: %d", id)
	app := App{}
	err := ctx.DB.Get(&app, "SELECT * FROM app WHERE id=?", id)
	log.Debugf("GetAPP: %d finish", id)
	if err == sql.ErrNoRows {
		return nil, ErrAppNotFound
	} else if err != nil {
		return nil, err
	}
	return &app, nil
}

func AppNameExist(ctx *models.Context, appName string) (bool, error) {
	log.Debug("AppNameExist: %s", appName)
	app := App{}
	err := ctx.DB.Get(&app, "SELECT * FROM app WHERE fullname=? LIMIT 1", appName)
	log.Debug("AppNameExist: %s finish", appName)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
