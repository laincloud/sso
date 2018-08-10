package group

import (
	"container/list"
	"database/sql"
	"errors"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/jmoiron/sqlx"
)

const createGroupDAGTableSQL = `
CREATE TABLE IF NOT EXISTS groupdag (
	father_id INT NOT NULL,
	son_id INT NOT NULL,
	role TINYINT NOT NULL COMMENT '0: normal, 1: admin',
	PRIMARY KEY (father_id, son_id),
	KEY (father_id),
	KEY (son_id)
)`

// whether we use nested group
var valid bool

// 0: unlimited
// 1: the group could have users as sons only
var maxDepth int

var (
	ErrNestedGroupUnsupported     = errors.New("Nested Group is unsupported")
	ErrGroupIncludingFailed       = errors.New("Group cycle not allowed, or group depth exceeding the limit")
	ErrBackendGroupShouldBeAtomic = errors.New("Backend group is atomic")
)

// the depth of a group is defined as the longest path to the user node (leaf)
// i.e. depth = max of the depth of sons + 1
const createGroupDepthTableSQL = `
CREATE TABLE IF NOT EXISTS groupdepth (
	group_id INT NOT NULL,
	depth INT NOT NULL DEFAULT 0,
	PRIMARY KEY (group_id)
)`

func init() {
	valid = false
	maxDepth = 0 // 0 means the depth of the dag is unlimited
}

func EnableNestedGroup() {
	valid = true
	log.Info("Enable nested group")
}

func SetMaxDepth(depth int) {
	log.Info("Set Max Depth of the DAG equals ", depth)
	maxDepth = depth
}

func GroupCanBeNested() bool {
	return valid
}

type GroupMember struct {
	Group
	Role MemberRole // the role of this group which is a son for some group
}

type GroupDepth struct {
	GroupId int `db:"group_id"`
	Depth   int
}

func (g *Group) checkValidAndAtom() error {
	if !valid {
		return ErrNestedGroupUnsupported
	}
	if g.GroupType != iuser.SSOLIBGROUP {
		return ErrBackendGroupShouldBeAtomic
	}
	return nil
}

// should be sure that the son is not a direct son for the father yet
// for all methods modify the groupdepth or groupdag's primary key,
// we must protect it by lock
func (g *Group) AddGroupMember(ctx *models.Context, son *Group, role MemberRole) error {
	log.Debug("Try to add group member")
	err := g.checkValidAndAtom()
	if err != nil {
		log.Error(err)
		return err
	}

	lock := ctx.Lock
	lock.Lock()
	defer func() {
		lock.Unlock()
	}()
	log.Debug("before search")
	// first we test if we can add the group
	// second we add the group or return
	ok, depthBuffer, err := upDagIsOk(ctx, g.Id, son.Id)
	log.Debug(ok, depthBuffer, err)
	if ok {
		err = writeDepths(ctx, depthBuffer)
		if err == nil {
			err = g.addGroupMember(ctx, son, role)
			return err
		} else {
			log.Error(err)
			panic(err)
		}
	} else {
		return err
	}
	return nil
}

func (g *Group) AddGroupMemberWithoutLock(ctx *models.Context, son *Group, role MemberRole) error {
	log.Debug("Try to add group member")
	err := g.checkValidAndAtom()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug("before search")
	// first we test if we can add the group
	// second we add the group or return
	ok, depthBuffer, err := upDagIsOk(ctx, g.Id, son.Id)
	log.Debug(ok, depthBuffer, err)
	if ok {
		err = writeDepths(ctx, depthBuffer)
		if err == nil {
			err = g.addGroupMember(ctx, son, role)
			return err
		} else {
			log.Error(err)
			panic(err)
		}
	} else {
		return err
	}
	return nil
}

// son must be a son of this group
func (g *Group) UpdateGroupMemberRole(ctx *models.Context, son *Group, role MemberRole) error {
	err := g.checkValidAndAtom()
	if err != nil {
		return err
	}
	_, err = ctx.DB.Exec("UPDATE groupdag SET role=? WHERE father_id=? AND son_id=?", role, g.Id, son.Id)

	return err
}

