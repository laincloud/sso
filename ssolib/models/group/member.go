package group

import (
	"database/sql"
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/pkg/errors"
)

const MAXREQUEST int = 100

type MemberRole int8

const (
	NORMAL MemberRole = iota
	ADMIN
)

type Member struct {
	iuser.User
	Role MemberRole
}

type GroupRole struct {
	Group
	Role MemberRole // some user's role in the group
}

const createUserGroupTableSQL = `
CREATE TABLE IF NOT EXISTS user_group (
	user_id INT NOT NULL,
	group_id INT NOT NULL,
	role TINYINT NOT NULL COMMENT '0: normal, 1: admin',
	PRIMARY KEY (user_id, group_id),
	KEY (group_id)
)`

func (g *Group) AddMember(ctx *models.Context, user iuser.User, role MemberRole) error {
	log.Debug("begin add:", user.GetId(), g.Id, role)
	_, err := ctx.DB.Exec(
		"INSERT INTO user_group (user_id, group_id, role) VALUES (?, ?, ?)",
		user.GetId(), g.Id, role)

	return err
}

func (g *Group) UpdateMember(ctx *models.Context, user iuser.User, role MemberRole) error {
	_, err := ctx.DB.Exec(
		"UPDATE user_group SET role=? WHERE user_id=? AND group_id=?",
		role, user.GetId(), g.Id)

	return err
}

func (g *Group) RemoveMember(ctx *models.Context, user iuser.User) error {
	_, err := ctx.DB.Exec("DELETE FROM user_group WHERE user_id=? AND group_id=?",
		user.GetId(), g.Id)

	return err
}

// not recursive for nested groups
func (g *Group) ListMembers(ctx *models.Context) ([]Member, error) {
	back := ctx.Back
	members := []Member{}
	if g.GroupType == iuser.SSOLIBGROUP {
		rows, err := ctx.DB.Query(
			"SELECT user_id, role FROM user_group WHERE group_id=?",
			g.Id)
		if err != nil {
			if err == sql.ErrNoRows {
				return members, nil
			}
			return nil, err
		}

		for rows.Next() {
			var userId int
			var role MemberRole
			if err = rows.Scan(&userId, &role); err != nil {
				return nil, err
			}
			user, err := back.GetUser(userId)
			if err != nil {
				return nil, err
			}

			members = append(members, Member{user, role})
		}

		return members, nil
	} else if g.GroupType == iuser.BACKENDGROUP {
		var ubg iuser.BackendWithGroup
		var ok bool
		if ubg, ok = back.(iuser.BackendWithGroup); !ok {
			return nil, ErrBackendUnsupported
		}
		bg, err := ubg.GetBackendGroupByName(g.Name)
		if err != nil {
			panic(err)
		}
		users, err := bg.ListUsers()
		if err != nil {
			if err == iuser.ErrMethodNotSupported {
				return members, nil
			} else {
				panic(err)
			}
		}
		members = backendUsersToMembers(users)
		return members, nil
	} else {
		panic("here")
	}
	return nil, nil
}

func (g *Group) GetGroupMembersID(ctx *models.Context) ([]int, error) {
	members := []int{}
	err := ctx.DB.Select(&members, "SELECT user_id FROM user_group WHERE group_id=?", g.Id)
	if err != nil {
		return nil, err
	}
	return members, err
}

// Return (true, role, nil) if u is member of g, otherwise return (false, 0, nil).
// error will be non-nil if anything unexpected happens.
// Must considering recursive if valid
func (g *Group) GetMember(ctx *models.Context, u iuser.User) (ok bool, role MemberRole, err error) {
	/*err = ctx.DB.QueryRow("SELECT role FROM user_group WHERE user_id=? AND group_id=?",
		u.GetId(), g.Id).Scan(&role)
	switch {
	case err == sql.ErrNoRows:
		err = nil
		return
	case err != nil:
		return
	default:
		ok = true
		return
	}*/
	groups, err := getGroupRolesRecursivelyOfUser(ctx, u, false)
	if err != nil {
		return false, 0, err
	}
	if role, ok = groups[g.Id]; ok {
		return true, role, nil
	} else {
		return false, 0, nil
	}
}

