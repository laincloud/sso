package iuser

import (
	"errors"
)

/* Backend Group 的语义是：sso lib 里的 group 表里的一项，
其 id, name 同普通 group 一样位于 group 表内;
但是，其创建逻辑，删除逻辑，成员逻辑由 BackendGroup 接口定义.
一般来说，无论创建还是删除，逻辑上先执行 BackendGroup 里的方法，
然后执行 ssolib 里的方法，最后将 Id 关联起来.
如果成功才调用 ssolib 里的方法.
关于调用参数, 为了提供一些灵活性，CreateBackendGroup 的最后一个参数是
rules interface{}; 用来表示后端相关的参数；

如果后端 group 和 lib group 事务方面有逻辑上的依赖，注意一致性和死锁.
*/

// 总体来说，lib 可能有多个 user backend, 每个 backend 会有多种 group
// 目前只准备实现一种 backend group.

var (
	ErrMethodNotSupported = errors.New("methods not supported")
)

// Type of Backend Group or SSOLIBGROUP
type GroupType int8

const (
	SSOLIBGROUP GroupType = iota
	BACKENDGROUP
)

// 拥有 BackendGroup 的 Backend
type BackendWithGroup interface {
	UserBackend

	// 给定 group 名字，返回是否创建成功，若不成功，建议返回 error 非空
	// 若 success 为 false, 可以返回非空 group, 表示已存在满足 rules 的 group
	CreateBackendGroup(name string, rules interface{}) (success bool, group BackendGroup, err error)

	// name 为 ssolib 中的组名
	GetBackendGroupByName(name string) (BackendGroup, error)

	// id 为 ssolib 中的组 Id
	GetBackendGroup(id int) (BackendGroup, error)

	DeleteBackendGroup(BackendGroup) error

	GetBackendGroupsOfUser(user User) ([]BackendGroup, error)

	// 现在还想不清楚 groupId 到底重不重要，按照已有的函数，可能会需要下面的接口函数
	SetBackendGroupId(name string, Id int) error

	//	GetBackendGroupsOfUserByIds(userIds []int) (map[int][]BackendGroup, error)
	//	RemoveUserFromAllBackendGroups(user User) error

}

// ssolib 中的 group 有 member 的概念，即 User + role
// 为了简单，Backend Group 不规定 Role 的接口
type BackendGroup interface {
	GetId() int
	GetName() string

	// rules as fullnames
	GetRules() interface{}

	AddUser(user User) error
	GetUser(u User) (bool, error)
	RemoveUser(user User) error
	ListUsers() ([]User, error)
}
