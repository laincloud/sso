package ssolib

import (
	"database/sql"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/mijia/sweb/form"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"time"
)

var (
	params = server.Params
)

type GroupsResource struct {
	server.BaseResource
}

type Group struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
}

type BackendGroup struct {
	Group
	Backend iuser.GroupType `json:"backend"`
	Rules   string          `json:"rules"`
}

type GroupWithMembers struct {
	Name         string            `json:"name"`
	FullName     string            `json:"fullname"`
	Members      []MemberRole      `json:"members"`
	GroupMembers []GroupMemberRole `json:"group_members"`
}

type MemberRole struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type GroupMemberRole struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Role     string `json:"role"`
}

type GroupWithRole struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Role     string `json:"role"`
}

func groupFromModel(g *group.Group) *Group {
	return &Group{
		Name:     g.Name,
		FullName: g.FullName,
	}
}

func (g *Group) Validate() error {
	if err := ValidateSlug(g.Name, 128); err != nil {
		return err
	}
	if g.FullName != "" {
		if err := ValidateFullName(g.FullName); err != nil {
			return err
		}
	}

	return nil
}

func (bg *BackendGroup) Validate() error {
	if err := bg.Group.Validate(); err != nil {
		return err
	}
	if bg.Backend >= 2 {
		return group.ErrBackendUnsupported
	}
	if bg.Rules != "" {
		//TODO check ldap rules
	}
	return nil
}

// 登录的用户得到自己所在的 groups 列表
func (gr GroupsResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "read:group", func(u iuser.User) (int, interface{}) {
		r.ParseForm()
		mctx := getModelContext(ctx)
		all := r.Form.Get("all")
		if all == "true" {
			g, err := group.GetAllDateBaseGroup(mctx)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, g
		}
		groups, err := group.GetGroupRolesOfUser(mctx, u)
		if err != nil {
			panic(err)
		}
		groupRoles := make([]GroupWithRole, len(groups))
		for i, g := range groups {
			role := "normal"
			if g.Role == group.ADMIN {
				role = "admin"
			}
			groupRoles[i] = GroupWithRole{
				Name:     g.Name,
				FullName: g.FullName,
				Role:     role,
			}
		}
		return http.StatusOK, groupRoles
	})
}

func (gr GroupsResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:group", func(u iuser.User) (int, interface{}) {
		req := BackendGroup{}
		if err := form.ParamBodyJson(r, &req); err != nil {
			return http.StatusBadRequest, err
		}

		if err := req.Validate(); err != nil {
			return http.StatusBadRequest, err
		}

		if req.FullName == "" {
			req.FullName = req.Name
		}

		var ubg iuser.BackendWithGroup
		var ok bool
		if req.Backend == iuser.BACKENDGROUP {
			ub := getUserBackend(ctx)
			if ubg, ok = ub.(iuser.BackendWithGroup); ok {
				// 注意 Rules 的唯一性
				success, g, err := ubg.CreateBackendGroup(req.Name, req.Rules)
				if !success {
					if g != nil {
						req.Name = g.GetName()
						return http.StatusConflict, err
					} else {
						return http.StatusBadRequest, err
					}
				}
			} else {
				log.Debug("type assertion fails")
				return http.StatusBadRequest, group.ErrBackendUnsupported
			}
		}

		mctx := getModelContext(ctx)
		g, err := group.CreateGroup(mctx, &group.Group{
			Name:      req.Name,
			FullName:  req.FullName,
			GroupType: req.Backend,
		})
		if err != nil {
			if mysqlError, ok := err.(*mysql.MySQLError); ok {
				if mysqlError.Number == 1062 {
					// for "Error 1062: Duplicate entry ..."
					log.Info(err.Error())
					return http.StatusConflict, "group already exists"
				}
			}
			panic(err)
		}

		if req.Backend == iuser.SSOLIBGROUP {
			if err = g.AddMember(mctx, getCurrentUser(ctx), group.ADMIN); err != nil {
				panic(err)
			}
		} else if req.Backend == iuser.BACKENDGROUP {
			if err = ubg.SetBackendGroupId(g.Name, g.Id); err != nil {
				panic(err)
			}
		} else {
			panic("check backend failed")
		}
		return http.StatusCreated, groupFromModel(g)
	})
}

type GroupResource struct {
	server.BaseResource
}

