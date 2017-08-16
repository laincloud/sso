package user

import (
	"bytes"
	"testing"

	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/testhelper"
	"golang.org/x/crypto/bcrypt"
)

var testBack iuser.UserBackend

func init() {
	generateHashFromPassword = mockedGenerateHashFromPassword
	compareHashAndPassword = mockedCompareHashAndPassword
}

func NewTestHelper(t *testing.T) testhelper.TestHelper {
	th := testhelper.NewTestHelper(t)
	mctx := th.Ctx
	testBack = New(testhelper.GetTestMysqlDSN(), "admin@example.com", "admin")
	mctx.Back = testBack
	testBack.(*UserBack).InitDatabase()
	return th
}

func mockedGenerateHashFromPassword(password []byte, cost int) ([]byte, error) {
	return password, nil
}

func mockedCompareHashAndPassword(hashedPassword, password []byte) error {
	if bytes.Equal(hashedPassword, password) {
		return nil
	} else {
		return bcrypt.ErrMismatchedHashAndPassword
	}
}
