package ssolib

import (
	"encoding/json"
	"net/http"

	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"strings"
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

		adminsGroup, err := group.GetGroupByName(mctx, "admins")
		if err != nil {
			panic(err)
		}
		isAdmin, _, _ := adminsGroup.GetMember(mctx, u)

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
			var profile iuser.UserProfile
			if isAdmin {
				profile = u.GetProfile()
			} else {
				profile = u.GetPublicProfile()
			}
			ug := &UserWithGroups{
				User:   profile,
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

//Batch return users' profiles and groups(optional)
func (s *Server) BatchUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	log.Debugf("sso_debug: batch-users api begin.")
	defer log.Debugf("sso_debug: batch-users api end.")
	mctx := getModelContext(ctx)
	ub := getUserBackend(ctx)

	r.ParseForm()
	names := r.Form.Get("name")
	if names == "" {
		http.Error(w, "name not given", http.StatusBadRequest)
		return ctx
	}
	nameSlice := strings.Split(names, ",")
	withGroup := false
	if rGroup := r.Form.Get("group"); rGroup == "true" {
		withGroup = true
	}

	currentUser := getCurrentUser(ctx)
	isAdmin := false
	if currentUser != nil {
		adminsGroup, err := group.GetGroupByName(mctx, "admins")
		if err != nil {
			panic(err)
		}
		isAdmin, _, _ = adminsGroup.GetMember(mctx, currentUser)
	}

	profiles := make([]iuser.UserProfile, 0, len(nameSlice))
	detailedProfiles := make([]*UserWithGroups, 0, len(nameSlice))
	for _, name := range nameSlice {
		u, err := ub.GetUserByName(name)
		if err != nil {
			if err == iuser.ErrUserNotFound {
				continue
			}
			panic(err)
		}
		isSelf := false
		if currentUser != nil && currentUser.GetName() == u.GetName() {
			isSelf = true
		}
		if currentUser == nil || (!isAdmin && !isSelf) {
			profiles = append(profiles, u.GetPublicProfile())
		} else {
			profiles = append(profiles, u.GetProfile())
		}

		if withGroup {
			gs, err := group.GetGroupsOfUser(mctx, u)
			if err != nil {
				panic(err)
			}
			groups := make([]string, len(gs))
			for i, g := range gs {
				groups[i] = g.Name
			}
			ug := &UserWithGroups{
				User:   profiles[len(profiles)-1],
				Groups: groups,
			}
			detailedProfiles = append(detailedProfiles, ug)
		}
	}

	if len(profiles) == 0 {
		http.Error(w, "no such users", http.StatusNotFound)
		return ctx
	}

	if withGroup {
		b, err := json.Marshal(detailedProfiles)
		if err != nil {
			panic(err)
		}
		w.Write(b)
	} else {
		b, err := json.Marshal(profiles)
		if err != nil {
			panic(err)
		}
		w.Write(b)
	}
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
	if err := r.ParseForm(); err != nil {
		log.Debug(err)
		return http.StatusBadRequest, err
	}
	DatabaseOnly := r.Form.Get("database")
	database := false
	if DatabaseOnly == "true" {
		database = true
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
	if database {
		ggs := []group.Group{}
		for _, g := range gs {
			if g.GroupType == 0 {
				ggs = append(ggs,g)
			}
		}
		gs = ggs
	}
	groups := make([]string, len(gs))
	for i, g := range gs {
		groups[i] = g.Name
	}

	var profile iuser.UserProfile
	currentUser := getCurrentUser(ctx)
	isAdmin := false
	isSelf := false
	if currentUser != nil {
		adminsGroup, err := group.GetGroupByName(mctx, "admins")
		if err != nil {
			panic(err)
		}
		isAdmin, _, _ = adminsGroup.GetMember(mctx, currentUser)
		if currentUser.GetName() == u.GetName() {
			isSelf = true
		}
	}

	if currentUser == nil || (!isAdmin && !isSelf) {
		profile = u.GetPublicProfile()
	} else {
		profile = u.GetProfile()
	}
	ret := &UserWithGroups{
		User:   profile,
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

		adminsGroup, err := group.GetGroupByName(mctx, "admins")
		if err != nil {
			panic(err)
		}

		ok, _, _ := adminsGroup.GetMember(mctx, currentUser)
		if !ok {
			return http.StatusForbidden, "have no permission"
		}

		err = group.RemoveUserFromAllGroups(mctx, u)
		if err != nil {
			panic(err)
		}

		err = ub.DeleteUser(u)
		if err != nil {
			// since some backend delete user means delete user from group
			log.Debug(err)
			return http.StatusNoContent, "User deleted from groups but " + err.Error()
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