func (gr GroupResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	groupname := params(ctx, "groupname")
	if groupname == "" {
		return http.StatusBadRequest, "groupname not given"
	}

	mctx := getModelContext(ctx)

	g, err := group.GetGroupIdByName(mctx, groupname)
	if err != nil {
		if err == group.ErrGroupNotFound {
			return http.StatusNotFound, "No such group"
		}
		panic(err)
	}

	members, err := g.ListMembers(mctx)

	if err == iuser.ErrMethodNotSupported {
		members = []group.Member{}
	} else if err != nil {
		panic(err)
	}

	memberRoles := make([]MemberRole, len(members))
	for i, m := range members {
		log.Debug(i, m)
		memberRoles[i] = MemberRole{Name: m.GetName()}
		if m.Role == group.ADMIN {
			memberRoles[i].Role = "admin"
		}
	}

	gMembers, err := g.ListGroupMembers(mctx)
	if err != nil {
		if err != group.ErrNestedGroupUnsupported && g.GroupType == iuser.SSOLIBGROUP {
			panic(err)
		}
	}

	groupMemberRoles := make([]GroupMemberRole, len(gMembers))
	for i, m := range gMembers {
		log.Debug("group:", i, m)
		groupMemberRoles[i] = GroupMemberRole{
			Name:     m.Name,
			FullName: m.FullName,
		}
		if m.Role == group.ADMIN {
			groupMemberRoles[i].Role = "admin"
		}
	}

	gwm := &GroupWithMembers{
		Name:         g.Name,
		FullName:     g.FullName,
		Members:      memberRoles,
		GroupMembers: groupMemberRoles,
	}
	return http.StatusOK, gwm
}

func (gr GroupResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:group", func(currentUser iuser.User) (int, interface{}) {
		groupname := params(ctx, "groupname")
		if groupname == "" {
			return http.StatusBadRequest, "groupname not given"
		}

		mctx := getModelContext(ctx)

		g, err := group.GetGroupIdByName(mctx, groupname)
		if err != nil {
			if err == group.ErrGroupNotFound {
				return http.StatusNotFound, "No such group"
			}
			panic(err)
		}

		if g.GroupType == iuser.SSOLIBGROUP {
			ok, role, err := g.GetMember(mctx, currentUser)
			if err != nil {
				panic(err)
			}

			if !(ok && role == group.ADMIN) {
				return http.StatusForbidden, "Not group admin"
			}

			if err = group.DeleteGroup(mctx, g); err != nil {
				panic(err)
			}

		} else if g.GroupType == iuser.BACKENDGROUP {
			ub := getUserBackend(ctx)
			if ubg, ok := ub.(iuser.BackendWithGroup); !ok {
				return http.StatusBadRequest, group.ErrBackendUnsupported
			} else {
				backendGroup, err := ubg.GetBackendGroup(g.Id)
				if err != nil {
					panic(err)
				}
				if err := ubg.DeleteBackendGroup(backendGroup); err != nil {
					return http.StatusBadRequest, err
				} else {
					if err = group.DeleteGroup(mctx, g); err != nil {
						panic(err)
					}
				}
			}
		} else {
			panic("here")
		}
		return http.StatusNoContent, "Group deleted"
	})
}

type MemberResource struct {
	server.BaseResource
}

func (mr MemberResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	t1 := time.Now()
	groupname := params(ctx, "groupname")
	if groupname == "" {
		return http.StatusBadRequest, "no groupname given"
	}
	username := params(ctx, "username")
	if username == "" {
		return http.StatusBadRequest, "no username given"
	}

	mctx := getModelContext(ctx)
	g, err := group.GetGroupIdByName(mctx, groupname)
	t2 := time.Now()
	log.Debug(t2.Sub(t1))
	switch {
	case err == group.ErrGroupNotFound:
		return http.StatusNotFound, "no such group"
	case err != nil:
		panic(err)
	}

	ub := getUserBackend(ctx)
	u, uerr := ub.GetUserByName(username)
	t3 := time.Now()
	log.Debug(t3.Sub(t2))
	if g.GroupType == iuser.SSOLIBGROUP {

		if uerr != nil {
			return http.StatusNotFound, "no such username"
		}

		ok, addingUserRole, err := g.GetMember(mctx, u)
		log.Debug(u, u.GetName())
		if ok {
			log.Debug("get role,", addingUserRole)
			retRole := "normal"
			if addingUserRole == group.ADMIN {
				retRole = "admin"
			}
			t4 := time.Now()
			log.Debug(t4.Sub(t3))
			return http.StatusOK, struct {
				Role string `json:"role"`
			}{
				retRole,
			}
		} else {
			if err != nil {
				panic(err)
			}
			return http.StatusNotFound, "no such member"
		}

	} else if g.GroupType == iuser.BACKENDGROUP {
		return http.StatusBadRequest, group.ErrBackendUnsupported
	} else {
		panic("Unexpected group type")
	}
	return http.StatusNoContent, nil
}

