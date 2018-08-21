package user

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models/iuser"
)

const createLdapGroupTableSQL = `
CREATE TABLE IF NOT EXISTS ldapgroup (
	id INT NOT NULL,
	name VARCHAR(32) NOT NULL,
	fullname VARCHAR(255) CHARACTER SET utf8 NOT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (name),
	UNIQUE KEY (fullname)
) DEFAULT CHARSET=latin1`

var (
	ErrOUNotFound        = errors.New("Ldap ou (Organizational Unit) is not found ")
	ErrLDAPGroupNotFound = errors.New("Ldap group is not found")
)

type LdapGroup struct {
	Id          int
	Name        string
	LdapOUsRule string `db:"fullname"`
	Created     string
	Updated     string

	back *UserBack
}

func (g *LdapGroup) GetId() int {
	return g.Id
}

func (g *LdapGroup) GetName() string {
	return g.Name
}

func (g *LdapGroup) GetRules() interface{} {
	return g.LdapOUsRule
}

func (g *LdapGroup) AddUser(user iuser.User) error {
	return iuser.ErrMethodNotSupported
}

func (g *LdapGroup) GetUser(u iuser.User) (bool, error) {
	log.Debug(u.(*User).dn)
	log.Debug(g.LdapOUsRule)
	if strings.HasSuffix(u.(*User).dn, g.LdapOUsRule) {
		return true, nil
	} else {
		return false, nil
	}
}

func (g *LdapGroup) RemoveUser(user iuser.User) error {
	return iuser.ErrMethodNotSupported
}

func (g *LdapGroup) ListUsers() ([]iuser.User, error) {
	return nil, iuser.ErrMethodNotSupported
}

func (ub *UserBack) CreateBackendGroup(name string, rules interface{}) (bool, iuser.BackendGroup, error) {
	log.Debug("begin creating backend group")
	defer log.Debug("backend group created")
	ouStr := rules.(string)
	result, err := ub.C.SearchForOU(ouStr)
	if err != nil {
		return false, nil, err
	}
	if len(result.Entries) > 1 {
		return false, nil, errors.New(ouStr + " OU is not unique")
	} else if len(result.Entries) < 1 {
		return false, nil, errors.New(ouStr + " OU not exist")
	} else {
		var fullname = result.Entries[0].DN
		lGroup, err := ub.GetGroupByFullname(fullname)
		if err != nil {
			if err != ErrLDAPGroupNotFound {
				panic(err)
			}
		}
		if lGroup == nil {
			lGroup, err = ub.createGroup(name, fullname)
			if err != nil {
				panic(err)
			}
			return true, lGroup, nil
		} else {
			return false, lGroup, errors.New("the group with name " + lGroup.GetName() + " has the same ldap rules")
		}
	}

}

func (ub *UserBack) GetBackendGroup(id int) (iuser.BackendGroup, error) {
	group := LdapGroup{}
	err := ub.DB.Get(&group, "SELECT * FROM ldapgroup WHERE id=?", id)
	log.Debug(err)
	if err == sql.ErrNoRows {
		return nil, ErrLDAPGroupNotFound
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

func (ub *UserBack) GetBackendGroupByName(name string) (iuser.BackendGroup, error) {
	group := LdapGroup{}
	err := ub.DB.Get(&group, "SELECT * FROM ldapgroup WHERE name=?", name)
	log.Debug(err)
	if err == sql.ErrNoRows {
		return nil, ErrLDAPGroupNotFound
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

func (ub *UserBack) GetGroupByFullname(fullname string) (iuser.BackendGroup, error) {
	group := LdapGroup{}
	err := ub.DB.Get(&group, "SELECT * FROM ldapgroup WHERE fullname=?", fullname)
	log.Debug(err)
	if err == sql.ErrNoRows {
		return nil, ErrLDAPGroupNotFound
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

func (ub *UserBack) DeleteBackendGroup(g iuser.BackendGroup) error {
	return iuser.ErrMethodNotSupported
}

func (ub *UserBack) GetBackendGroupsOfUser(user iuser.User) ([]iuser.BackendGroup, error) {
	groups := []iuser.BackendGroup{}
	dn := user.(*User).dn
	for {
		ouIndex := strings.Index(dn, ",OU=")
		if ouIndex < 0 {
			break
		}
		lastou := strings.Index(dn, ",OU=HABROOT")
		if ouIndex == lastou {
			break
		}
		dn = dn[(ouIndex + 1):]
		log.Debug(dn)
		group, err := ub.GetGroupByFullname(dn)
		if err != nil {
			if err != ErrLDAPGroupNotFound {
				return nil, err
			}
		}
		if group != nil {
			groups = append(groups, group)
		}
	}
	log.Debug(groups)
	return groups, nil
}

func (ub *UserBack) SetBackendGroupId(name string, Id int) error {
	tx := ub.DB.MustBegin()
	_, err1 := tx.Exec(
		"UPDATE ldapgroup SET id=? WHERE name=?",
		Id, name)

	err2 := tx.Commit()

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}
	return nil
}

func (ub *UserBack) createGroup(name string, fullname string) (iuser.BackendGroup, error) {
	tx := ub.DB.MustBegin()
	_, err := tx.Exec(
		"INSERT INTO ldapgroup (id, name, fullname) VALUES (?, ?, ?)",
		0, name, fullname)

	if err2 := tx.Commit(); err2 != nil {
		return nil, err2
	}

	if err != nil {
		return nil, err
	}

	return ub.GetBackendGroupByName(name)
}
