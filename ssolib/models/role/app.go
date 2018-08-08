package role

import (
	//	"database/sql"
	"errors"

	//	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

var (
	ErrAppHasNoRole = errors.New("app has no admin role")
)

func CreateAppDefaultRole(ctx *models.Context, appId int, roleName string, fullName string) (*app.App, error) {
	a, err := app.GetApp(ctx, appId)
	if err != nil {
		return nil, err
	}
	r, err := CreateRoleWithGroup(ctx, roleName, fullName, appId, a.AdminGroupId)
	if err != nil {
		return nil, err
	}
	return SetAppRole(ctx, r.Id, appId)
}

func DeleteAppRole(ctx *models.Context, appId int) (*app.App, error) {
	return SetAppRole(ctx, appId, -1)
}

func SetAppRole(ctx *models.Context, roleId int, appId int) (*app.App, error) {
	tx := ctx.DB.MustBegin()
	_, err1 := tx.Exec(
		"UPDATE app SET admin_role_id=? WHERE id=?",
		roleId, appId)
	err2 := tx.Commit()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	return app.GetApp(ctx, appId)
}

func IsUserInAppAdminRole(ctx *models.Context, user iuser.User, appId int) (bool, group.MemberRole) {
	role, err := GetAppAdminRole(ctx, appId)
	if err != nil {
		log.Debug(err)
		if err == ErrRoleNotFound {
			a, err := app.GetApp(ctx, appId)
			if err != nil {
				log.Error(err)
				return false, group.NORMAL
			}
			g, err := group.GetGroup(ctx, a.AdminGroupId)
			if err != nil {
				log.Error(err)
				return false, group.NORMAL
			}
			ok, memberType, err := g.GetMember(ctx, user)
			if err != nil {
				log.Error(err)
				return false, group.NORMAL
			}
			return ok, memberType
		}
		return false, group.NORMAL
	}
	return IsUserInRole(ctx, user, role)
}

func GetAppAdminRole(ctx *models.Context, appId int) (*Role, error) {
	app, err := app.GetApp(ctx, appId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return GetRole(ctx, app.AdminGroupId)
}