func (mr MemberResource) Put(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:group", func(currentUser iuser.User) (int, interface{}) {
		groupname := params(ctx, "groupname")
		if groupname == "" {
			return http.StatusBadRequest, "no groupname given"
		}
		username := params(ctx, "username")
		if username == "" {
			return http.StatusBadRequest, "no username given"
		}

		role := MemberRole{}
		if err := form.ParamBodyJson(r, &role); err != nil {
			return http.StatusBadRequest, err
		}

		var mrole group.MemberRole
		if role.Role == "admin" {
			mrole = group.ADMIN
		} else if role.Role == "normal" || role.Role == "" {
			mrole = group.NORMAL
		} else {
			return http.StatusBadRequest, "unknown role"
		}

		mctx := getModelContext(ctx)

		g, err := group.GetGroupIdByName(mctx, groupname)
		switch {
		case err == group.ErrGroupNotFound:
			return http.StatusNotFound, "no such group"
		case err != nil:
			panic(err)
		}

		ub := getUserBackend(ctx)
		u, uerr := ub.GetUserByName(username)

		if g.GroupType == iuser.SSOLIBGROUP {
			ok, currentUserRole, err := g.GetMember(mctx, currentUser)
			if err != nil {
				panic(err)
			}
			if !(ok && currentUserRole == group.ADMIN) {
				return http.StatusForbidden, "Not group admin"
			}

			if uerr != nil {
				return http.StatusNotFound, "no such username"
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
		} else if g.GroupType == iuser.BACKENDGROUP {
			if ubg, ok := ub.(iuser.BackendWithGroup); ok {
				backendGroup, err := ubg.GetBackendGroup(g.Id)
				if err != nil {
					panic(err)
				}
				err = backendGroup.AddUser(u)
				if err != nil {
					if err == iuser.ErrMethodNotSupported {
						return http.StatusBadRequest, err
					} else {
						panic(err)
					}
				}
			} else {
				panic(group.ErrBackendUnsupported)
			}
		} else {
			panic("Unexpected group type")
		}
		return http.StatusOK, "member added"
	})
}

func (mr MemberResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:group", func(currentUser iuser.User) (int, interface{}) {
		groupname := params(ctx, "groupname")
		if groupname == "" {
			return http.StatusBadRequest, "no groupname given"
		}
		username := params(ctx, "username")
		if username == "" {
			return http.StatusBadRequest, "no username given"
		}

		mctx := getModelContext(ctx)

		g, err := group.GetGroupIdByName(mctx, groupname)
		switch {
		case err == group.ErrGroupNotFound:
			return http.StatusNotFound, "no such group"
		case err != nil:
			panic(err)
		}

		ub := getUserBackend(ctx)
		u, uerr := ub.GetUserByName(username)

		if g.GroupType == iuser.SSOLIBGROUP {
			ok, currentUserRole, err := g.GetMember(mctx, currentUser)
			if err != nil {
				panic(err)
			}
			if !(ok && currentUserRole == group.ADMIN) {
				return http.StatusForbidden, "Not group admin"
			}

			if uerr != nil {
				return http.StatusNotFound, "no such username"
			}

			if err := g.RemoveMember(mctx, u); err != nil {
				panic(err)
			}
		} else if g.GroupType == iuser.BACKENDGROUP {
			if ubg, ok := ub.(iuser.BackendWithGroup); ok {
				backendGroup, err := ubg.GetBackendGroup(g.Id)
				if err != nil {
					panic(err)
				}
				err = backendGroup.RemoveUser(u)
				if err != nil {
					if err == iuser.ErrMethodNotSupported {
						return http.StatusBadRequest, err
					} else {
						panic(err)
					}
				}
			} else {
				panic(group.ErrBackendUnsupported)
			}
		} else {
			panic("Unexpected group type")
		}
		return http.StatusNoContent, "member deleted"
	})
}

type GroupMemberResource struct {
	server.BaseResource
}

