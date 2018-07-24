package ssolib

import (
	"strings"
	"testing"

	//	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/oauth2"
	"github.com/laincloud/sso/ssolib/models/testbackend"
	"github.com/laincloud/sso/ssolib/models/testhelper"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	testBack = &testbackend.TestBackend{}
)

type TestHelper struct {
	T   *testing.T
	Ctx context.Context
}

func NewTestHelper(t *testing.T) *TestHelper {
	ctx := context.Background()
	t.Log(testBack)
	testBack = &testbackend.TestBackend{}
	t.Log(testBack)
	mctx := testhelper.NewTestHelper(t).Ctx
	mctx.Back = testBack
	ctx = context.WithValue(ctx, "mctx", mctx)
	ctx = context.WithValue(ctx, "emailSuffix", "@example.com")
	ctx = context.WithValue(ctx, "userBackend", testBack)

	app.InitDatabase(mctx)
	group.EnableNestedGroup()
	group.InitDatabase(mctx)
	oauth2.InitDatabase(mctx)

	th := &TestHelper{
		T:   t,
		Ctx: ctx,
	}
	return th
}

func (th *TestHelper) login(username string) {
	th.loginWithScope(username, "")
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

type mock struct {
	restore func()
}

func mockParams(th *TestHelper, pairs map[string]string) *mock {
	origParams := params
	params = func(ctx context.Context, key string) string {
		value, ok := pairs[key]
		assert.True(th.T, ok, key)
		return value
	}
	return &mock{func() { params = origParams }}
}

func mockReverse(th *TestHelper, url string) *mock {
	origReverse := reverse
	reverse = func(s *Server, route string, params ...interface{}) string {
		return url
	}
	return &mock{func() { reverse = origReverse }}
}

// for test only
type Mock struct {
	Restore func()
}

// for test only
func MockParams(t *testing.T, pairs map[string]string) *Mock {
	origParams := params
	params = func(ctx context.Context, key string) string {
		value, ok := pairs[key]
		assert.True(t, ok, key)
		return value
	}
	return &Mock{func() { params = origParams }}
}

// for test only
func MockReverse(t *testing.T, url string) *Mock {
	origReverse := reverse
	reverse = func(s *Server, route string, params ...interface{}) string {
		return url
	}
	return &Mock{func() { reverse = origReverse }}
}
