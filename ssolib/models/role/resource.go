package role

import (
	"container/list"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

const createResourceTableSQL = `
CREATE TABLE IF NOT EXISTS resource (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	fullname VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	app_id INT NOT NULL,
	data VARCHAR(1024) CHARACTER SET utf8 NOT NULL,
	owner VARCHAR(64) NOT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) DEFAULT CHARSET=latin1`


type Resource struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `db:"fullname" json:"description"`
	AppId       int    `db:"app_id" json:"app_id"`
	Data        string `json:"data"`
	Owner       string `json:"owner"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

type RoleResources struct {
	RoleId    int        `json:"role_id"`
	Resources []Resource `json:"resources"`
}

func CreateResource(ctx *models.Context, resource *Resource) (*Resource, error) {
	log.Debug("CreateResource")
	tx := ctx.DB.MustBegin()
	result, err := tx.Exec(
		"INSERT INTO resource (name, fullname, app_id, data, owner) VALUES(?, ?, ?, ?, ?)", resource.Name, resource.Description, resource.AppId, resource.Data, resource.Owner)
	if err2 := tx.Commit(); err2 != nil {
		log.Debug(err2)
		return nil, err2
	}
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	return GetResource(ctx, int(id))
}

func GetResource(ctx *models.Context, id int) (*Resource, error) {
	resource := Resource{}
	err := ctx.DB.Get(&resource, "SELECT * FROM resource WHERE id=?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &resource, nil
}

func UpdateResource(ctx *models.Context, id int, name string, desc string, data string) (*Resource, error) {
	tx := ctx.DB.MustBegin()
	_, err1 := tx.Exec("UPDATE resource SET name=?, fullname=?, data=? WHERE id=?", name, desc, data, id)
	err2 := tx.Commit()
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	return GetResource(ctx, id)
}

func DeleteResource(ctx *models.Context, id int) error {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec("DELETE FROM role_resource WHERE resource_id=?", id)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err1 := tx.Exec("DELETE FROM resource WHERE id=?",
		id)
	if err2 := tx.Commit(); err2 != nil {
		return err2
	}
	if err1 != nil {
		return err1
	}
	return nil
}

func GetResourcesByIds(ctx *models.Context, ids []int) ([]Resource, error) {
	if len(ids) == 0 {
		return []Resource{}, nil
	}
	query, args, err := sqlx.In("SELECT * FROM resource WHERE id IN(?)", ids)
	if err != nil {
		return nil, err
	}
	resources := []Resource{}
	err = ctx.DB.Select(&resources, query, args...)
	return resources, err
}

func GetResourcesByRoleId(ctx *models.Context, roleId int) ([]Resource, error) {
	rIds, err := getResourceIdsByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}
	return GetResourcesByIds(ctx, rIds)
}

func getResourceIdsByRoleId(ctx *models.Context, roleId int) ([]int, error) {
	resourceIds := []int{}
	rows, err := ctx.DB.Query(
		"SELECT resource_id FROM role_resource WHERE role_id=?",
		roleId)
	if err != nil {
		if err == sql.ErrNoRows {
			return resourceIds, nil
		}
		return nil, err
	}
	for rows.Next() {
		var resId int
		if err = rows.Scan(&resId); err != nil {
			return nil, err
		}
		resourceIds = append(resourceIds, resId)
	}
	return resourceIds, nil
}

func GetResourcesByRoleIds(ctx *models.Context, roleIds []int) ([]Resource, error) {
	rIds, err := getResourceIdsByRoleIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	return GetResourcesByIds(ctx, rIds)
}

func getResourceIdsByRoleIds(ctx *models.Context, roleIds []int) ([]int, error) {
	if len(roleIds) == 0 {
		return []int{}, nil
	}
	query, args, err := sqlx.In("SELECT resource_id FROM role_resource WHERE role_id IN(?)", roleIds)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	rIds := []int{}
	err = ctx.DB.Select(&rIds, query, args...)
	return rIds, err
}

func GetResources(ctx *models.Context, appId int, user iuser.User) ([]Resource, error) {
	rIds, err := getLeafRoleIds(ctx, user, appId)
	if err != nil {
		return nil, err
	}
	return GetResourcesByRoleIds(ctx, rIds)
}

func GetResourcesForRole(ctx *models.Context, appId int) ([]RoleResources, error) {
	roles, err := GetRolesByAppId(ctx, appId)
	if err != nil {
		return nil, err
	}
	ret := []RoleResources{}
	for _, role := range roles {
		resources, err := GetResourcesByRoleId(ctx, role.Id)
		if err != nil {
			return nil, err
		}
		ret = append(ret, RoleResources{
			RoleId:	   role.Id,
			Resources: resources,
		})
	}
	return ret, nil
}

func GetResourcesForRoleByUser(ctx *models.Context, appId int, user iuser.User) ([]RoleResources, error) {
	// for leaf role and it's resources
	rIds, err := getLeafRoleIds(ctx, user, appId)
	if err != nil {
		return nil, err
	}
	ret := []RoleResources{}
	m := make(map[int]struct{})
	for _, rId := range rIds {
		ress, err := GetResourcesByRoleId(ctx, rId)
		if err != nil {
			return nil, err
		}
		if _, ok := m[rId]; ok{
			continue
		}
		ret = append(ret, RoleResources{
			RoleId:    rId,
			Resources: ress,
		})
		m[rId] = struct{}{}
	}
	return ret, nil
}

func getLeafRoleIds(ctx *models.Context, user iuser.User, appId int) (roleIds []int, err error) {
	roleIds = []int{}
	roles, err := GetDirectRoles(ctx, user, appId)
	if err != nil {
		return nil, err
	}
	l := list.New()
	for _, r := range roles {
		l.PushBack(r)
	}
	for {
		if l.Len() == 0 {
			break
		}
		iter := l.Front()
		v := iter.Value.(*Role)
		subRoles, err := getSubRoles(ctx, v.Id)
		if err == ErrLeafRoleHasNoSubRole {
			roleIds = append(roleIds, v.Id)
		} else if err != nil {
			return nil, err
		} else {
			for _, subRole := range subRoles {
				l.PushBack(subRole)
			}
		}
		l.Remove(iter)
	}
	return roleIds, nil
}

func GetAllResources(ctx *models.Context, appId int) ([]Resource, error) {
	res := []Resource{}
	err := ctx.DB.Select(&res, "SELECT * FROM resource WHERE app_id=?", appId)
	if err == sql.ErrNoRows {
		return res, nil
	} else if err != nil {
		log.Error(err)
		return nil, err
	} else {
		return res, nil
	}
}