// must not recursive
func (g *Group) removeAllMembers(ctx *models.Context) error {
	_, err := ctx.DB.Exec("DELETE FROM user_group WHERE group_id=?", g.Id)

	return err
}

// direct groups of the users, for now it is used only in "管理员特供"
func GetGroupsOfUserByIds(ctx *models.Context, userIds []int) (map[int][]Group, error) {
	groupMap := make(map[int][]Group)
	if len(userIds) == 0 {
		return groupMap, nil
	}

	userGroupIds := make(map[int][]int)
	groupIds := []int{}
	if query, args, err := sqlx.In("SELECT group_id, user_id FROM user_group WHERE user_id IN(?)", userIds); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return groupMap, err
	} else {
		if rows, err := ctx.DB.Query(query, args...); err != nil {
			return groupMap, err
		} else {
			for rows.Next() {
				var (
					userId  int
					groupId int
				)
				if err := rows.Scan(&groupId, &userId); err != nil {
					return groupMap, err
				}
				groupIds = append(groupIds, groupId)
				userGroupIds[userId] = append(userGroupIds[userId], groupId)
			}
		}
	}

	groups, err := ListGroups(ctx, groupIds...)
	if err != nil {
		return groupMap, err
	}
	groupSet := make(map[int]Group)
	for _, g := range groups {
		groupSet[g.Id] = g
	}
	for userId, groups := range userGroupIds {
		for _, g := range groups {
			if group, ok := groupSet[g]; ok {
				groupMap[userId] = append(groupMap[userId], group)
			}
		}
	}
	return groupMap, nil
}

// should be recursive, but may be slow if we find recursive roles
func GetGroupRolesOfUser(ctx *models.Context, user iuser.User) ([]GroupRole, error) {
	roles := []GroupRole{}
	mapRoles, err := getGroupRolesRecursivelyOfUser(ctx, user, false)
	if err != nil {
		log.Error("strange:", user)
		panic(err)
	}
	sortedGroupIds := []int{}
	for k, _ := range mapRoles {
		sortedGroupIds = append(sortedGroupIds, k)
	}
	sort.Ints(sortedGroupIds)
	for _, k := range sortedGroupIds {
		g, err := GetGroup(ctx, k)
		v := mapRoles[k]
		if err != nil {
			panic(err)
		}
		tmp := GroupRole{
			Group: *g,
			Role:  v,
		}
		roles = append(roles, tmp)
	}
	return roles, nil
}

func GetGroupRolesDirectlyOfUser(ctx *models.Context, user iuser.User) ([]GroupRole, error) {
	groupRoles, err := getSSOLIBGroupRolesDirectlyOfUser(ctx, user)
	ub := ctx.Back
	var bGroups []iuser.BackendGroup
	if ubg, ok := ub.(iuser.BackendWithGroup); ok {
		bGroups, err = ubg.GetBackendGroupsOfUser(user)
		if err != nil {
			panic(err)
		}
	}
	ret := make([]GroupRole, (len(groupRoles) + len(bGroups)))
	copy(ret, groupRoles)
	for i, g := range bGroups {
		role := NORMAL
		ret[(i + len(groupRoles))] = GroupRole{
			Group: Group{
				Id:       g.GetId(),
				Name:     g.GetName(),
				FullName: g.GetRules().(string),
			},
			Role: role,
		}
	}
	return ret, nil
}

func getSSOLIBRoleDirectlyOfUser(ctx *models.Context, user iuser.User) (map[int]MemberRole, error) {
	grMap := make(map[int]MemberRole)
	if rows, err := ctx.DB.Query("SELECT group_id, role FROM user_group WHERE user_id=?", user.GetId()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	} else {
		for rows.Next() {
			var groupId, role int
			if err := rows.Scan(&groupId, &role); err == nil {
				grMap[groupId] = MemberRole(role)
			}
		}
	}
	return grMap, nil
}