func (g *Group) RemoveGroupMemberWithoutLock(ctx *models.Context, son *Group) error {
	sonDepth, _ := getGroupDepth(ctx, son.Id)
	fatherDepth, _ := getGroupDepth(ctx, g.Id)
	log.Debug(fatherDepth, sonDepth)
	if fatherDepth == sonDepth+1 {
		updateMap, err := upSearchAsSmallSon(ctx, g.Id, fatherDepth, son.Id, 0)

		err = writeDepths(ctx, updateMap)
		if err != nil {
			panic(err)
		}
	}
	return g.removeGroupMember(ctx, son)
}

// son must be a son of this group
func (g *Group) RemoveGroupMember(ctx *models.Context, son *Group) error {
	err := g.checkValidAndAtom()
	if err != nil {
		return err
	}
	if g.Id == son.Id {
		panic("can not remove oneself as group member")
	}
	lock := ctx.Lock
	lock.Lock()
	defer func() {
		lock.Unlock()
	}()
	return g.RemoveGroupMemberWithoutLock(ctx, son)
}

func (g *Group) removeGroupMember(ctx *models.Context, son *Group) error {
	_, err := ctx.DB.Exec("DELETE FROM groupdag WHERE father_id=? AND son_id=?",
		g.Id, son.Id)

	return err

}

func getExpectedFatherDepthForUpdatingSonDepth(ctx *models.Context, fatherId int, changedSonDepths map[int]int) (int, error) {
	father, err := GetGroup(ctx, fatherId)
	if err != nil {
		panic(err)
	}
	gMember, err := father.ListGroupMembers(ctx)
	if err != nil {
		panic(err)
	}
	if len(gMember) == 0 {
		return 1, nil
	}
	ret := -1
	for _, g := range gMember {
		var depth int
		var ok bool
		if depth, ok = changedSonDepths[g.Id]; !ok {
			depth, _ = getGroupDepth(ctx, g.Id)
		}
		if ret == -1 {
			ret = depth + 1
		} else if ret < depth+1 {
			ret = depth + 1
		}
	}
	return ret, nil
}

func upSearchAsSmallSon(ctx *models.Context, fatherId int, fatherDepth int, sonId int, sonNewDepthForFather int) (map[int]int, error) {
	retMap := make(map[int]int)
	changedDepth := make(map[int]int)
	changedDepth[sonId] = sonNewDepthForFather
	expectedFatherDepth, err := getExpectedFatherDepthForUpdatingSonDepth(ctx, fatherId, changedDepth)
	if err != nil {
		return nil, err
	}
	if expectedFatherDepth == fatherDepth {
		return retMap, nil
	} else if expectedFatherDepth > fatherDepth {
		panic("something error")
	}
	l0 := list.New()
	l1 := list.New()
	l0.PushBack(&GroupDepth{fatherId, expectedFatherDepth})
	tag := 0
	var l *list.List
	var theOther *list.List
	for {
		if l0.Len() == 0 && l1.Len() == 0 {
			break
		}
		if tag == 0 {
			l = l0
			theOther = l1
			tag = 1
		} else {
			l = l1
			theOther = l0
			tag = 0
		}
		for iter := l.Front(); iter != nil; iter = iter.Next() {
			v := iter.Value.(*GroupDepth)
			if oldDepth, ok := retMap[v.GroupId]; ok {
				if oldDepth >= v.Depth {
					retMap[v.GroupId] = v.Depth
				} else {
					panic("strange")
				}
			} else {
				retMap[v.GroupId] = v.Depth
			}
		}
		for l.Len() > 0 {
			iter := l.Front()
			v := iter.Value.(*GroupDepth)
			l.Remove(iter)
			fathers, err := ListGroupFathersById(ctx, v.GroupId)
			if err != nil {
				panic(err)
			}
			for _, g := range fathers {
				dep, err := getExpectedFatherDepthForUpdatingSonDepth(ctx, g.Id, retMap)
				if err != nil {
					panic(err)
				}
				oldDep, _ := getGroupDepth(ctx, g.Id)
				if dep >= oldDep {
					continue
				}
				find := false
				for iter := theOther.Front(); iter != nil; iter = iter.Next() {
					gd := iter.Value.(*GroupDepth)
					if gd.GroupId == g.Id {
						find = true
						if gd.Depth < dep {
							gd.Depth = dep
						}
					}
				}
				if !find {
					theOther.PushBack(&GroupDepth{g.Id, dep})
				}
			}
		}
	}
	return retMap, nil
}

