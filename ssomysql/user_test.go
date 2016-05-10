package ssomysql

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib"
	//	"github.com/laincloud/sso/ssomysql/user"
)

func TestPostUsersShouldReturn202WhenRequestIsValid(t *testing.T) {
	th := NewTestHelper(t)

	//var w *httptest.ResponseRecorder
	w := httptest.NewRecorder()
	callPostUsers(th, "test", "test@example.com", w)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestDeleteUserShouldReturn403WhenCurrentUserIsNotAdmin(t *testing.T) {
	th := NewTestHelper(t)

	createTestUser(th, "testuser1")
	th.loginWithScope("testuser1", "write:group")

	code, _ := callPostGroupsWithGroupName(th, "admins")
	assert.Equal(t, http.StatusCreated, code)

	code, _ = callDeleteGroupMember(th, "admins", "testuser1")

	// I'm not an admin now, so it will fail if I delete myself
	code, resp := callDeleteUser(th, "testuser1")
	assert.Equal(t, http.StatusForbidden, code, resp)
}

func TestGetMeShouldReturn401WhenUserIsNotLoggedIn(t *testing.T) {
	th := NewTestHelper(t)

	createTestUser(th, "testuser1")

	th.logout()

	code, _ := callGetMe(th)
	assert.Equal(t, http.StatusUnauthorized, code)
}

func callPostUsers(th *TestHelper, username, email string, w *httptest.ResponseRecorder) context.Context {
	mctx := getModelContext(th.Ctx)
	origSMTP := mctx.SMTPAddr
	mctx.SMTPAddr = ""
	defer func() { mctx.SMTPAddr = origSMTP }()

	r, _ := http.NewRequest("POST",
		"http://sso.example.com/api/users",
		strings.NewReader(fmt.Sprintf(`{"name": "%s", "email": "%s", "password": "%s"}`,
			username, email, "testpassword")))
	return UsersPost(th.Ctx, w, r)
}

func callDeleteUser(th *TestHelper, username string) (int, interface{}) {
	r, _ := http.NewRequest("DELETE",
		fmt.Sprintf("http://sso.example.com/api/users/%s", username), nil)
	aMock := ssolib.MockParams(th.T, map[string]string{"username": username})
	defer aMock.Restore()

	return ssolib.UserResource{}.Delete(th.Ctx, r)
}

func callGetMe(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/me", nil)
	return ssolib.MeResource{}.Get(th.Ctx, r)
}