func getSSOLIBGroupRolesDirectlyOfUser(ctx *models.Context, user iuser.User) ([]GroupRole, error) {
	groupIds := []int{}
	grMap := make(map[int]MemberRole)
	if rows, err := ctx.DB.Query("SELECT group_id, role FROM user_group WHERE user_id=?", user.GetId()); err != nil {
		if err == sql.ErrNoRows {
			return []GroupRole{}, nil
		}
		return nil, err
	} else {
		for rows.Next() {
			var groupId, role int
			if err := rows.Scan(&groupId, &role); err == nil {
				groupIds = append(groupIds, groupId)
				grMap[groupId] = MemberRole(role)
			}
		}
	}

	groupRoles := make([]GroupRole, 0, len(groupIds))
	if len(groupIds) == 0 {
		return groupRoles, nil
	}
	if groups, err := ListGroups(ctx, groupIds...); err != nil {
		return nil, err
	} else {
		for _, g := range groups {
			gr := GroupRole{
				Group: g,
				Role:  grMap[g.Id],
			}
			groupRoles = append(groupRoles, gr)
		}
	}
	return groupRoles, nil
}

// must be recursive to find out all the groups
func GetGroupsOfUser(ctx *models.Context, user iuser.User) ([]Group, error) {
	groups := []Group{}
	mapGroups, err := getGroupsRecursivelyOfUser(ctx, user)
	if err != nil {
		panic(err)
	}
	for k, _ := range mapGroups {
		g, err := GetGroup(ctx, k)
		if err != nil {
			panic(err)
		}
		groups = append(groups, *g)
	}
	return groups, nil
}

func getGroupsDirectlyOfUser(ctx *models.Context, user iuser.User) ([]Group, error) {
	groupIds := []int{}
	err := ctx.DB.Select(&groupIds,
		"SELECT group_id FROM user_group WHERE user_id=?",
		user.GetId())
	groups := []Group{}
	if err != nil {
		if err == sql.ErrNoRows {
			return groups, nil
		}
		return nil, err
	}

	for _, gid := range groupIds {
		g, err := GetGroup(ctx, gid)
		if err != nil {
			if err != ErrGroupNotFound {
				return nil, err
			} else {
				log.Warnf("User %s in a non-exist group %d?", user.GetName(), gid)
			}
		} else {
			groups = append(groups, *g)
		}
	}

	ub := ctx.Back
	var bGroups []iuser.BackendGroup
	if ubg, ok := ub.(iuser.BackendWithGroup); ok {
		bGroups, err = ubg.GetBackendGroupsOfUser(user)
		if err != nil {
			panic(err)
		}
	}
	for _, g := range bGroups {
		groups = append(groups, Group{
			Id:       g.GetId(),
			Name:     g.GetName(),
			FullName: g.GetRules().(string),
		})
	}

	return groups, nil
}

func RemoveUserFromAllGroups(ctx *models.Context, user iuser.User) error {
	_, err := ctx.DB.Exec("DELETE FROM user_group WHERE user_id=?", user.GetId())

	return err
}

func backendUsersToMembers(users []iuser.User) []Member {
	members := make([]Member, len(users))
	for i, u := range users {
		members[i] = Member{
			User: u,
			Role: NORMAL,
		}
	}
	return members
}

func getGroupRolesRecursivelyOfUser(ctx *models.Context, user iuser.User, adminOnly bool) (map[int]MemberRole, error) {
	if adminOnly {
		return getAdminGroupsRecursivelyOfUser(ctx, user)
	} else {
		admins, err := getAdminGroupsRecursivelyOfUser(ctx, user)
		if err != nil {
			panic(err)
		}
		ret, err := getGroupsRecursivelyOfUser(ctx, user)
		if err != nil {
			panic(err)
		}
		for k, v := range admins {
			ret[k] = v
		}
		return ret, nil
	}
}

