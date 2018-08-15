package iuser

import (
	"errors"
	"golang.org/x/net/context"
)

var (
	ErrUserNotFound = errors.New("User not found")
	ErrInvalidSub   = errors.New("Invalid sub")
)

const (
	PASSWORD string = "password"
	SMS      string = "sms"
)

type UserBackend interface {
	// 全局已经注册两个认证用户身份的方法：
	// "password" "sms"
	SupportedVerificationMethods() []string

	// 可能的返回值为 "sso" "sso-ldap"
	Name() string

	// 应该有一种将 id 对应成全局 sub 的方法
	UserIdToSub(id int) string

	UserSubToId(sub string) (int, error)

	GetUserByFeature(string) (User, error)
	GetUserByEmail(string) (User, error)
	GetUser(id int) (User, error)

	AuthPassword(sub, password string) (bool, error)
	AuthPasswordByFeature(feature, password string) (bool, User, error)

	// 下面的方法是为了简单地兼容 sso 原有代码，也许会有很多冗余
	ListUsers(ctx context.Context) ([]User, error)
	GetUserByName(name string) (User, error)
	CreateUser(user User, passwordHashed bool) error
	DeleteUser(User) error

	// 处理一些如管理员组设定等依赖后端的初始化
	InitModel(model interface{})

	// 添加一些与后端相关的 handler, 如用户注册相关等
	//AddHandlers(*ssolib.Server)

}