func (gmr GroupMemberResource) Put(ctx context.Context, r *http.Request) (int, interface{}) {
	log.Debug("Put group member")
	return requireScope(ctx, "write:group", func(currentUser iuser.User) (int, interface{}) {
		groupname := params(ctx, "groupname")
		if groupname == "" {
			return http.StatusBadRequest, "no groupname given"
		}
		sonname := params(ctx, "sonname")
		if sonname == "" {
			return http.StatusBadRequest, "no sub group name given"
		}

		role := MemberRole{}
		if err := form.ParamBodyJson(r, &role); err != nil {
			return http.StatusBadRequest, err
		}

		var mrole group.MemberRole
		if role.Role == "admin" {
			mrole = group.ADMIN
		} else if role.Role == "normal" || role.Role == "" {
			mrole = group.NORMAL
		} else {
			return http.StatusBadRequest, "unknown role"
		}

		mctx := getModelContext(ctx)

		g, err := group.GetGroupIdByName(mctx, groupname)
		switch {
		case err == group.ErrGroupNotFound:
			return http.StatusNotFound, "no such group"
		case err != nil:
			panic(err)
		}

		gson, gerr := group.GetGroupIdByName(mctx, sonname)

		if g.GroupType == iuser.SSOLIBGROUP {
			ok, currentUserRole, err := g.GetMember(mctx, currentUser)
			if err != nil {
				panic(err)
			}
			if !(ok && currentUserRole == group.ADMIN) {
				return http.StatusForbidden, "Not group admin"
			}

			switch {
			case gerr == group.ErrGroupNotFound:
				return http.StatusNotFound, "no such sub group"
			case gerr != nil:
				panic(err)
			}

			addingGroupRole, err := group.GetGroupMemberRole(mctx, g.Id, gson.Id)
			log.Debug(addingGroupRole, err)
			if err == nil {
				if addingGroupRole != mrole {
					log.Debug("update group role,", addingGroupRole, mrole)
					if err := g.UpdateGroupMemberRole(mctx, gson, mrole); err != nil {
						panic(err)
					}
				}
			} else {
				if err != sql.ErrNoRows {
					panic(err)
				}
				if err := g.AddGroupMember(mctx, gson, mrole); err != nil {
					if err == group.ErrGroupIncludingFailed || err == group.ErrNestedGroupUnsupported {
						return http.StatusBadRequest, err
					}
					panic(err)
				}
			}
		} else if g.GroupType == iuser.BACKENDGROUP {
			return http.StatusBadRequest, "backend group is atomic"
		} else {
			panic("Unexpected group type")
		}
		return http.StatusOK, "group member added"
	})

}

func (gmr GroupMemberResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	log.Debug("Delete group member")
	return requireScope(ctx, "write:group", func(currentUser iuser.User) (int, interface{}) {
		groupname := params(ctx, "groupname")
		if groupname == "" {
			return http.StatusBadRequest, "no groupname given"
		}
		sonname := params(ctx, "sonname")
		if sonname == "" {
			return http.StatusBadRequest, "no sub group name given"
		}

		mctx := getModelContext(ctx)

		g, err := group.GetGroupIdByName(mctx, groupname)
		switch {
		case err == group.ErrGroupNotFound:
			return http.StatusNotFound, "no such group"
		case err != nil:
			panic(err)
		}

		gson, gerr := group.GetGroupIdByName(mctx, sonname)

		if g.GroupType == iuser.SSOLIBGROUP {
			ok, currentUserRole, err := g.GetMember(mctx, currentUser)
			if err != nil {
				panic(err)
			}
			if !(ok && currentUserRole == group.ADMIN) {
				return http.StatusForbidden, "Not group admin"
			}

			switch {
			case gerr == group.ErrGroupNotFound:
				return http.StatusNotFound, "no such sub group"
			case gerr != nil:
				panic(err)
			}

			if _, err := group.GetGroupMemberRole(mctx, g.Id, gson.Id); err == nil {
				if err := g.RemoveGroupMember(mctx, gson); err != nil {
					panic(err)
				}
			} else {
				if err != sql.ErrNoRows {
					panic(err)
				}
			}
		} else if g.GroupType == iuser.BACKENDGROUP {
			return http.StatusBadRequest, "backend group is atomic"
		} else {
			panic("Unexpected group type")
		}
		return http.StatusNoContent, "group member deleted"
	})
}
