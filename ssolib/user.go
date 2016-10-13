package ssolib

import (
	"encoding/json"
	"net/http"

	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

type UserWithGroups struct {
	User   iuser.UserProfile `json:"user,omitempty"`
	Groups []string          `json:"groups"`
}

// 该函数必须是 *UserWithGroups, 否则会产生递归调用，即
// 绝对不可写成 func (ug UserWithGroups) MarshalJSON() ([]byte, error) {}
// 即想要调用也必须传入 *UserWithGroups
func (ug *UserWithGroups) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(ug.User)
	if err != nil {
		panic(err)
	}
	b2, err := json.Marshal(UserWithGroups{
		Groups: ug.Groups,
	})
	s := string(b)
	s2 := string(b2)
	ns := s[:len(s)-1] + "," + s2[1:]
	return []byte(ns), err
}

func (s *Server) UsersList(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	status, obj := requireScope(ctx, "read:user", func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		ub := getUserBackend(ctx)
		// FIXME 增加一个参数，以防数据库里条目过多
		users, err := ub.ListUsers(ctx)
		if err != nil {
			panic(err)
		}
		userIds := make([]int, len(users))
		for i, u := range users {
			userIds[i] = u.GetId()
		}
		groupMap, err := group.GetGroupsOfUserByIds(mctx, userIds)
		if err != nil {
			panic(err)
		}

		results := make([]*UserWithGroups, len(users))
		for i, u := range users {
			groups := make([]string, 0)
			if gs, ok := groupMap[u.GetId()]; ok {
				for _, g := range gs {
					groups = append(groups, g.Name)
				}
			}
			ug := &UserWithGroups{
				User:   u.GetProfile(),
				Groups: groups,
			}
			results[i] = ug
		}
		return http.StatusOK, results
	})
	w.WriteHeader(status)
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	w.Write(b)
	return ctx
}

type UserResource struct {
	server.BaseResource
}

func (ur UserResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	username := server.Params(ctx, "username")
	if username == "" {
		return http.StatusBadRequest, "username not given"
	}

	mctx := getModelContext(ctx)

	ub := getUserBackend(ctx)
	u, err := ub.GetUserByName(username)
	if err != nil {
		if err == iuser.ErrUserNotFound {
			return http.StatusNotFound, "no such user"
		}
		panic(err)
	}

	gs, err := group.GetGroupsOfUser(mctx, u)
	if err != nil {
		panic(err)
	}

	groups := make([]string, len(gs))
	for i, g := range gs {
		groups[i] = g.Name
	}

	ret := &UserWithGroups{
		User:   u.GetPublicProfile(),
		Groups: groups,
	}
	return http.StatusOK, ret
}

func (ur UserResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:user", func(currentUser iuser.User) (int, interface{}) {
		username := params(ctx, "username")
		if username == "" {
			return http.StatusBadRequest, "username not give"
		}

		mctx := getModelContext(ctx)
		ub := getUserBackend(ctx)
		u, err := ub.GetUserByName(username)
		if err != nil {
			panic(err)
		} else if u == nil {
			return http.StatusNotFound, "no such user"
		}

		isCurrentUserAdmin := false
		adminsGroup, err := group.GetGroupByName(mctx, "admins")
		if err != nil {
			panic(err)
		}
		// bug FIXME: should be the resursive group member.
		admins, err := adminsGroup.ListMembers(mctx)
		if err != nil {
			panic(err)
		}
		for _, admin := range admins {
			if admin.GetId() == currentUser.GetId() {
				isCurrentUserAdmin = true
				break
			}
		}

		if !isCurrentUserAdmin {
			return http.StatusForbidden, "have no permission"
		}

		err = group.RemoveUserFromAllGroups(mctx, u)
		if err != nil {
			panic(err)
		}

		err = ub.DeleteUser(u)
		if err != nil {
			panic(err)
		}

		return http.StatusNoContent, "User deleted"
	})
}

type MeResource struct {
	server.BaseResource
}

func GetUserWithGroups(ctx context.Context, u iuser.User) *UserWithGroups {
	log.Debugf("sso_debug: sso_me api begin")
	mctx := getModelContext(ctx)

	gs, err := group.GetGroupsOfUser(mctx, u)
	if err != nil {
		panic(err)
	}

	groups := make([]string, len(gs))
	for i, g := range gs {
		groups[i] = g.Name
	}

	ret := &UserWithGroups{
		User:   u.GetProfile(),
		Groups: groups,
	}
	log.Debugf("sso_debug: sso_me api end")
	return ret
}

func (mr MeResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		ret := GetUserWithGroups(ctx, u)
		return http.StatusOK, ret
	})
}