func getAdminGroupsRecursivelyOfUser(ctx *models.Context, user iuser.User) (map[int]MemberRole, error) {
	preQueue := make([]int, 0, MAXREQUEST)
	ret := make(map[int]MemberRole)
	groupRoles, err := getSSOLIBRoleDirectlyOfUser(ctx, user)
	if err != nil {
		return nil, err
	}
	for gId, role := range groupRoles {
		if role == ADMIN {
			ret[gId] = ADMIN
			preQueue = append(preQueue, gId)
		}
	}
	for true {
		if len(preQueue) == 0 {
			break
		}
		rawFathers := make([]int, 0, MAXREQUEST)
		//when 1 < len <= MAXREQUEST, loop one time
		for i := 0; ; i++ {
			offset := i * MAXREQUEST
			remain := len(preQueue) - offset
			if remain <= MAXREQUEST {
				queue := make([]int, remain, MAXREQUEST)
				copy(queue, preQueue[offset:offset + remain])
				partFathers, err := ListAdminFathersOfGroups(ctx, queue)
				if err != nil {
					return nil, err
				}
				rawFathers = append(rawFathers,partFathers...)
				break
			}
			queue := make([]int, MAXREQUEST, MAXREQUEST)
			copy(queue, preQueue[offset:offset + MAXREQUEST])
			partFathers, err := ListAdminFathersOfGroups(ctx, queue)
			if err != nil {
				return nil, err
			}
			rawFathers = append(rawFathers,partFathers...)
		}
		//deduplicate
		fathers := make([]int, 0, MAXREQUEST)
		for _,f := range rawFathers {
			if _, ok := ret[f]; !ok {
				ret[f] = ADMIN
				fathers = append(fathers, f)
			}
		}
		preQueue = fathers
	}
	return ret, nil
}

func getGroupsRecursivelyOfUser(ctx *models.Context, user iuser.User) (map[int]MemberRole, error) {
	preQueue := make([]int, 0, MAXREQUEST)
	ret := make(map[int]MemberRole)
	groups, err := getGroupsDirectlyOfUser(ctx, user)
	if err != nil {
		return nil, err
	}
	for _, v := range groups {
		ret[v.Id] = NORMAL
		preQueue = append(preQueue, v.Id)
	}
	for true {
		if len(preQueue) == 0 {
			break
		}
		rawFathers := make([]int, 0, MAXREQUEST)
		//when 1 < len <= MAXREQUEST, loop one time
		for i := 0; ; i++ {
			offset := i * MAXREQUEST
			remain := len(preQueue) - offset
			if remain <= MAXREQUEST {
				queue := make([]int, remain, MAXREQUEST)
				copy(queue, preQueue[offset:offset + remain])
				partFathers, err := ListFathersOfGroups(ctx, queue)
				if err != nil {
					return nil, err
				}
				rawFathers = append(rawFathers,partFathers...)
				break
			}
			queue := make([]int, MAXREQUEST, MAXREQUEST)
			copy(queue, preQueue[offset:offset + MAXREQUEST])
			partFathers, err := ListFathersOfGroups(ctx, queue)
			if err != nil {
				return nil, err
			}
			rawFathers = append(rawFathers,partFathers...)
		}
		//deduplicate
		fathers := make([]int, 0, MAXREQUEST)
		for _,f := range rawFathers {
			if _, ok := ret[f]; !ok {
				ret[f] = NORMAL
				fathers = append(fathers, f)
			}
		}
		preQueue = fathers
	}
	return ret, nil
}

func GetRoleOfUser(mctx *models.Context, uId int, gId int) (MemberRole, error){
	Roles := []MemberRole{}
	err := mctx.DB.Select(&Roles, "SELECT role FROM user_group WHERE user_id =? AND group_id=?",
		uId, gId)
	if err != nil {
		return NORMAL, err
	}
	if len(Roles) != 1 {
		return NORMAL, errors.New("wrong number of roles")
	}
	return Roles[0], nil
}
