package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssoldap/user/ldap"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/utils"
)

// the feature in this file is either the email(also as upn), or the email prefix
// if the user is in the ldap backend, only the email and id are valid in the user table; in this code as a example, the mobile is also valid, but the name should be the prefix of the UPN(email) by default, so not consider the old data.
// if the user is not in the ldap, all the fields in the user table is valid

const (
	LDAPREALMNAME = "sso-ldap"
)

// for compatible with sso database
const createUserTableSQL = `
CREATE TABLE IF NOT EXISTS user (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(32) NOT NULL DEFAULT "",
	fullname VARCHAR(128) CHARACTER SET utf8 NOT NULL DEFAULT "",
	email VARCHAR(64) NOT NULL DEFAULT "",
	password VARBINARY(60) NOT NULL DEFAULT "",
	mobile VARCHAR(11) NOT NULL DEFAULT "",
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY (name),
	UNIQUE KEY (email)
) DEFAULT CHARSET=latin1
`

var EMAIL_SUFFIX string

var (
	ErrForbidden = errors.New("functions not developed")
)

type UserBack struct {
	C  *ldap.LdapClient
	DB *sqlx.DB

	MailId map[string]int
}

func New(ldapUrl, cn, passwd string, mysqlDSN string, email string, ldapBase string) *UserBack {
	// ldap://xxx.xxx.xxx.xxx:389/
	u, err := url.ParseRequestURI(ldapUrl)
	if err != nil {
		// todo 虽然我们不能强耦合ldap, 这里先这样，之后重构
		log.Errorf("failed to parse ldapUrl %s, err %+v", ldapUrl, err)
		return nil
	}

	port, _ := strconv.Atoi(u.Port())
	client, err := ldap.NewClient(u.Hostname(), port, ldapBase, cn, passwd)
	if err != nil {
		panic(err)
	}
	db, err := utils.InitMysql(mysqlDSN)
	if err != nil {
		panic(err)
	}
	EMAIL_SUFFIX = email
	return &UserBack{
		C:      client,
		DB:     db,
		MailId: map[string]int{}, //used as cache TODO
	}
}

func (ub *UserBack) InitDatabase() {
	tx := ub.DB.MustBegin()
	tx.MustExec(createLdapGroupTableSQL)
	tx.MustExec(createUserTableSQL)
	ub.DB.SetMaxOpenConns(50)
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func (ub *UserBack) ListUsers(ctx context.Context) ([]iuser.User, error) {
	// 加一个参数, 来防止太长的返回值

	// FIXME 临时解决方案
	db := ctx.Value("db").(*sqlx.DB)
	userIds := []int{}
	//	err := db.Select(&userIds, "SELECT DISTINCT user_id FROM user_group")
	err := db.Select(&userIds, "SELECT id FROM user")
	ret := make([]iuser.User, len(userIds))

	log.Debug(userIds)
	for i, v := range userIds {
		log.Debug(i, v)
		ret[i], err = ub.GetUser(int(v))
		log.Debug(err)
	}
	return ret, nil
}

func (ub *UserBack) AddUser(info UserInfo) error {
	// first, check whether the user exists
	// if yes, update the info;
	// else, check if the user in the ldap; if so, return forbidden
	// otherwise the user should be local and will be inserted in the local mysql user table;
	// but there may be a problem if a user with that name will be added in the ldap;
	// so, the admin should add some user with a particular name.
	var passwordHash []byte
	var err error
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(info.Password), BCRYPT_COST)
	if err != nil {
		return err
	}

	if _, err := ub.getUserByEmailFromMysql(info.Email); err == iuser.ErrUserNotFound {
		_, err := ub.Search(info.Email)
		if err != nil {
			if err != iuser.ErrUserNotFound {
				return err
			} else {
				// create the user in the local mysql

				tx := ub.DB.MustBegin()
				_, err1 := tx.Exec(
					"INSERT INTO user (name, fullname, email, password, mobile) "+
						"VALUES (?, ?, ?, ?, ?)",
					info.Name, info.FullName, info.Email, passwordHash, info.Mobile)

				if err2 := tx.Commit(); err2 != nil {
					return err2
				}
				if err1 != nil {
					return err1
				}
				return nil
			}
		} else {
			return ErrForbidden
		}
	} else if err != nil {
		return err
	} else {
		// update the user info
		tx := ub.DB.MustBegin()
		_, err1 := tx.Exec(
			"UPDATE user SET fullname=?, password=?, mobile=? where name=? ",
			info.FullName, passwordHash, info.Mobile, info.Name)
		if err2 := tx.Commit(); err2 != nil {
			return err2
		}
		if err1 != nil {
			return err1
		}
		return nil
	}
	return nil
}

