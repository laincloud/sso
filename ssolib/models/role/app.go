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

func DeleteApp(ctx *models.Context, id int) ( error) {
	tx := ctx.DB.MustBegin()
	roles := []Role{}
	err1 := ctx.DB.Select(&roles, "SELECT id FROM role WHERE app_id=?", id)
	if err1 != nil {
		return err1
	}
	_, err2 := tx.Exec("DELETE FROM role WHERE app_id=?", id)
	if err2 != nil {
		return err2
	}
	_, err3 := tx.Exec("DELETE FROM resource WHERE app_id=?", id)
	if err3 != nil {
		return err3
	}
	for _,r := range roles {
		if IsLeafRole(ctx, r.Id) {
			_, err := tx.Exec("DELETE FROM role_resource WHERE role_id=?", id)
			if err != nil {
				return err
			}
		}
	}
	_, err4 := tx.Exec("DELETE FROM app WHERE id=?", id)
	if err4 != nil {
		return err4
	}
	for _,r := range roles {
		g, err := group.GetGroup(ctx, r.Id)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		err = group.DeleteGroup(ctx, g)
		if err != nil {
			panic(err)
		}
	}
	if err5 := tx.Commit(); err5 != nil {
		log.Debug(err5)
		return err5
	}
	return nil
}