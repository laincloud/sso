package ssolib

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/mijia/sweb/form"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/role"
)

type Role struct {
	RoleId int    `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Parent int    `json:"parent_id"`
}

type AppRolesOfUser struct {
	AppId       int        `json:"id"`
	AppFullName string     `json:"fullname"`
	Roles       []UserRole `json:"roles"`
}

type UserRole struct {
	RoleName string `json:"name"`
	RoleId   int    `json:"id"`
	Type     string `json:"type"`
	Parent   int    `json:"parent_id"`
}

type UserRoles []UserRole

// AppRole: can get roles of everyone related to the app, not only the app owner
type AppRoleResource struct {
	server.BaseResource
}

func (ar AppRoleResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "read:app", func(u iuser.User) (int, interface{}) {
		// first get all the group
		// some groups are app admin groups
		// some groups are roles
		// note that some app uses other's role tree, maybe we should return these apps TODO.
		mctx := getModelContext(ctx)
		userGroups, err := group.GetGroupRolesOfUser(mctx, u)
		gIds := []int{}
		index := make(map[int]int)
		for i, g := range userGroups {
			gIds = append(gIds, g.Id)
			index[g.Id] = i
		}
		roles, err := role.GetRolesByGroupIds(mctx, gIds)
		if err != nil {
			return http.StatusNotFound, err
		}
		appRoles := []UserRoles{}
		appURs := make(map[int]int)
		urApps := make(map[int]int)
		appNum := 0

		for _, r := range roles {
			var rType string
			if userGroups[index[r.Id]].Role == group.ADMIN {
				rType = "admin"
			} else {
				rType = "normal"
			}
			appId := r.AppId
			ur := UserRole{
				RoleName: r.Name,
				RoleId:   r.Id,
				Type:     rType,
				Parent:   r.SuperRoleId,
			}
			if _, ok := appURs[appId]; ok {
				tmp := appRoles[appURs[appId]]
				tmp = append(tmp, ur)
				appRoles[appURs[appId]] = tmp
			} else {
				tmp := []UserRole{}
				tmp = append(tmp, ur)
				appRoles = append(appRoles, UserRoles(tmp))
				appURs[appId] = appNum
				urApps[appNum] = appId
				appNum += 1
			}
		}
		ret := []AppRolesOfUser{}
		log.Debug(appRoles)
		log.Debug(appURs)
		for i, ur := range appRoles {
			appId := urApps[i]
			a, err := app.GetApp(mctx, appId)
			if err != nil {
				log.Debug(appId)
				panic(err)
			}
			tmp := AppRolesOfUser{
				AppId:       appId,
				AppFullName: a.FullName,
				Roles:       ur,
			}
			ret = append(ret, tmp)
		}
		return http.StatusOK, ret

	})
}

type AppRole struct {
	AppId    int    `json:"app_id"`
	RoleName string `json:"role_name"`
	RoleId   int    `json:"role_id"`
}

var (
	ErrNotAdmin = errors.New("only user in app's admin role can modify the app's role tree.")
)

func (ar AppRoleResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	// create the default app root role, i.e. use the app group or
	// use some existed role as a root role, in this case the role id is different from the admin group id
	mctx := getModelContext(ctx)

	req := AppRole{}
	if err := form.ParamBodyJson(r, &req); err != nil {
		return http.StatusBadRequest, err
	}
	log.Debug(req)
	if req.AppId == 0 {
		return http.StatusBadRequest, errors.New("app id is required")
	} else if _, err := app.GetApp(mctx, req.AppId); err != nil {
		return http.StatusBadRequest, err
	}
	if req.RoleId == 0 && req.RoleName == "" {
		return http.StatusBadRequest, errors.New("role id is required if use old role, otherwise role name is required ")
	}
	if req.RoleId != 0 && req.RoleName != "" {
		return http.StatusBadRequest, errors.New("should either set role id or role name, both existed parameters are confused.")
	}

	secret := r.Header.Get("secret")
	client, _ := app.GetApp(mctx, req.AppId)
	if client.GetSecret() == secret {
		if req.RoleName != "" {
			app, err := role.CreateAppDefaultRole(mctx, req.AppId, req.RoleName, req.RoleName)
			if err != nil {
				return http.StatusBadRequest, err
			} else {
				app.Secret = ""
				return http.StatusOK, app
			}
		} else {
			targetRole, err := role.GetRole(mctx, req.RoleId)
			if err != nil {
				return http.StatusBadRequest, err
			}
			if targetRole.SuperRoleId != -1 {
				return http.StatusBadRequest, errors.New(
					"only the root role can be used as a root role by other apps")
			}
			app, err := role.SetAppRole(mctx, req.RoleId, req.AppId)
			if err != nil {
				return http.StatusBadRequest, err
			} else {
				return http.StatusOK, app
			}
		}
	}

	return requireScope(ctx, "write:app", func(u iuser.User) (int, interface{}) {

		// ACL
		if ok, roleType := role.IsUserInAppAdminRole(mctx, u, req.AppId); ok {
			if roleType != group.ADMIN {
				return http.StatusForbidden, ErrNotAdmin
			}
		} else {
			return http.StatusForbidden, ErrNotAdmin
		}

		if r, err := role.GetAppAdminRole(mctx, req.AppId); r != nil {
			if err != role.ErrRoleNotFound {
				log.Error(err)
			}
			return http.StatusBadRequest, errors.New("admin role already exists, please delete it.")
		}

		if req.RoleName != "" {
			app, err := role.CreateAppDefaultRole(mctx, req.AppId, req.RoleName, req.RoleName)
			if err != nil {
				return http.StatusBadRequest, err
			} else {
				app.Secret = ""
				return http.StatusOK, app
			}
		} else {
			targetRole, err := role.GetRole(mctx, req.RoleId)
			if err != nil {
				return http.StatusBadRequest, err
			}
			if targetRole.SuperRoleId != -1 {
				return http.StatusBadRequest, errors.New(
					"only the root role can be used as a root role by other apps")
			}
			app, err := role.SetAppRole(mctx, req.RoleId, req.AppId)
			if err != nil {
				return http.StatusBadRequest, err
			} else {
				return http.StatusOK, app
			}
		}

	})
}

func (ar AppRoleResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	mctx := getModelContext(ctx)

	req := AppRole{}
	if err := form.ParamBodyJson(r, &req); err != nil {
		return http.StatusBadRequest, err
	}
	if req.AppId == 0 {
		return http.StatusBadRequest, errors.New("app id is required")
	}
	secret := r.Header.Get("secret")
	client, _ := app.GetApp(mctx, req.AppId)
	if client.GetSecret() == secret {
		if r, err := role.GetAppAdminRole(mctx, req.AppId); r == nil {
			if err == role.ErrRoleNotFound {
				return http.StatusBadRequest, errors.New("admin role not exists.")
			} else {
				panic(err)
			}
		}

		app, err := role.DeleteAppRole(mctx, req.AppId)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		return http.StatusOK, app
	}
	return requireScope(ctx, "write:app", func(u iuser.User) (int, interface{}) {

		// ACL
		if ok, roleType := role.IsUserInAppAdminRole(mctx, u, req.AppId); ok {
			if roleType != group.ADMIN {
				return http.StatusForbidden, ErrNotAdmin
			}
		} else {
			return http.StatusForbidden, ErrNotAdmin
		}

		if r, err := role.GetAppAdminRole(mctx, req.AppId); r == nil {
			if err == role.ErrRoleNotFound {
				return http.StatusBadRequest, errors.New("admin role not exists.")
			} else {
				panic(err)
			}
		}

		app, err := role.DeleteAppRole(mctx, req.AppId)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		return http.StatusOK, app

	})
}

type RolesResource struct {
	server.BaseResource
}

type RoleMembers struct {
	role.Role
	Type    string       `json:"type"`
	Members []RoleMember `json:"members"`
}

func (rsr RolesResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "read:role", func(u iuser.User) (int, interface{}) {
		r.ParseForm()
		sAppId := r.Form.Get("app_id")
		if sAppId == "" {
			return http.StatusBadRequest, "app_id required"
		}
		appId, err := strconv.Atoi(sAppId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		userName := r.Form.Get("username")
		queryUser := u
		if userName != "" {
			ub := getUserBackend(ctx)
			queryUser, err = ub.GetUserByName(userName)
			if err == iuser.ErrUserNotFound {
				return http.StatusNotFound, err
			} else {
				panic(err)
			}
		}
		mctx := getModelContext(ctx)
		roleMembers, err := role.GetAllRoleMembers(mctx, queryUser, appId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		ret := []RoleMembers{}
		for _, rM := range roleMembers {

			memList := []RoleMember{}
			for _, gM := range rM.Members {
				var sType string
				if gM.Role == group.ADMIN {
					sType = "admin"
				} else {
					sType = "normal"
				}
				memList = append(memList, RoleMember{
					UserName:   gM.GetName(),
					MemberType: sType,
				})
			}

			var sType string
			if rM.Type == group.ADMIN {
				sType = "admin"
			} else {
				sType = "normal"
			}
			tmp := RoleMembers{
				Role:    rM.Role,
				Type:    sType,
				Members: memList,
			}
			ret = append(ret, tmp)
		}
		log.Debug(ret)
		return http.StatusOK, ret

	})
}

type RoleReq struct {
	AppId  int    `json:"app_id"`
	Name   string `json:"name"`
	Desc   string `json:"description"`
	Parent int    `json:"parent_id"`
}

func (rsr RolesResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:role", func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		roleReq := RoleReq{}
		if err := form.ParamBodyJson(r, &roleReq); err != nil {
			return http.StatusBadRequest, err
		}
		appId := roleReq.AppId
		if _, err := app.GetApp(mctx, appId); err != nil {
			return http.StatusBadRequest, err
		}
		if _, err := role.GetRoleByName(mctx, roleReq.Name, appId); err != role.ErrRoleNotFound {
			return http.StatusConflict, err
		}
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		if ok && mType == group.ADMIN {
			resp, err := role.CreateRoleWithoutGroup(mctx, roleReq.Name, roleReq.Desc, roleReq.AppId, roleReq.Parent)
			if err != nil {
				return http.StatusBadRequest, err
			}
			return http.StatusOK, resp
		} else {
			return http.StatusForbidden, "only the admin of the root role can create role"
		}
	})
}

type RoleResource struct {
	server.BaseResource
}

func (rr RoleResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "read:role", func(u iuser.User) (int, interface{}) {
		rId := params(ctx, "id")
		if rId == "" {
			return http.StatusBadRequest, "role id required"
		}
		id, err := strconv.Atoi(rId)
		if err != nil {
			return http.StatusBadRequest, "role id invalid"
		}
		mctx := getModelContext(ctx)
		role, err := role.GetRole(mctx, id)
		if err != nil {
			return http.StatusNotFound, err
		}
		return http.StatusOK, role
	})
}

type RoleModifyReq struct {
	Name   string `json:"name"`
	Desc   string `json:"description"`
	Parent int    `json:"parent_id"`
}

func (rr RoleResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:role", func(u iuser.User) (int, interface{}) {
		rId := params(ctx, "id")
		if rId == "" {
			return http.StatusBadRequest, "role id required"
		}
		id, err := strconv.Atoi(rId)
		if err != nil {
			return http.StatusBadRequest, "role id invalid"
		}
		mctx := getModelContext(ctx)
		oldRole, err := role.GetRole(mctx, id)
		if err != nil {
			return http.StatusNotFound, err
		}

		req := RoleModifyReq{}
		if err := form.ParamBodyJson(r, &req); err != nil {
			return http.StatusBadRequest, err
		}

		appId := oldRole.AppId
		if ok, roleType := role.IsUserInAppAdminRole(mctx, u, appId); ok {
			if roleType != group.ADMIN {
				return http.StatusForbidden, ErrNotAdmin.Error()
			}
		} else {
			return http.StatusForbidden, ErrNotAdmin.Error()
		}

		newRole, err := role.UpdateRole(mctx, oldRole.Id, req.Name, req.Desc, req.Parent)
		if err != nil {
			return http.StatusBadRequest, err
		} else {
			return http.StatusOK, newRole
		}

	})
}

func (rr RoleResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:role", func(u iuser.User) (int, interface{}) {
		rId := params(ctx, "id")
		if rId == "" {
			return http.StatusBadRequest, "role id required"
		}
		id, err := strconv.Atoi(rId)
		if err != nil {
			return http.StatusBadRequest, "role id invalid"
		}
		mctx := getModelContext(ctx)
		oldRole, err := role.GetRole(mctx, id)
		if err != nil {
			return http.StatusNotFound, err
		}

		appId := oldRole.AppId
		if ok, roleType := role.IsUserInAppAdminRole(mctx, u, appId); ok {
			if roleType != group.ADMIN {
				return http.StatusForbidden, ErrNotAdmin.Error()
			}
		} else {
			return http.StatusForbidden, ErrNotAdmin.Error()
		}

		err = role.DeleteRole(mctx, id)
		if err != nil {
			return http.StatusBadRequest, err
		} else {
			return http.StatusNoContent, "role deleted"
		}

	})
}

type RoleMemberResource struct {
	server.BaseResource
}

type RoleMemberType struct {
	MemberType string `json:"type"`
}

func (rmr RoleMemberResource) Put(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:role", func(currentUser iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)

		roleId := params(ctx, "id")
		if roleId == "" {
			return http.StatusBadRequest, "role id required"
		}
		id, err := strconv.Atoi(roleId)
		if err != nil {
			return http.StatusBadRequest, "role id invalid"
		}
		username := params(ctx, "username")
		if username == "" {
			return http.StatusBadRequest, "no username given"
		}
		memberType := RoleMemberType{}
		if err := form.ParamBodyJson(r, &memberType); err != nil {
			return http.StatusBadRequest, err
		}
		var mrole group.MemberRole
		if memberType.MemberType == "admin" {
			mrole = group.ADMIN
		} else if memberType.MemberType == "normal" || memberType.MemberType == "" {
			mrole = group.NORMAL
		} else {
			return http.StatusBadRequest, "unknown role"
		}

		_, err = role.GetRole(mctx, id)
		if err != nil {
			if err == role.ErrRoleNotFound {
				return http.StatusBadRequest, "role not found"
			}
			panic(err)
		}

		g, err := group.GetGroup(mctx, id)
		switch {
		case err == group.ErrGroupNotFound:
			return http.StatusNotFound, "no such group"
		case err != nil:
			panic(err)
		}

		ub := getUserBackend(ctx)
		u, uerr := ub.GetUserByName(username)

		if uerr != nil {
			return http.StatusNotFound, "no such user"
		}

		ok, currentUserRole, err := g.GetMember(mctx, currentUser)
		if err != nil {
			panic(err)
		}
		if !(ok && currentUserRole == group.ADMIN) {
			return http.StatusForbidden, "Not group admin"
		}

		directMembers, err := g.ListMembers(mctx)
		if err != nil {
			panic(err)
		}
		isAlreadyMember := false
		for _, m := range directMembers {
			if m.GetId() == u.GetId() {
				isAlreadyMember = true
				if m.Role != mrole {
					if err := g.UpdateMember(mctx, u, mrole); err != nil {
						panic(err)
					}
				}
				break
			}
		}
		if !isAlreadyMember {
			if err := g.AddMember(mctx, u, mrole); err != nil {
				panic(err)
			}

		}
		return http.StatusOK, "member added"
	})
}

func (rmr RoleMemberResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:role", func(currentUser iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)

		roleId := params(ctx, "id")
		if roleId == "" {
			return http.StatusBadRequest, "role id required"
		}
		id, err := strconv.Atoi(roleId)
		if err != nil {
			return http.StatusBadRequest, "role id invalid"
		}
		username := params(ctx, "username")
		if username == "" {
			return http.StatusBadRequest, "no username given"
		}

		_, err = role.GetRole(mctx, id)
		if err != nil {
			if err == role.ErrRoleNotFound {
				return http.StatusBadRequest, "role not found"
			}
			panic(err)
		}

		g, err := group.GetGroup(mctx, id)
		switch {
		case err == group.ErrGroupNotFound:
			return http.StatusNotFound, "no such group"
		case err != nil:
			panic(err)
		}

		ub := getUserBackend(ctx)
		u, uerr := ub.GetUserByName(username)

		if uerr != nil {
			return http.StatusNotFound, "no such user"
		}

		ok, currentUserRole, err := g.GetMember(mctx, currentUser)
		if err != nil {
			panic(err)
		}
		if !(ok && currentUserRole == group.ADMIN) {
			return http.StatusForbidden, "Not group admin"
		}
		if err := g.RemoveMember(mctx, u); err != nil {
			panic(err)
		}

		return http.StatusNoContent, "member deleted"
	})

}

type RoleMember struct {
	UserName   string `json:"user"`
	MemberType string `json:"type"`
}

type RoleMembersReq struct {
	RoleId     int
	Action     string
	MemberList []RoleMember `json:"members"`
}

func (s *Server) RoleMembers(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	status, v := func() (int, string) {
		if r.Method != "POST" {
			return http.StatusBadRequest, "only support POST for this url"
		}
		s, iface := requireScope(ctx, "write:role", func(u iuser.User) (int, interface{}) {
			mctx := getModelContext(ctx)

			ub := getUserBackend(ctx)
			req := RoleMembersReq{}
			if err := form.ParamBodyJson(r, &req); err != nil {
				return http.StatusBadRequest, err.Error()
			}
			_, err := role.GetRole(mctx, req.RoleId)
			if err != nil {
				if err == role.ErrRoleNotFound {
					return http.StatusBadRequest, "role not found"
				}
				panic(err)
			}
			for _, m := range req.MemberList {
				if m.MemberType != "admin" && m.MemberType != "" && m.MemberType != "normal" {
					return http.StatusBadRequest, "unknown member type: " + m.MemberType
				}
				username := m.UserName
				_, uerr := ub.GetUserByName(username)

				if uerr != nil {
					return http.StatusNotFound, "no such user: " + username
				}

			}
			g, err := group.GetGroup(mctx, req.RoleId)
			switch {
			case err == group.ErrGroupNotFound:
				return http.StatusNotFound, "no such group"
			case err != nil:
				panic(err)
			}
			ok, currentUserRole, err := g.GetMember(mctx, u)
			if err != nil {
				panic(err)
			}
			if !(ok && currentUserRole == group.ADMIN) {
				return http.StatusForbidden, "Not group admin"
			}
			directMembers, err := g.ListMembers(mctx)
			if err != nil {
				panic(err)
			}

			if req.Action == "add" {
				for _, m := range req.MemberList {

					var mrole group.MemberRole
					if m.MemberType == "admin" {
						mrole = group.ADMIN
					} else if m.MemberType == "normal" || m.MemberType == "" {
						mrole = group.NORMAL
					} else {
						panic("check failed")
					}

					uAdding, _ := ub.GetUserByName(m.UserName)

					isAlreadyMember := false
					for _, oldM := range directMembers {

						if oldM.GetId() == uAdding.GetId() {
							isAlreadyMember = true
							if oldM.Role != mrole {
								if err := g.UpdateMember(mctx, uAdding, mrole); err != nil {
									panic(err)
								}
							}
							break
						}
					}
					if !isAlreadyMember {
						if err := g.AddMember(mctx, uAdding, mrole); err != nil {
							panic(err)
						}
					}
				}
				return http.StatusOK, "members added"

			} else if req.Action == "delete" {
				for _, m := range req.MemberList {
					uDel, _ := ub.GetUserByName(m.UserName)
					if err := g.RemoveMember(mctx, uDel); err != nil {
						panic(err)
					}
				}
				return http.StatusOK, "members deleted"
			} else {
				return http.StatusBadRequest, "action should be either add or delete"
			}
		})
		return s, iface.(string)
	}()

	var data []byte
	var err error
	if status != http.StatusBadRequest && status != http.StatusConflict {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		apiError := ApiError{v, v}
		data, err = json.MarshalIndent(apiError, "", "  ")
	}
	if err != nil {
		status = http.StatusInternalServerError
		data = []byte(err.Error())
	}
	data = append(data, '\n')
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	w.Write(data)

	return ctx
}
