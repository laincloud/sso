package ssolib

import (
	"net/http"

	"github.com/mijia/sweb/form"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

type AppResource struct {
	server.BaseResource
}

func (ar AppResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "read:app", func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)

		userGroups, err := group.GetGroupsOfUser(mctx, u)
		if err != nil {
			panic(err)
		}
		if len(userGroups) == 0 {
			return http.StatusOK, []App{}
		}

		groupIds := make([]int, len(userGroups))
		for i := range groupIds {
			groupIds[i] = userGroups[i].Id
		}
		apps, err := app.ListAppsByAdminGroupIds(mctx, groupIds)
		if err != nil {
			panic(err)
		}
		adminGroups := make(map[int]group.Group)
		for _, g := range userGroups {
			adminGroups[g.Id] = g
		}

		results := make([]App, len(apps))
		for i := range apps {
			group := adminGroups[apps[i].AdminGroupId]
			ag := groupFromModel(&group)
			results[i] = App{
				Id:          apps[i].Id,
				FullName:    apps[i].FullName,
				Secret:      apps[i].SecretString(),
				RedirectUri: apps[i].RedirectUri,
				AdminGroup:  ag,
			}
		}
		return http.StatusOK, results
	})
}

func (ar AppResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:app", func(u iuser.User) (int, interface{}) {
		var appSpec struct {
			FullName    string `json:"fullname"`
			RedirectUri string `json:"redirect_uri"`
		}
		if err := form.ParamBodyJson(r, &appSpec); err != nil {
			return http.StatusBadRequest, ""
		}

		if err := ValidateFullName(appSpec.FullName); err != nil {
			return http.StatusBadRequest, err
		}

		if err := ValidateURI(appSpec.RedirectUri); err != nil {
			return http.StatusBadRequest, err
		}

		mctx := getModelContext(ctx)
		a, err := app.CreateApp(mctx, &app.App{FullName: appSpec.FullName, RedirectUri: appSpec.RedirectUri}, u)
		if err != nil {
			panic(err)
		}

		adminGroup, err := group.GetGroup(mctx, a.AdminGroupId)
		if err != nil {
			panic(err)
		}

		result := &App{
			Id:          a.Id,
			FullName:    a.FullName,
			Secret:      a.SecretString(),
			RedirectUri: a.RedirectUri,
			AdminGroup:  groupFromModel(adminGroup),
		}

		return http.StatusCreated, result
	})
}

type App struct {
	Id          int    `json:"id"`
	FullName    string `json:"fullname"`
	Secret      string `json:"secret"`
	RedirectUri string `json:"redirect_uri"`
	AdminGroup  *Group `json:"admin_group"`
}
