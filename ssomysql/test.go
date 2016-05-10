package ssomysql

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/testhelper"
	"github.com/laincloud/sso/ssomysql/user"
)

type TestHelper struct {
	T   *testing.T
	Ctx context.Context
}

var testBack iuser.UserBackend

func NewTestHelper(t *testing.T) *TestHelper {
	ctx := context.Background()
	mctx := testhelper.NewTestHelper(t).Ctx
	testBack = user.New(testhelper.GetTestMysqlDSN())
	mctx.Back = testBack
	ctx = context.WithValue(ctx, "mctx", mctx)
	ctx = context.WithValue(ctx, "emailSuffix", "@example.com")
	ctx = context.WithValue(ctx, "userBackend", testBack)

	th := &TestHelper{
		T:   t,
		Ctx: ctx,
	}
	return th
}

func (th *TestHelper) loginWithScope(username string, scope string) {
	ub := getUserBackend(th.Ctx)
	u, err := ub.GetUserByName(username)
	assert.Nil(th.T, err)
	th.Ctx = context.WithValue(th.Ctx, "user", u)
	th.Ctx = context.WithValue(th.Ctx, "scope", strings.Split(scope, " "))
}

func (th *TestHelper) logout() {
	th.Ctx = context.WithValue(th.Ctx, "user", nil)
	th.Ctx = context.WithValue(th.Ctx, "scope", nil)
}

func createTestUser(th *TestHelper, username string) iuser.User {
	//mctx := getModelContext(th.Ctx)
	err := testBack.CreateUser(&user.User{
		Name:         username,
		FullName:     username,
		PasswordHash: []byte("test"),
	}, true)
	assert.Nil(th.T, err)
	u, err := testBack.GetUserByName(username)
	assert.Nil(th.T, err)
	return u
}

func callPostGroupsWithGroupName(th *TestHelper, groupname string) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/groups",
		strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, groupname)))
	return ssolib.GroupsResource{}.Post(th.Ctx, r)
}

func callDeleteGroupMember(th *TestHelper, groupname, username string) (int, interface{}) {
	r, _ := http.NewRequest("DELETE",
		fmt.Sprintf("http://sso.example.com/api/groups/%s/members/%s",
			groupname, username),
		nil)
	aMock := ssolib.MockParams(th.T, map[string]string{
		"groupname": groupname,
		"username":  username})
	defer aMock.Restore()

	return ssolib.MemberResource{}.Delete(th.Ctx, r)
}
