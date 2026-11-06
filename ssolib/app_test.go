package ssolib

import (
	"testing"
	"net/http"
	"strings"
	"github.com/laincloud/sso/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestAppResource_Put(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app")
	createApp(th)
	code, resp := updateApp(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*App)
	assert.True(t, ok)
	assert.Equal(t, "app2", a.FullName)
}


func TestAppResource_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app read:app")
	createApp(th)
	code, resp := getApp(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*App)
	assert.True(t, ok)
	assert.Equal(t, "app1", a.FullName)
}

func TestAppsResource_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app read:app")
	code, resp := createApp(th)
	assert.Equal(t, http.StatusCreated, code)
	a, ok := resp.(*App)
	assert.True(t, ok)
	assert.Equal(t, "app1", a.FullName)
}

func TestAppsResource_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app read:app")
	createApp(th)
	code, resp := getApps(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]App)
	assert.True(t, ok)
	assert.Equal(t, "app1", a[0].FullName)
}

func updateApp(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("PUT", "http://sso.example.com/api/app/1",
		strings.NewReader(`{"fullname": "app2", "redirect_uri": "https://example2.com"}`))
	aMock := mockParams(th, map[string]string{
		"id":   "1"})
	defer aMock.restore()
	return AppResource{}.Put(th.Ctx, r)
}

func getApp(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/app/1",
		nil)
	aMock := mockParams(th, map[string]string{
		"id":   "1"})
	defer aMock.restore()
	return AppResource{}.Get(th.Ctx, r)
}

func getApps(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/apps",
		nil)
	return AppsResource{}.Get(th.Ctx, r)
}

func createApp(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/apps",
		strings.NewReader(`{"fullname": "app1", "redirect_uri": "https://example.com"}`))
	return AppsResource{}.Post(th.Ctx, r)
}

func createApp2(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/apps",
		strings.NewReader(`{"fullname": "app2", "redirect_uri": "https://example.com"}`))
	return AppsResource{}.Post(th.Ctx, r)
}