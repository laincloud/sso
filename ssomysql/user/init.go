package user

import (
	"database/sql"
	"fmt"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
)

func (ub *UserBack) InitModel(ctx interface{}) {
	var mctx *models.Context

	mctx, ok := ctx.(*models.Context)

	if !ok {
		panic("Unexpected context of models")
	}

	ub.InitDatabase()

	adminsGroup, err := group.GetGroupByName(mctx, "admins")
	if err == nil {
		// admins group exists, already initialized
		return
	}

	if err != group.ErrGroupNotFound {
		panic(err)
	}

	adminsGroup, err = group.CreateGroup(mctx, &group.Group{Name: "admins"})
	if err != nil {
		panic(err)
	}

	lainGroup, err := group.CreateGroup(mctx, &group.Group{Name: "lain"})
	if err != nil {
		panic(err)
	}

	admin := &User{
		Name:         "admin",
		FullName:     "Admin",
		Email:        sql.NullString{String: "admin@example.com", Valid: true},
		PasswordHash: []byte("admin"),
	}
	err = ub.CreateUser(admin, false)
	if err != nil {
		panic(err)
	}
	iadmin, err := ub.GetUserByName("admin")
	if err != nil {
		panic(err)
	}

	err = adminsGroup.AddMember(mctx, iadmin, group.ADMIN)
	if err != nil {
		panic(err)
	}

	err = lainGroup.AddMember(mctx, iadmin, group.ADMIN)
	if err != nil {
		panic(err)
	}

	appSpec := &app.App{
		FullName:    "SSO",
		RedirectUri: mctx.SSOSiteURL.String(),
		Secret:      "admin",
	}
	_, err = app.CreateApp(mctx, appSpec, iadmin)
	if err != nil {
		panic(err)
	}

	siteAppSpec := &app.App{
		FullName:    "SSO-Site",
		RedirectUri: fmt.Sprintf("%s/spa/admin/authorize", mctx.SSOSiteURL.String()),
		Secret:      "sso_admin",
	}
	_, err = app.CreateApp(mctx, siteAppSpec, iadmin)
	if err != nil {
		panic(err)
	}

	lainCliSpec := &app.App{
		FullName:    "lain-client",
		RedirectUri: "https://example.com",
		Secret:      "lain-cli_admin",
	}
	_, err = app.CreateApp(mctx, lainCliSpec, iadmin)
	if err != nil {
		panic(err)
	}

}
