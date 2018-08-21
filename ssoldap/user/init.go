package user

import (
	"fmt"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/mijia/sweb/log"
)

var InitAdmin string

func (ub *UserBack) InitModel(ctx interface{}) {

	ub.InitDatabase()

	mctx, ok := ctx.(*models.Context)

	if !ok {
		panic("Unexpected context of models")
	}

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

	iadmin, err := ub.GetUserByName(InitAdmin)

	if err != nil {
		log.Debug(InitAdmin)
		log.Error(err)
		panic(err)
	}

	err = adminsGroup.AddMember(mctx, iadmin, group.ADMIN)
	if err != nil {
		panic(err)
	}

	appSpec := &app.App{
		FullName:    "SSO-ldap",
		RedirectUri: mctx.SSOSiteURL.String(),
		Secret:      "admin",
	}
	_, err = app.CreateApp(mctx, appSpec, iadmin)
	if err != nil {
		panic(err)
	}

	siteAppSpec := &app.App{
		FullName:    "SSO-ldap-Site",
		RedirectUri: fmt.Sprintf("%s/spa/admin/authorize", mctx.SSOSiteURL.String()),
		Secret:      "sso_admin",
	}
	_, err = app.CreateApp(mctx, siteAppSpec, iadmin)
	if err != nil {
		panic(err)
	}
}
