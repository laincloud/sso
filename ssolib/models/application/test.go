package application

import (
"testing"

"github.com/laincloud/sso/ssolib/models/testhelper"
)

func NewTestHelper(t *testing.T) testhelper.TestHelper {
	th := testhelper.NewTestHelper(t)
	InitDatabase(th.Ctx)
	return th
}

