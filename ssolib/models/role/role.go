package role

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

var (
	ErrRoleNotFound            = errors.New("role not found")
	ErrLeafRoleHasNoSubRole    = errors.New("leaf role has no subrole")
	ErrRootRoleHasNoParent     = errors.New("can't set root role's parent")
	ErrRootRoleCannotBeDeleted = errors.New("can't delete root role")
)

const createRoleTableSQL = `
CREATE TABLE IF NOT EXISTS role (
	id INT NOT NULL,
	name VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	fullname VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	super_id INT NOT NULL DEFAULT -1,
	app_id INT NOT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY (app_id, name)
) DEFAULT CHARSET=latin1`

const createRoleResourceTableSQL = `
CREATE TABLE IF NOT EXISTS role_resource (
	role_id INT NOT NULL,
	resource_id INT NOT NULL,
	PRIMARY KEY (role_id, resource_id),
	KEY (role_id)
) DEFAULT CHARSET=latin1`

type Role struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"description"`
	SuperRoleId int    `db:"super_id" json:"parent_id"`
	AppId       int    `db:"app_id" json:"app_id"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

type RoleMembers struct {
	Role
	Type    group.MemberRole `json:"type"`
	Members []group.Member   `json:"members"`
}

func InitDatabase(ctx *models.Context) {
	ctx.DB.MustExec(createRoleTableSQL)
	ctx.DB.MustExec(createResourceTableSQL)
	ctx.DB.MustExec(createRoleResourceTableSQL)
}

func CreateRoleWithGroup(ctx *models.Context, roleName string, fullName string, appId int, groupId int) (*Role, error) {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec(
		"INSERT INTO role (id, name, fullname, app_id) VALUES (?, ?, ?, ?)",
		groupId, roleName, fullName, appId)
	if err2 := tx.Commit(); err2 != nil {
		log.Debug(err2)
		return nil, err2
	}
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	return GetRole(ctx, groupId)
}

func CreateRoleWithoutGroup(ctx *models.Context, roleName string, fullName string, appId int, superId int) (*Role, error) {
	// first confirm the super role existed
	// then create a group using a default name ".app${app_id}role$TIMESTAMP-rand(99)"
	// set the current superId group as admin sub group
	// then create a role use the roleName and set the parent

	_, err := GetRole(ctx, superId)
	if err != nil {
		if err != ErrRoleNotFound {
			log.Error(err)
		}
		return nil, err
	}
	if ok := IsLeafRole(ctx, superId); ok {
		return nil, ErrLeafRoleHasNoSubRole
	}
	groupName := getGroupName(appId)
	g, err := group.CreateGroup(ctx, &group.Group{
		Name:      groupName,
		FullName:  roleName,
		GroupType: iuser.SSOLIBGROUP,
	})
	if err != nil {
		log.Error(err)
		//TODO maybe need rollback
		return nil, err
	}
	sonGroup, _ := group.GetGroup(ctx, superId)
	g.AddGroupMember(ctx, sonGroup, group.ADMIN)

	r, err := CreateRoleWithGroup(ctx, roleName, fullName, appId, g.Id)
	if err != nil {
		//TODO rollback
		return nil, err
	}

	tx := ctx.DB.MustBegin()
	_, err1 := tx.Exec("UPDATE role SET super_id=? WHERE id=?", superId, r.Id)
	err2 := tx.Commit()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	return GetRole(ctx, r.Id)

}

func getGroupName(appId int) string {
	timestamp := time.Now().Unix()
	rand.NewSource(timestamp)
	r := rand.Intn(99)
	return ".app" + strconv.Itoa(appId) + "role" + fmt.Sprint(timestamp) + "-" + strconv.Itoa(r)
}

func GetRolesByGroupIds(ctx *models.Context, groupIds []int) ([]Role, error) {
	query, args, err := sqlx.In("SELECT  * FROM role WHERE id IN(?)", groupIds)
	if err != nil {
		return nil, err
	}
	roles := []Role{}
	err = ctx.DB.Select(&roles, query, args...)
	return roles, err
}

func GetRolesByAppId(ctx *models.Context, appId int) ([]Role, error) {
	roles := []Role{}
	err := ctx.DB.Select(&roles, "SELECT * FROM role WHERE app_id=?", appId)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return roles, err
}

func GetRoles(ctx *models.Context, username string, appId int) ([]Role, error) {
	return nil, nil
}

func GetRole(ctx *models.Context, id int) (*Role, error) {
	role := Role{}
	err := ctx.DB.Get(&role, "SELECT * FROM role WHERE id=?", id)
	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFound
	} else if err != nil {
		return nil, err
	}
	return &role, nil
}

func GetRoleIdByName(ctx *models.Context, Name string, appId int) (int, error) {
	roleIds := []int{}
	err := ctx.DB.Get(&roleIds, "SELECT id FROM role WHERE name=? AND app_id=?", Name, appId)
	if err == sql.ErrNoRows {
		return -1, nil
	} else if err != nil {
		return -1, err
	}
	if len(roleIds) != 1 {
		return -1, nil
	}
	return roleIds[0], nil
}