func (g *Group) ListGroupMembers(ctx *models.Context) ([]GroupMember, error) {
	gMembers := []GroupMember{}
	err := g.checkValidAndAtom()
	if err != nil {
		return gMembers, err
	}
	rows, err := ctx.DB.Query("SELECT son_id, role FROM groupdag WHERE father_id=?", g.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return gMembers, nil
		}
		return nil, err
	}

	for rows.Next() {
		var sonId int
		var role MemberRole
		if err = rows.Scan(&sonId, &role); err != nil {
			return nil, err
		}
		sonGroup, err := GetGroup(ctx, sonId)
		if err != nil {
			return nil, err
		}
		gMembers = append(gMembers, GroupMember{*sonGroup, role})
	}
	return gMembers, nil
}

func ListGroupFathersById(ctx *models.Context, gId int) ([]Group, error) {
	groups := []Group{}
	rows, err := ctx.DB.Query("SELECT father_id FROM groupdag WHERE son_id=?", gId)
	if err != nil {
		if err == sql.ErrNoRows {
			return groups, nil
		}
		return nil, err
	}
	for rows.Next() {
		var fatherId int
		if err = rows.Scan(&fatherId); err != nil {
			return nil, err
		}
		fatherGroup, err := GetGroup(ctx, fatherId)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *fatherGroup)
	}
	return groups, nil
}

func upDagIsOk(ctx *models.Context, fatherId int, sonId int) (bool, map[int]int, error) {
	if fatherId == sonId {
		return false, nil, ErrGroupIncludingFailed
	}
	sonDepth, _ := getGroupDepth(ctx, sonId)
	fatherDepth, _ := getGroupDepth(ctx, fatherId)
	if sonDepth < fatherDepth {
		return true, nil, nil
	} else {
		ok, depthBuffer, err := upSearch(ctx, fatherId, sonId, (sonDepth + 1))
		return ok, depthBuffer, err
	}
}

// since depth of some son node becomes bigger
func upSearch(ctx *models.Context, fatherId, sonId, fatherDepth int) (bool, map[int]int, error) {
	buffer := make(map[int]int)
	l0 := list.New()
	l1 := list.New()
	l0.PushBack(&GroupDepth{fatherId, fatherDepth})
	tag := 0
	var l *list.List
	var another *list.List
	for {
		if l0.Len() == 0 && l1.Len() == 0 {
			break
		}
		if tag == 0 {
			l = l0
			another = l1
			tag = 1
		} else {
			l = l1
			another = l0
			tag = 0
		}
		for l.Len() > 0 {
			iter := l.Front()
			v := iter.Value.(*GroupDepth)
			l.Remove(iter)
			if maxDepth != 0 && v.Depth > maxDepth {
				return false, nil, ErrGroupIncludingFailed
			}
			if v.GroupId == sonId {
				return false, nil, ErrGroupIncludingFailed
			}
			if oldDepth, ok := buffer[v.GroupId]; ok {
				if oldDepth <= v.Depth {
					buffer[v.GroupId] = v.Depth
				} else {
					log.Error("strange")
				}
			} else {
				buffer[v.GroupId] = v.Depth
			}
			// 找到 v 的父亲，如果深度不超过 v，则将其预测的深度值加入 another 链表
			// 加入前需判断 another 是否已经有该父亲
			fathers, err := ListGroupFathersById(ctx, v.GroupId)
			if err != nil {
				panic(err)
			}
			for _, g := range fathers {
				dep, _ := getGroupDepth(ctx, g.Id)
				if dep > v.Depth {
					continue
				}
				newDepth := v.Depth + 1
				find := false
				for iter := another.Front(); iter != nil; iter = iter.Next() {
					gd := iter.Value.(*GroupDepth)
					if gd.GroupId == g.Id {
						find = true
						if gd.Depth < newDepth {
							gd.Depth = newDepth
						}
					}
				}
				if !find {
					another.PushBack(&GroupDepth{g.Id, newDepth})
				}
			}
		}
	}
	return true, buffer, nil
}

