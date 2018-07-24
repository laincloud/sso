package ssolib

import (
	"testing"

	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/oauth2"
	"github.com/laincloud/sso/ssolib/models/testbackend"
	"github.com/laincloud/sso/ssolib/models/testhelper"
	"github.com/laincloud/sso/ssolib/models/application"

	"golang.org/x/net/context"
	"github.com/laincloud/sso/ssolib/models/role"
)



func NewTestHelper2(t *testing.T) *TestHelper {
	ctx := context.Background()
	t.Log(testBack)
	testBack = &testbackend.TestBackend{}
	t.Log(testBack)
	mctx := testhelper.NewTestHelper2(t).Ctx
	mctx.Back = testBack
	ctx = context.WithValue(ctx, "mctx", mctx)
	ctx = context.WithValue(ctx, "emailSuffix", "@example.com")
	ctx = context.WithValue(ctx, "userBackend", testBack)

	app.InitDatabase(mctx)
	group.EnableNestedGroup()
	group.InitDatabase(mctx)
	oauth2.InitDatabase(mctx)
	application.InitDatabase(mctx)
	role.InitDatabase(mctx)

	th := &TestHelper{
		T:   t,
		Ctx: ctx,
	}
	return th
}
