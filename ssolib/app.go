package ssolib

import (
	"net/http"
	"errors"

	"github.com/mijia/sweb/form"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"strconv"
	"github.com/laincloud/sso/ssolib/models/role"
	"github.com/laincloud/sso/Godeps/_workspace/src/github.com/mijia/sweb/log"
)

type AppsResource struct {
	server.BaseResource
}

func (ar AppsResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
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

func (ar AppsResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
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
		nameExist, err := app.AppNameExist(mctx, appSpec.FullName)
		if err != nil {
			panic(err)
		}
		if nameExist {
			return http.StatusBadRequest, errors.New("app name exists!")
		}

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

type AppResource struct {
	server.BaseResource
}

func (ar AppResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "read:app", func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		aId := params(ctx, "id")
		if aId == "" {
			return http.StatusBadRequest, "app id not given"
		}
		id, err := strconv.Atoi(aId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		queryApp, err := app.GetApp(mctx, id)
		if err != nil {
			if err == app.ErrAppNotFound {
				return http.StatusBadRequest, "app doesn't exist"
			}
			return http.StatusBadRequest, err
		}
		adminGroup, err := group.GetGroup(mctx, queryApp.AdminGroupId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		admins, err := adminGroup.GetGroupMembersID(mctx)
		uid := u.GetId()
		qualified := false
		for _,admin := range admins {
			if admin == uid {
				qualified = true
				break
			}
		}
		if !qualified {
			return http.StatusForbidden, "only admins of the app can read it"
		}
		resp := &App{
			Id:          queryApp.Id,
			FullName:    queryApp.FullName,
			Secret:      queryApp.Secret,
			RedirectUri: queryApp.RedirectUri,
		}
		return http.StatusOK, resp
	})
}

func (ar AppResource) Put(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:app", func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		aId := params(ctx, "id")
		if aId == "" {
			return http.StatusBadRequest, "app id not given"
		}
		id, err := strconv.Atoi(aId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		var appSpec struct {
			FullName    string `json:"fullname"`
			RedirectUri string `json:"redirect_uri"`
		}
		if err := form.ParamBodyJson(r, &appSpec); err != nil {
			return http.StatusBadRequest, ""
		}
		oldApp, err := app.GetApp(mctx, id)
		if err != nil {
			if err == app.ErrAppNotFound {
				return http.StatusBadRequest, "app doesn't exist"
			}
			return http.StatusBadRequest, err
		}
		if appSpec.FullName == "" && appSpec.RedirectUri == "" {
			return http.StatusBadRequest, "please enter fullname or redirecturi"
		}
		if appSpec.FullName == oldApp.FullName && appSpec.RedirectUri == oldApp.RedirectUri {
			return http.StatusBadRequest, "please enter different fullname or redirectUri"
		}
		if appSpec.FullName == "" && appSpec.RedirectUri == oldApp.RedirectUri {
			return http.StatusBadRequest, "please enter different redirectUri"
		}
		if appSpec.RedirectUri == "" && appSpec.FullName == oldApp.FullName {
			return http.StatusBadRequest, "please enter different fullname"
		}
		if appSpec.FullName != "" {
			if err := ValidateFullName(appSpec.FullName); err != nil {
				return http.StatusBadRequest, err
			}
		}
		if appSpec.RedirectUri != "" {
			if err := ValidateURI(appSpec.RedirectUri); err != nil {
				return http.StatusBadRequest, err
			}

		}
		if appSpec.FullName != "" && appSpec.FullName != oldApp.FullName {
			nameExist, err := app.AppNameExist(mctx, appSpec.FullName)
			if err != nil {
				panic(err)
			}
			if nameExist {
				return http.StatusBadRequest, errors.New("app name exists!")
			}
		}
		adminGroup, err := group.GetGroup(mctx, oldApp.AdminGroupId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		admins, err := adminGroup.GetGroupMembersID(mctx)
		uid := u.GetId()
		qualified := false
		for _,admin := range admins {
			if admin == uid {
				qualified = true
				break
			}
		}
		if !qualified {
			return http.StatusForbidden, "only admins of the app can modify it"
		}
		newApp, err := app.UpdateApp(mctx, &app.App{Id: id, FullName: appSpec.FullName, RedirectUri: appSpec.RedirectUri})
		if err != nil {
			return http.StatusBadRequest, err
		}
		resp := &App{
			Id:          newApp.Id,
			FullName:    newApp.FullName,
			RedirectUri: newApp.RedirectUri,
		}
		return http.StatusOK, resp
	})
}



func (ar AppResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:app", func(u iuser.User) (int, interface{}) {
		aId := params(ctx, "id")
		if aId == "" {
			return http.StatusBadRequest, "app id not given"
		}
		id, err := strconv.Atoi(aId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		mctx := getModelContext(ctx)
		oldApp, err := app.GetApp(mctx, id)
		if err != nil {
			if err == app.ErrAppNotFound {
				return http.StatusBadRequest, "app doesn't exist"
			}
			return http.StatusBadRequest, err
		}
		adminGroup, err := group.GetGroup(mctx, oldApp.AdminGroupId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		admins, err := adminGroup.GetGroupMembersID(mctx)
		uid := u.GetId()
		qualified := false
		for _,admin := range admins {
			if admin == uid {
				qualified = true
				break
			}
		}
		if !qualified {
			return http.StatusForbidden, "only admins of the app can modify it"
		}
		log.Debug("deleteing app")
		err = role.DeleteApp(mctx, id)
		if err != nil {
			return http.StatusBadRequest, err
		}
		return http.StatusNoContent, "app deleted"
	})
}


type App struct {
	Id          int    `json:"id"`
	FullName    string `json:"fullname"`
	Secret      string `json:"secret"`
	RedirectUri string `json:"redirect_uri"`
	AdminGroup  *Group `json:"admin_group"`
}

type AppInformation struct {
	server.BaseResource
}


func (ai AppInformation) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		Apps, err := app.ListApps(mctx)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, Apps
	})
}