func (ub *UserBack) GetUser(id int) (iuser.User, error) {
	if id < 0 {
		log.Error("unexpect id")
		return nil, nil
	}
	upn, err := ub.getUPNById(id)
	if err != nil {
		log.Debug(id)
		log.Error(err)
		return nil, err
	}
	user, err := ub.Search(upn)
	if err != nil {
		if err != iuser.ErrUserNotFound {
			return user, err
		} else {
			user, err = ub.getUserFromMysql(id)
			if err != nil {
				log.Error(err)
				return nil, err
			}
		}
	}
	user.SetBackend(ub)

	return user, nil
}

func (ub *UserBack) GetUserByName(name string) (iuser.User, error) {
	return ub.GetUserByEmail(name + EMAIL_SUFFIX)
}

func (ub *UserBack) GetUserByEmail(email string) (iuser.User, error) {
	log.Debug(email)
	user, err := ub.Search(email)
	if err != nil {
		if err != iuser.ErrUserNotFound {
			return user, err
		} else {
			user, err = ub.getUserByEmailFromMysql(email)
			if err != nil {
				return nil, err
			}
		}
	}
	user.SetBackend(ub)
	return user, nil
}

func (ub *UserBack) CreateUser(user iuser.User, passwordHashed bool) error {
	return ErrForbidden
}

func (ub *UserBack) DeleteUser(user iuser.User) error {
	id := user.GetId()
	upn, err := ub.getUPNById(id)
	if err != nil {
		log.Debug(id)
		log.Error(err)
		return err
	}
	_, err = ub.Search(upn)
	if err != nil {
		if err != iuser.ErrUserNotFound {
			return err
		} else { // if user is local, delete it
			tx := ub.DB.MustBegin()
			_, err1 := tx.Exec(
				"DELETE FROM user WHERE id=?",
				user.GetId())

			if err2 := tx.Commit(); err2 != nil {
				return err2
			}
			if err1 != nil {
				return err1
			}
			return nil
		}
	} else {
		return ErrForbidden
	}
}

// deprecated
func (ub *UserBack) AuthPassword(sub, passwd string) (bool, error) {
	log.Debug(sub)
	id, err := ub.UserSubToId(sub)
	if err != nil {
		log.Error(err)
		return false, err
	}
	u, err := ub.GetUser(id)
	log.Debug(id)
	if err != nil {
		log.Debug(err)
		return false, err
	}
	b, _ := json.Marshal(u)
	log.Debug(string(b))
	return ub.C.Auth(u.(*User).Email, passwd)
}

func (ub *UserBack) AuthPasswordByFeature(feature, passwd string) (bool, iuser.User, error) {
	if passwd == "" {
		return false, nil, errors.New("passwd required")
	}
	if !strings.HasSuffix(feature, EMAIL_SUFFIX) {
		feature = feature + EMAIL_SUFFIX
	}
	success, err := ub.C.Auth(feature, passwd)
	log.Debug(err)
	if success {
		u, err := ub.GetUserByEmail(feature)
		return true, u, err
	} else {
		_, err = ub.Search(feature)
		if err != nil { //if in ladp, not use the local password
			log.Debug(err)
			if err != iuser.ErrUserNotFound {
				return false, nil, nil
			} else {
				u, err := ub.getUserByEmailFromMysql(feature)
				if err == nil && u != nil {
					if u.VerifyPassword([]byte(passwd)) {
						return true, u, err
					}
				}
			}
		}
	}
	return false, nil, nil
}

func (ub *UserBack) GetUserByFeature(f string) (iuser.User, error) {
	if strings.HasSuffix(f, EMAIL_SUFFIX) {
		return ub.GetUserByEmail(f)
	} else {
		return ub.GetUserByEmail(f + EMAIL_SUFFIX)
	}
}

func (ub *UserBack) Name() string {
	return LDAPREALMNAME
}

func (ub *UserBack) SupportedVerificationMethods() []string {
	ret := []string{}
	ret = append(ret, iuser.PASSWORD)
	return ret
}

func (ub *UserBack) UserIdToSub(id int) string {
	return LDAPREALMNAME + fmt.Sprint(id)
}

func (ub *UserBack) UserSubToId(sub string) (int, error) {
	if !strings.HasPrefix(sub, LDAPREALMNAME) {
		return -1, iuser.ErrInvalidSub
	} else {
		return strconv.Atoi(sub[len(LDAPREALMNAME):])
	}
}

func (ub *UserBack) getUserFromMysql(id int) (*User, error) {
	user := User{}
	err := ub.DB.Get(&user, "SELECT * FROM user WHERE id=?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, iuser.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (ub *UserBack) getUserByEmailFromMysql(email string) (*User, error) {
	user := User{}
	err := ub.DB.Get(&user, "SELECT * FROM user WHERE email=?", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, iuser.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil

}