// key is group id, value is depth, group must exist
func writeDepths(ctx *models.Context, depths map[int]int) error {
	tx := ctx.DB.MustBegin()
	var err error
	for k, v := range depths {
		_, err = tx.Exec("UPDATE groupdepth SET depth=? WHERE group_id=?", v, k)
		if err != nil {
			tx.Rollback()
			break
		}
	}
	err = tx.Commit()
	return err
}

func deleteGroupDepth(ctx *models.Context, group *Group) error {
	_, err := ctx.DB.Exec("DELETE FROM groupdepth WHERE group_id=?", group.Id)

	return err
}

func (g *Group) addGroupMember(ctx *models.Context, son *Group, role MemberRole) error {
	log.Debug("add group member")
	_, err := ctx.DB.Exec(
		"INSERT INTO groupdag (father_id, son_id, role) VALUES (?, ?, ?)",
		g.Id, son.Id, role)

	return err
}

func getGroupDepth(ctx *models.Context, groupId int) (int, error) {
	depth := GroupDepth{}
	err := ctx.DB.Get(&depth, "SELECT * FROM groupdepth WHERE group_id=?", groupId)
	if err != nil {
		// 为了兼容老组
		if err == sql.ErrNoRows {
			log.Info("for old group, write depth as 1")
			err = writeGroupDepth(ctx, groupId, 1)
			if err != nil {
				panic(err)
			}
			depth.Depth = 1
		} else {
			panic(err)
		}
	}
	return depth.Depth, nil
}

func writeGroupDepth(ctx *models.Context, groupId int, depth int) error {
	_, err := ctx.DB.Exec("INSERT INTO groupdepth (group_id, depth) VALUES (?, ?)", groupId, depth)

	return err
}

func GetGroupMemberRole(ctx *models.Context, fatherId int, sonId int) (role MemberRole, err error) {
	err = ctx.DB.QueryRow("SELECT role FROM groupdag WHERE father_id=? AND son_id=?",
		fatherId, sonId).Scan(&role)
	return
}

func ListFathersOfGroups(ctx *models.Context, sonIds []int) ([]int, error) {
	fathers := []int{}
	query, args, err := sqlx.In("SELECT father_id FROM groupdag WHERE son_id IN(?)", sonIds)
	if err != nil {
		return nil, err
	}
	query = ctx.DB.Rebind(query)
	err = ctx.DB.Select(&fathers, query, args...)
	if err != nil {
		log.Debug(err)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return fathers, nil
}

func ListAdminFathersOfGroups(ctx *models.Context, sonIds []int) ([]int, error) {
	fathers := []int{}
	query, args, err := sqlx.In("SELECT father_id FROM groupdag WHERE son_id IN(?) AND role=?", sonIds, ADMIN)
	if err != nil {
		return nil, err
	}
	query = ctx.DB.Rebind(query)
	err = ctx.DB.Select(&fathers, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return fathers, nil
}

func GetAdminIdsOfGroup(ctx *models.Context, gId int)([]int, error) {
	var admins []int
	err := ctx.DB.Select(&admins, "SELECT user_id FROM user_group WHERE group_id=? AND role=?", gId, ADMIN)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return admins, nil
}