func UpdateRole(ctx *models.Context, id int, name string, fullname string, parent int) (*Role, error) {
	// if parent id is changed:
	// 1. if the new parent is the group's offspring
	// 2. if the new parent has resources, i.e. is a leaf role
	// 3. the role's group delete the old parent role's group
	// 4. the role's group add the new parent role's group as group member
	// 5. sub role set new parent id
	role, err := GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if role.SuperRoleId == parent {
		tx := ctx.DB.MustBegin()
		_, err1 := tx.Exec(
			"UPDATE role SET name=?, fullname=? WHERE id=?",
			name, fullname, id)
		err2 := tx.Commit()
		if err1 != nil {
			return nil, err1
		}
		if err2 != nil {
			return nil, err2
		}
		return GetRole(ctx, id)
	} else if role.SuperRoleId == -1 {
		return nil, ErrRootRoleHasNoParent
	} else {
		if ok := IsLeafRole(ctx, parent); ok {
			return nil, ErrLeafRoleHasNoSubRole
		} else if ok, err := IsOffspring(ctx, id, parent); err == nil {
			if ok {
				return nil, errors.New("can't set offspring as the parent")
			} else {
				ctx.Lock.Lock()
				defer ctx.Lock.Unlock()
				g, err := group.GetGroup(ctx, id)
				if err != nil {
					log.Error(err)
					return nil, err
				}
				sonG, err := group.GetGroup(ctx, role.SuperRoleId)
				if err != nil {
					log.Error(err)
					return nil, err
				}
				newSonG, err := group.GetGroup(ctx, parent)
				if err != nil {
					log.Error(err)
					return nil, err
				}

				// TODO check error and rollback
				g.RemoveGroupMemberWithoutLock(ctx, sonG)
				g.AddGroupMemberWithoutLock(ctx, newSonG, group.ADMIN)

				tx := ctx.DB.MustBegin()
				_, err1 := tx.Exec(
					"UPDATE role SET name=?, fullname=?, super_id=? WHERE id=?",
					name, fullname, parent, id)
				err2 := tx.Commit()
				if err1 != nil {
					return nil, err1
				}
				if err2 != nil {
					return nil, err2
				}
				return GetRole(ctx, id)
			}
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func IsOffspring(ctx *models.Context, roleId int, queryId int) (bool, error) {
	nowId := queryId
	for {
		rows, err := ctx.DB.Query("SELECT super_id FROM role WHERE id=?", nowId)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Error(err)
				return true, err
			}
		}
		for rows.Next() {
			var fatherId int
			if err = rows.Scan(&fatherId); err != nil {
				return true, err
			}
			if fatherId == roleId {
				return true, nil
			} else if fatherId == -1 {
				return false, nil
			} else {
				nowId = fatherId
			}
		}
	}
}

func DeleteRole(ctx *models.Context, id int) error {
	// 1. get role check the role already existed
	// 2. check whether the role is root, if yes, return forbidden
	// 3. check whether the role is leaf, if yes, check the role's resources, delete the relationship
	// 4. delete the related group
	// 5. sub role's related groups add its parent role's group as group member
	// 6. change sub role's parent as its parent
	// 7. delete the role
	// TODO check fail and rollback
	role, err := GetRole(ctx, id)
	if err != nil {
		return err
	}
	superId := role.SuperRoleId
	if superId == -1 {
		return ErrRootRoleCannotBeDeleted
	}
	sGroup, err := group.GetGroup(ctx, superId)
	if err != nil {
		panic(err)
	}
	tx := ctx.DB.MustBegin()
	if IsLeafRole(ctx, id) {
		_, err = tx.Exec("DELETE FROM role_resource WHERE role_id=?", id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if err2 := tx.Commit(); err2 != nil {
		return err2
	}

	subRoles, err := getSubRoles(ctx, id)
	if err == ErrLeafRoleHasNoSubRole {
		subRoles = []*Role{}
	} else if err != nil {
		return err
	}

	g, err := group.GetGroup(ctx, id)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	err = group.DeleteGroup(ctx, g)
	if err != nil {
		panic(err)
	}

	for _, sRole := range subRoles {
		tmpGroup, err := group.GetGroup(ctx, sRole.Id)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		err = tmpGroup.AddGroupMember(ctx, sGroup, group.ADMIN)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		tx = ctx.DB.MustBegin()
		_, err1 := tx.Exec(
			"UPDATE role SET super_id=? WHERE id=?",
			superId, sRole.Id)
		err2 := tx.Commit()
		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}
	}

	tx = ctx.DB.MustBegin()
	_, err1 := tx.Exec(
		"DELETE FROM role WHERE id=?", id)
	err2 := tx.Commit()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func getSuperRole(ctx *models.Context, roleId int) (*Role, error) {
	return nil, nil
}

func getSubRoles(ctx *models.Context, roleId int) ([]*Role, error) {
	groups, err := group.ListGroupFathersById(ctx, roleId)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, ErrLeafRoleHasNoSubRole
	}
	gIds := []int{}
	for _, g := range groups {
		gIds = append(gIds, g.Id)
	}
	roles, err := GetRolesByGroupIds(ctx, gIds)
	if err != nil {
		return nil, err
	}
	ret := make([]*Role, len(roles))
	for i := 0; i < len(roles); i++ {
		ret[i] = &roles[i]
	}
	return ret, nil
}

func GetSubTreeRoles(ctx *models.Context, roleId int) ([]*Role, error) {
	return nil, nil
}

func GetDirectRoles(ctx *models.Context, user iuser.User, appId int) ([]*Role, error) {
	groupRoles, err := group.GetGroupRolesDirectlyOfUser(ctx, user)
	if err != nil {
		return nil, err
	}
	roles := []*Role{}
	for _, g := range groupRoles {
		r, err := GetRole(ctx, g.Id)
		if err == ErrRoleNotFound {
			continue
		} else if err != nil {
			return nil, err
		} else {
			if r.AppId == appId {
				roles = append(roles, r)
			}
		}
	}
	return roles, nil
}

func GetAllRoleMembers(ctx *models.Context, user iuser.User, appId int) ([]*RoleMembers, error) {
	return getRoleMembersForApp(ctx, user, appId, false)
}

func GetDirectRoleMembers(ctx *models.Context, user iuser.User, appId int) ([]*RoleMembers, error) {
	return getRoleMembersForApp(ctx, user, appId, true)
}

func getRoleMembersForApp(ctx *models.Context, user iuser.User, appId int, isDirect bool) ([]*RoleMembers, error) {
	var err error
	var groupRoles []group.GroupRole
	if isDirect {
		groupRoles, err = group.GetGroupRolesDirectlyOfUser(ctx, user)
	} else {
		groupRoles, err = group.GetGroupRolesOfUser(ctx, user)
	}
	if err != nil {
		return nil, err
	}
	roleMembers := []*RoleMembers{}
	for _, g := range groupRoles {
		r, err := GetRole(ctx, g.Id)
		if err == ErrRoleNotFound {
			continue
		} else if err != nil {
			return nil, err
		} else {
			groupMembers, err := g.ListMembers(ctx)
			if err != nil {
				return nil, err
			}
			if r.AppId == appId {
				roleMembers = append(roleMembers, &RoleMembers{
					Role:    *r,
					Type:    g.Role,
					Members: groupMembers,
				})
			}
		}
	}
	return roleMembers, nil

}

func IsUserInRole(ctx *models.Context, user iuser.User, role *Role) (bool, group.MemberRole) {
	g, err := group.GetGroup(ctx, role.Id)
	if err != nil {
		return false, group.NORMAL
	}

	ok, t, err := g.GetMember(ctx, user)
	if err != nil {
		return false, group.NORMAL
	}
	return ok, t
}

func UpdateRoleResource(ctx *models.Context, roleId int, resourcesIds []int) error {
	tx := ctx.DB.MustBegin()
	_, err0 := tx.Exec("DELETE FROM role_resource WHERE role_id=?", roleId)
	if err0 != nil {
		return err0
	}
	for _, rId := range resourcesIds {
		_, err1 := tx.Exec(
			"INSERT INTO role_resource (role_id, resource_id) VALUES (?, ?)",
			roleId, rId)
		if err1 != nil {
			tx.Rollback()
			return err1
		}
	} 
	err2 := tx.Commit()
	if err2 != nil {
		return err2
	}
	return nil
}

func AddRoleResource(ctx *models.Context, roleId int, resourcesIds []int) error {
	// only leaf role can have resources

	tx := ctx.DB.MustBegin()
	for _, rId := range resourcesIds {
		_, err1 := tx.Exec(
			"INSERT INTO role_resource (role_id, resource_id) VALUES (?, ?)",
			roleId, rId)
		if err1 != nil {
			return err1
		}
	}
	err2 := tx.Commit()
	if err2 != nil {
		return err2
	}
	return nil
}


func IsLeafRole(ctx *models.Context, roleId int) bool {
	log.Debug(roleId)
	rows, err := ctx.DB.Query("SELECT resource_id FROM role_resource WHERE role_id=?", roleId)
	if err != nil {
		log.Debug(err)
		if err == sql.ErrNoRows {
			return false
		} else {
			panic(err)
		}
	}
	resourceIds := []int{}
	var tmpId int
	for rows.Next() {
		if err = rows.Scan(&tmpId); err != nil {
			log.Error(err)
			continue
		}

		resourceIds = append(resourceIds, tmpId)
	}
	log.Debug(resourceIds)
	if len(resourceIds) == 0 {
		return false
	}

	return true
}

func RemoveRoleResource(ctx *models.Context, roleId int, resourceIds []int) error {
	tx := ctx.DB.MustBegin()
	for _, rId := range resourceIds {
		_, err1 := tx.Exec("DELETE FROM role_resource WHERE role_id=? AND resource_id=?",
			roleId, rId)
		if err1 != nil {
			return err1
		}
	}
	if err2 := tx.Commit(); err2 != nil {
		return err2
	}
	return nil
}
