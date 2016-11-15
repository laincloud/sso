package group

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/mijia/sweb/log"
)

var (
	ErrGroupNotFound      = errors.New("Group not found")
	ErrBackendUnsupported = errors.New("unsupported backend type")
)

const createGroupTableSQL = `
CREATE TABLE IF NOT EXISTS ` + "`group`" + ` (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(128) NOT NULL,
	fullname VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	backend TINYINT NOT NULL DEFAULT 0,
	PRIMARY KEY (id),
	UNIQUE KEY (name)
) DEFAULT CHARSET=latin1`

type Group struct {
	Id       int
	Name     string
	FullName string
	Created  string
	Updated  string

	GroupType iuser.GroupType `db:"backend"` // default is 0 for ssolib-mysql
}

func InitDatabase(ctx *models.Context) {
	ctx.DB.MustExec(createGroupTableSQL)
	ctx.DB.MustExec(createUserGroupTableSQL)
	if valid {
		ctx.DB.MustExec(createGroupDAGTableSQL)
		ctx.DB.MustExec(createGroupDepthTableSQL)
	}
}

func CreateGroup(ctx *models.Context, group *Group) (*Group, error) {
	log.Debug("CreateGroup:")
	log.Debug("type:", group.GroupType)
	tx := ctx.DB.MustBegin()
	result, err := tx.Exec(
		"INSERT INTO `group` (name, fullname, backend) VALUES (?, ?, ?)",
		group.Name, group.FullName, group.GroupType)

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
		return nil, err
	}

	if valid {
		err = writeGroupDepth(ctx, int(id), 1)
		if err != nil {
			panic(err)
		}
	}
	return GetGroup(ctx, int(id))
}

func GetGroupByName(ctx *models.Context, name string) (*Group, error) {
	group := Group{}
	err := ctx.DB.Get(&group, "SELECT * FROM `group` WHERE name=?", name)
	log.Debug(err)
	if err == sql.ErrNoRows {
		return nil, ErrGroupNotFound
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

func ListGroups(ctx *models.Context, ids ...int) ([]Group, error) {
	query := "SELECT * FROM `group`"
	var args []interface{}
	if len(ids) > 0 {
		_query, _args, err := sqlx.In("SELECT * FROM `group` WHERE id IN(?)", ids)
		if err != nil {
			return nil, err
		}
		query = _query
		args = _args
	}

	groups := []Group{}
	err := ctx.DB.Select(&groups, query, args...)
	return groups, err
}

func GetGroup(ctx *models.Context, id int) (*Group, error) {
	group := Group{}
	err := ctx.DB.Get(&group, "SELECT * FROM `group` WHERE id=?", id)
	if err == sql.ErrNoRows {
		return nil, ErrGroupNotFound
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

// 1. 作为子组：从多个父节点删掉该组
// 2. 作为父组：删掉多个与子组的关联关系
// 3. 删掉 groupdepth 表项
// 4. 删掉 member
// 5. 删掉 group 表项
func DeleteGroup(ctx *models.Context, group *Group) error {
	log.Debugf("DeleteGroup %s", group.Name)

	if valid {
		lock := ctx.Lock
		lock.Lock()
		defer func() {
			lock.Unlock()
		}()

		fathers, err := ListGroupFathersById(ctx, group.Id)
		if err != nil {
			return err
		}
		for _, v := range fathers {
			err = v.RemoveGroupMemberWithoutLock(ctx, group)
			if err != nil {
				return err
			}
		}

		sons, err := group.ListGroupMembers(ctx)
		if err != nil {
			return nil
		}
		for _, v := range sons {
			son := v.Group
			err = group.RemoveGroupMemberWithoutLock(ctx, &son)
			if err != nil {
				return err
			}
		}

		if err := deleteGroupDepth(ctx, group); err != nil {
			return err
		}
	}

	if err := group.removeAllMembers(ctx); err != nil {
		return err
	}

	return deleteGroup(ctx, group)
}

func deleteGroup(ctx *models.Context, group *Group) error {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec("DELETE FROM `group` WHERE id=?", group.Id)

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}

	if err != nil {
		return err
	}

	return nil
}
