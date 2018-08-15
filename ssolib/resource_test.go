package ssolib

import (
	"testing"
	"net/http"
	"strings"
	"github.com/stretchr/testify/assert"
	"github.com/laincloud/sso/ssolib/models/role"
	"strconv"
)

func TestResourceResource_Delete(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:resource read:resource write:role")
	createApp(th)
	createResources(th)
	code, _ := deleteResource(th)
	assert.Equal(t, http.StatusNoContent, code)

}

func TestResourceResource_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:resource read:resource write:role")
	createApp(th)
	createResources(th)
	code, resp := getResource(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*role.Resource)
	assert.True(t, ok)
	assert.Equal(t, "resource 1", (*a).Name)
}


func TestResourceResource_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:resource read:resource write:role")
	createApp(th)
	createResources(th)
	code, resp:= updateResource(th)
	assert.Equal(t, code, http.StatusOK)
	a, ok := resp.(*role.Resource)
	assert.True(t, ok)
	assert.Equal(t, "resource1Updated", (*a).Name)
}

func TestResourcesResource_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:resource read:resource write:role")
	createApp(th)
	createResources(th)
	//test type=raw
	code, resp := getResources(th, "raw")
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]role.Resource)
	assert.True(t, ok)
	assert.Equal(t, 2, len(a))
	//test type=byrole
	createRole1(th)
	createRole2(th)
	createRole3(th)
	addRoleResource(th, 2)
	code, resp = getResources(th, "byrole")
	assert.Equal(t, http.StatusOK, code)
	b, ok := resp.([]role.RoleResources)
	assert.True(t, ok)
	assert.Equal(t, 4, len(b))
	//test type=byapp
	th.logout()
	createTestUserWithEmail(th, "testuser2")
	th.loginWithScope("testuser2", "read:resource")
	code, resp = getResources(th, "byapp")
	assert.Equal(t, http.StatusOK, code)
	a, _ = resp.([]role.Resource)
	assert.Equal(t, 0, len(a))
}

func TestResourcesResource_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:resource read:resource write:role")
	createApp(th)
	//test create resources
	code, _:= createResources(th)
	assert.Equal(t, http.StatusOK, code)
	code, resp := getResources(th, "raw")
	assert.Equal(t, code, http.StatusOK)
	a, ok := resp.([]role.Resource)
	assert.True(t, ok)
	assert.Equal(t, 2, len(a))
	//test delete resources
	code, _ = deleteResources(th)
	assert.Equal(t, http.StatusNoContent, code)
	code, resp = getResources(th, "raw")
	assert.Equal(t, code, http.StatusOK)
	a, _ = resp.([]role.Resource)
	assert.Equal(t, 0, len(a))
}

func TestRoleResourceResource_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app write:resource read:resource write:role")
	createApp(th)
	createResources(th)
	//test add
	code, _ := addRoleResource(th, 1)
	assert.Equal(t, http.StatusOK, code)
	code, resp := getResources(th, "byrole")
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]role.RoleResources)
	assert.True(t, ok)
	assert.Equal(t, "resource 1", a[0].Resources[0].Name)
	assert.Equal(t, 2, len(a[0].Resources))
	//test update
	code, _ = updateRoleResource(th, 1)
	assert.Equal(t, http.StatusOK, code)
	code, resp = getResources(th, "byrole")
	assert.Equal(t, http.StatusOK, code)
	a, ok = resp.([]role.RoleResources)
	assert.True(t, ok)
	assert.Equal(t, "resource 1", a[0].Resources[0].Name)
	assert.Equal(t, 1, len(a[0].Resources))
	//test delete
	code, _ = deleteRoleResource(th, 1)
	assert.Equal(t, http.StatusNoContent, code)
	code, resp = getResources(th, "byrole")
	assert.Equal(t, http.StatusOK, code)
	a, _ = resp.([]role.RoleResources)
	assert.Equal(t, 0, len(a[0].Resources))
}

func createResources(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/resources?app_id=1&action=add",
		strings.NewReader(`{"resources" : [{"name":"resource 1","data":"test","description":"test"},
    {"name": "resource 2", "data": "test2", "description":"test2"}]}`))
	return ResourcesResource{}.Post(th.Ctx, r)
}

func updateResource(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/resources/1",
		strings.NewReader(`{"name": "resource1Updated", "data": "test data", "description":"test desc"}`))
	aMock := mockParams(th, map[string]string{
		"id":   "1"})
	defer aMock.restore()
	return ResourceResource{}.Post(th.Ctx, r)
}

func updateResources(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/resources?app_id=1&action=update",
		strings.NewReader(`{"resources":[{"id": 1, name": "resource1updated", "data": "test data1", "description":"test desc1"},
    {"id": 2, name": "resource2updated", "data": "test data2", "description":"test desc2"}]}`))
	return ResourcesResource{}.Post(th.Ctx, r)
}

func deleteResource(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("DELETE", "http://sso.example.com/api/resources/1",
		nil)
	aMock := mockParams(th, map[string]string{
		"id":   "1"})
	defer aMock.restore()
	return ResourceResource{}.Delete(th.Ctx, r)
}

func deleteResources(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/resources?app_id=1&action=delete",
		strings.NewReader(`{"resources":[{"id": 1}, {"id": 2}]}`))
	return ResourcesResource{}.Post(th.Ctx, r)
}

func getResources(th *TestHelper, t string) (int, interface{}){
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/resources?app_id=1&type=" + t,
		nil)
	return ResourcesResource{}.Get(th.Ctx, r)
}

func addRoleResource(th *TestHelper, role int) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/roles/" + strconv.Itoa(role)+ "/resources",
		strings.NewReader(`{"resource_list":[1,2],"action":"add"}`))
	aMock := mockParams(th, map[string]string{
		"id":   strconv.Itoa(role)})
	defer aMock.restore()
	return RoleResourceResource{}.Post(th.Ctx, r)
}

func updateRoleResource(th *TestHelper, role int) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/roles/" + strconv.Itoa(role)+ "/resources",
		strings.NewReader(`{"resource_list":[1],"action":"update"}`))
	aMock := mockParams(th, map[string]string{
		"id":   strconv.Itoa(role)})
	defer aMock.restore()
	return RoleResourceResource{}.Post(th.Ctx, r)
}

func deleteRoleResource(th *TestHelper, role int) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/roles/" + strconv.Itoa(role)+ "/resources",
		strings.NewReader(`{"resource_list":[1],"action":"delete"}`))
	aMock := mockParams(th, map[string]string{
		"id":   strconv.Itoa(role)})
	defer aMock.restore()
	return RoleResourceResource{}.Post(th.Ctx, r)
}

func getResource(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/resources/1",
		nil)
	aMock := mockParams(th, map[string]string{
		"id":   "1"})
	defer aMock.restore()
	return ResourceResource{}.Get(th.Ctx, r)
}
