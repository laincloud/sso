package user

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"

	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/utils"
)

const createUserTableSQL = `
CREATE TABLE IF NOT EXISTS user (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(32) NOT NULL,
	fullname VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	email VARCHAR(64) NULL DEFAULT NULL,
	password VARBINARY(60) NOT NULL,
	mobile VARCHAR(11) NULL DEFAULT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY (name),
	UNIQUE KEY (email)
) DEFAULT CHARSET=latin1
`

const (
	SSOREALMNAME = "sso-mysql"
)

type UserBack struct {
	DB *sqlx.DB
}

func New(dsn string, adminEmail string, adminPasswd string) *UserBack {
	db, err := utils.InitMysql(dsn)
	if err != nil {
		panic(err)
	}
	ADMINEMAIL = adminEmail
	ADMINPASSWD = adminPasswd
	return &UserBack{
		DB: db,
	}
}

func (ub *UserBack) InitDatabase() {
	tx := ub.DB.MustBegin()
	tx.MustExec(createUserTableSQL)
	tx.MustExec(createUserRegisterTableSQL)
	tx.MustExec(createResetPasswordTableSQL)
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func (ub *UserBack) ListUsers(ctx context.Context) ([]iuser.User, error) {
	// TODO 加一个参数, 来防止太长的返回值
	users := []User{}
	err := ub.DB.Select(&users, "SELECT * FROM user")
	ret := make([]iuser.User, len(users))
	// 下面的代码需要测试
	for i, _ := range users {
		users[i].SetBackend(ub)
		ret[i] = &(users[i])
	}
	return ret, err
}

func (ub *UserBack) DeleteAllActivationCodes(ctx context.Context) error {
	tx := ub.DB.MustBegin()
	_, err1 := tx.Exec("DELETE FROM user_register")

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}
	if err1 != nil {
		return err1
	}
	return nil
}

func (ub *UserBack) ListInactiveUsers(ctx context.Context) ([]UserRegistration, error) {
	users, err := ub.ListUsers(ctx)
	if err != nil {
		panic(err)
	}

	inactiveUsers := []UserRegistration{}
	err = ub.DB.Select(&inactiveUsers, "SELECT * FROM user_register")

	if err != nil {
		panic(err)
	}

	activeUsers := make(map[string]struct{})
	activeUserEmails := make(map[string]struct{})
	for _, u := range users {
		activeUsers[u.GetName()] = struct{}{}
		activeUserEmails[u.(*User).GetEmail()] = struct{}{}
	}

	ret := []UserRegistration{}
	for _, u := range inactiveUsers {
		if _, ok := activeUsers[u.Name]; !ok {
			if _, ok := activeUserEmails[u.Email.String]; !ok {
				u.Password = ""
				ret = append(ret, u)
			}
		}
	}
	return ret, nil
}

func (ub *UserBack) GetUser(id int) (iuser.User, error) {
	user := User{}
	err := ub.DB.Get(&user, "SELECT * FROM user WHERE id=?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, iuser.ErrUserNotFound
		}
		return nil, err
	}
	user.SetBackend(ub)
	return &user, nil
}

func (ub *UserBack) GetUserByName(name string) (iuser.User, error) {
	user := User{}
	// TODO sql 注入
	err := ub.DB.Get(&user, "SELECT * FROM user WHERE name=?", name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, iuser.ErrUserNotFound
		}
		return nil, err
	}
	user.SetBackend(ub)
	return &user, nil
}

func (ub *UserBack) GetUserByEmail(email string) (iuser.User, error) {
	user := User{}
	err := ub.DB.Get(&user, "SELECT * FROM user WHERE email=?", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, iuser.ErrUserNotFound
		}
		return nil, err
	}
	user.SetBackend(ub)
	return &user, nil
}

func (ub *UserBack) CreateUser(user iuser.User, passwordHashed bool) error {
	var passwordHash []byte
	if passwordHashed {
		passwordHash = []byte(user.(*User).GetPasswordHash())
	} else {
		var err error
		passwordHash, err = generateHashFromPassword([]byte(user.(*User).GetPasswordHash()), BCRYPT_COST)
		if err != nil {
			return err
		}
	}

	tx := ub.DB.MustBegin()
	_, err1 := tx.Exec(
		"INSERT INTO user (name, fullname, email, password, mobile) "+
			"VALUES (?, ?, ?, ?, ?)",
		user.GetName(), user.(*User).GetFullName(), user.(*User).GetEmail(), passwordHash, user.GetMobile())

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}
	if err1 != nil {
		return err1
	}

	return nil
}

func (ub *UserBack) DeleteUser(user iuser.User) error {
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
	if u.(*User).VerifyPassword([]byte(passwd)) {
		return true, nil
	}
	return false, nil
}

func (ub *UserBack) AuthPasswordByFeature(feature, passwd string) (bool, iuser.User, error) {
	return false, nil, nil
}

func (ub *UserBack) GetUserByFeature(f string) (iuser.User, error) {
	return ub.GetUserByName(f)
}

func (ub *UserBack) Name() string {
	return SSOREALMNAME
}

func (ub *UserBack) SupportedVerificationMethods() []string {
	ret := []string{}
	ret = append(ret, iuser.PASSWORD)
	return ret
}

func (ub *UserBack) UserIdToSub(id int) string {
	return SSOREALMNAME + fmt.Sprint(id)
}

func (ub *UserBack) UserSubToId(sub string) (int, error) {
	if !strings.HasPrefix(sub, SSOREALMNAME) {
		return -1, iuser.ErrInvalidSub
	} else {
		return strconv.Atoi(sub[len(SSOREALMNAME):])
	}
}
