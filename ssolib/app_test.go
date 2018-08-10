package ssolib

import (
	"testing"
	"net/http"
	"strings"
	"github.com/laincloud/sso/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/laincloud/sso/ssolib/models/app"
)

func TestAppResource_Put(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app")
	g,_ := createApp(th)
	t.Log(g)
	code, resp := updateApp(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*App)
	assert.True(t, ok)
	assert.Equal(t, "app2", a.FullName)
}

func updateApp(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("PUT", "http://sso.example.com/api/app/1",
		strings.NewReader(`{"fullname": "app2", "redirect_uri": "https://example2.com"}`))
	th.T.Log(r)
	aMock := mockParams(th, map[string]string{
		"id":   "1"})
	defer aMock.restore()
	return AppResource{}.Put(th.Ctx, r)
}

func TestAppInformation_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app")
	createApp(th)
	code, resp := getappinfo(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([] app.AppInfo)
	assert.True(t, ok)
	assert.Equal(t,1, len(a))

}

func getappinfo(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/app_info", nil)
	return AppInformation{}.Get(th.Ctx, r)

}
