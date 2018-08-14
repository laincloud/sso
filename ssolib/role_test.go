package ssolib

import (
	"testing"
	"net/http"
	"github.com/laincloud/sso/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"strings"
	"github.com/laincloud/sso/ssolib/models/role"
	"github.com/laincloud/sso/ssolib/models/app"
)

func TestRolesResource_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app read:role")
	createApp(th)
	createRootRole(th)
	code ,resp := getRole(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]role.Role)
	assert.True(t, ok)
	assert.Equal(t, 1, len(a))
}

func TestRolesResource_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role")
	createApp(th)
	createRootRole(th)
	code, resp := createRole1(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*role.Role)
	assert.True(t, ok)
	assert.Equal(t, "roleOne", a.Name)
}

func TestAppRoleResource_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role read:app")
	createApp(th)
	createApp2(th)
	createRootRole(th)
	createRootRole2(th)
	createRole1(th)
	code, resp := getAppRole(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]AppRolesOfUser)
	assert.True(t, ok)
	assert.Equal(t, 2, len(a))
}

func TestAppRoleResource_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role read:app")
	createApp(th)
	code, resp:= createRootRole(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*app.App)
	assert.True(t, ok)
	assert.Equal(t, 1, a.AdminRoleId)
}

func TestAppRoleResource_Delete(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role read:app")
	createApp(th)
	createRootRole(th)
	code, resp := deleteRootRole(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*app.App)
	assert.True(t, ok)
	assert.Equal(t, -1, a.AdminRoleId)
}

func TestRoleResource_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role read:app")
	createApp(th)
	createRootRole(th)
	createRole1(th)
	code, resp := updateRole1(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*role.Role)
	assert.True(t, ok)
	assert.Equal(t, "roleOneupdated", a.Name)
}

func TestRoleResource_Delete(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role read:app")
	createApp(th)
	createRootRole(th)
	createRole1(th)
	code, resp := deleteRole1(th)
	assert.Equal(t, http.StatusNoContent, code)
	_, ok := resp.(*role.Role)
	assert.Equal(t, false, ok)
}

func TestRoleMemberResource_Put(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role read:app")
	createApp(th)
	createRootRole(th)
	createTestUserWithEmail(th, "testuser2")
	code, _ := addMember(th)
	th.logout()
	th.loginWithScope("testuser2", "read:app")
	assert.Equal(t, http.StatusOK, code)
	code, resp := getAppRole(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]AppRolesOfUser)
	assert.True(t, ok)
	assert.Equal(t, 1, len(a[0].Roles))
}

func TestRoleMemberResource_Delete(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:role read:app")
	createApp(th)
	createRootRole(th)
	code, _ := deleteMember(th)
	assert.Equal(t, http.StatusNoContent, code)
	code, _ = getAppRole(th)
	assert.Equal(t, http.StatusNotFound, code)
}

func getRole(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/roles?app_id=1&all=true", nil)
	return RolesResource{}.Get(th.Ctx, r)
}

func createRootRole(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/app_roles",
		strings.NewReader(`{"app_id": 1, "role_name": "role1"}`))
	return AppRoleResource{}.Post(th.Ctx, r)
}

func createRootRole2(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/app_roles",
		strings.NewReader(`{"app_id": 2, "role_name": "role2"}`))
	return AppRoleResource{}.Post(th.Ctx, r)
}

func createRole1(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/roles",
		strings.NewReader(`{"app_id": 1, "name": "roleOne", "description":"test desc", "parent_id":1 }`))
	return RolesResource{}.Post(th.Ctx, r)
}

func createRole2(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/roles",
		strings.NewReader(`{"app_id": 1, "name": "roleTwo", "description":"test desc", "parent_id":1 }`))
	return RolesResource{}.Post(th.Ctx, r)
}

func createRole3(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/roles",
		strings.NewReader(`{"app_id": 1, "name": "roleThree", "description":"test desc", "parent_id":1 }`))
	return RolesResource{}.Post(th.Ctx, r)
}

func getAppRole(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/app_roles",
		nil)
	return AppRoleResource{}.Get(th.Ctx, r)
}

func deleteRootRole(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("DELETE", "http://sso.example.com/api/app_roles",
		strings.NewReader(`{"app_id": 1}`))
	return AppRoleResource{}.Delete(th.Ctx, r)
}

func updateRole1(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/roles/2",
		strings.NewReader(`{"name":"roleOneupdated", "description":"test", "parent_id":1}`))
	aMock := mockParams(th, map[string]string{
		"id":   "2"})
	defer aMock.restore()
	return RoleResource{}.Post(th.Ctx, r)
}

func deleteRole1(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("DELETE", "http://sso.example.com/api/roles/2",
		nil)
	aMock := mockParams(th, map[string]string{
		"id":   "2"})
	defer aMock.restore()
	return RoleResource{}.Delete(th.Ctx, r)
}

func addMember(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("PUT", "http://sso.example.com/api/roles/1/members/testuser2",
		strings.NewReader(`{"type":"admin"}`))
	aMock := mockParams(th, map[string]string{
		"id":   "1", "username": "testuser2"})
	defer aMock.restore()
	return RoleMemberResource{}.Put(th.Ctx, r)
}

func deleteMember(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("DELETE", "http://sso.example.com/api/roles/1/members/testuser",
		nil)
	aMock := mockParams(th, map[string]string{
		"id":   "1", "username": "testuser"})
	defer aMock.restore()
	return RoleMemberResource{}.Delete(th.Ctx, r)
}

