package ssolib

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/testbackend"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestPostGroupsShouldCreateANewGroup(t *testing.T) {
	th := NewTestHelper(t)

	code, resp := createTestUserAndCallPostGroups(th)
	assert.Equal(t, code, http.StatusCreated, fmt.Sprint(resp))
	g, ok := resp.(*Group)
	assert.True(t, ok)
	assert.Equal(t, "testgroup", g.Name)
	assert.Equal(t, "Test Group", g.FullName)
}

func TestPostGroupShouldCreateANewGroupWithCurrentUserAsAdmin(t *testing.T) {
	th := NewTestHelper(t)

	code, resp := createTestUserAndCallPostGroups(th)
	assert.Equal(t, code, http.StatusCreated)
	g, ok := resp.(*Group)
	assert.True(t, ok)

	r, _ := http.NewRequest("GET",
		fmt.Sprintf("http://sso.example.com/api/groups/%s", g.Name), nil)

	// mock server.Params
	origParams := params
	params = func(ctx context.Context, key string) string {
		assert.Equal(t, key, "groupname")
		return "testgroup"
	}
	defer func() {
		params = origParams
	}()

	code, resp = GroupResource{}.Get(th.Ctx, r)
	t.Log(code)
	t.Log(resp)
	assert.Equal(t, code, http.StatusOK, fmt.Sprint(resp))
	gwm, ok := resp.(*GroupWithMembers)
	assert.True(t, ok)
	assert.Equal(t, "testgroup", gwm.Name)
	assert.Equal(t, "Test Group", gwm.FullName)
	assert.Equal(t, []MemberRole{MemberRole{Name: "testuser", Role: "admin"}}, gwm.Members)
}

func TestPostGroupsShouldFailWith401WhenOAuth2ScopeHasNoWriteGroups(t *testing.T) {
	// FIXME: complete the test
	th := NewTestHelper(t)

	createTestUser(th, "testuser")
	th.loginWithScope("testuser", "")

	code, _ := callPostGroups(th)
	assert.Equal(t, http.StatusUnauthorized, code)
}

func TestDeleteGroupShouldWork(t *testing.T) {
	th := NewTestHelper(t)

	code, resp := createTestUserAndCallPostGroups(th)
	assert.Equal(t, http.StatusCreated, code)
	groupname := resp.(*Group).Name

	th.loginWithScope("testuser", "write:group")

	code, resp = callDeleteGroup(th, groupname)
	assert.Equal(t, http.StatusNoContent, code)

	code, resp = callGetGroup(th, groupname)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestPutGroupMemberShouldWork(t *testing.T) {
	th := NewTestHelper(t)

	code, resp := createTestUserAndCallPostGroups(th)
	assert.Equal(t, http.StatusCreated, code)
	groupname := resp.(*Group).Name

	createTestUser(th, "testuser2")
	th.loginWithScope("testuser", "write:group")
	code, resp = callPutGroupMember(th, groupname, "testuser2")
	assert.Equal(t, http.StatusOK, code)

	code, resp = callGetGroup(th, groupname)
	assert.Equal(t, http.StatusOK, code)
	gwm, ok := resp.(*GroupWithMembers)
	assert.True(t, ok)
	assert.Equal(t, []MemberRole{
		MemberRole{Name: "testuser", Role: "admin"},
		MemberRole{Name: "testuser2"},
	}, gwm.Members)
}

func TestDeleteGroupMemberShouldWork(t *testing.T) {
	th := NewTestHelper(t)

	code, resp := createTestUserAndCallPostGroups(th)
	assert.Equal(t, http.StatusCreated, code)
	groupname := resp.(*Group).Name

	th.loginWithScope("testuser", "write:group")
	createTestUser(th, "testuser2")
	code, resp = callPutGroupMember(th, groupname, "testuser2")
	assert.Equal(t, http.StatusOK, code)

	code, resp = callDeleteGroupMember(th, groupname, "testuser2")
	assert.Equal(t, http.StatusNoContent, code)

	code, resp = callGetGroup(th, groupname)
	assert.Equal(t, http.StatusOK, code)
	gwm, ok := resp.(*GroupWithMembers)
	assert.True(t, ok)
	assert.Equal(t, []MemberRole{
		MemberRole{Name: "testuser", Role: "admin"},
	}, gwm.Members)
}

func TestPutMemberGroupShouldWork(t *testing.T) {
	group.EnableNestedGroup()
	th := NewTestHelper(t)

	code, resp := createTestUserAndCallPostGroups(th)
	assert.Equal(t, http.StatusCreated, code)
	groupname := resp.(*Group).Name

	t.Log(groupname)
	th.loginWithScope("testuser", "write:group")
	createTestUser(th, "testuser2")
	th.loginWithScope("testuser2", "write:group")
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/groups",
		strings.NewReader(`{"name": "songroup", "fullname": "Test Group"}`))
	code2, resp2 := GroupsResource{}.Post(th.Ctx, r)
	th.logout()

	assert.Equal(t, http.StatusCreated, code2)

	sonname := resp2.(*Group).Name

	th.loginWithScope("testuser", "write:group")
	url := "http://sso.example.com/api/groups/" + groupname + "/group-members/" + sonname
	r, _ = http.NewRequest("PUT", url, strings.NewReader(`{"role":"admin"}`))
	aMock := mockParams(th, map[string]string{
		"groupname": groupname,
		"sonname":   sonname})
	defer aMock.restore()
	code3, resp3 := GroupMemberResource{}.Put(th.Ctx, r)
	assert.Equal(t, http.StatusOK, code3)
	assert.Equal(t, "group member added", resp3.(string))
	t.Log(resp3)

	r, _ = http.NewRequest("PUT", url, strings.NewReader(`{"role":"normal"}`))
	code4, resp4 := GroupMemberResource{}.Put(th.Ctx, r)
	assert.Equal(t, http.StatusOK, code4)
	assert.Equal(t, "group member added", resp4.(string))

}

func TestDeleteMemberGroupShouldWork(t *testing.T) {
	group.EnableNestedGroup()
	th := NewTestHelper(t)

	code, resp := createTestUserAndCallPostGroups(th)
	assert.Equal(t, http.StatusCreated, code)
	groupname := resp.(*Group).Name

	t.Log(groupname)
	th.loginWithScope("testuser", "write:group")
	createTestUser(th, "testuser2")
	th.loginWithScope("testuser2", "write:group")
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/groups",
		strings.NewReader(`{"name": "songroup", "fullname": "Test Group"}`))
	code2, resp2 := GroupsResource{}.Post(th.Ctx, r)
	th.logout()

	assert.Equal(t, http.StatusCreated, code2)

	sonname := resp2.(*Group).Name

	th.loginWithScope("testuser", "write:group")
	url := "http://sso.example.com/api/groups/" + groupname + "/group-members/" + sonname
	r, _ = http.NewRequest("PUT", url, strings.NewReader(`{"role":"admin"}`))
	aMock := mockParams(th, map[string]string{
		"groupname": groupname,
		"sonname":   sonname})
	defer aMock.restore()
	code3, resp3 := GroupMemberResource{}.Put(th.Ctx, r)
	assert.Equal(t, http.StatusOK, code3)
	assert.Equal(t, "group member added", resp3.(string))
	t.Log(resp3)

	th.logout()
	th.loginWithScope("testuser2", "write:group")
	r, _ = http.NewRequest("DELETE", url, nil)
	code4, resp4 := GroupMemberResource{}.Delete(th.Ctx, r)
	assert.Equal(t, http.StatusNoContent, code4)
	assert.Equal(t, "group member deleted", resp4)

	th.logout()
	th.loginWithScope("testuser", "write:group")
	r, _ = http.NewRequest("PUT", url, strings.NewReader(`{"role":"normal"}`))
	code5, resp5 := GroupMemberResource{}.Put(th.Ctx, r)
	assert.Equal(t, http.StatusOK, code5)
	assert.Equal(t, "group member added", resp5.(string))

	th.logout()
	th.loginWithScope("testuser2", "write:group")
	r, _ = http.NewRequest("DELETE", url, nil)
	code4, resp4 = GroupMemberResource{}.Delete(th.Ctx, r)
	assert.Equal(t, http.StatusForbidden, code4)
	t.Log(resp4)
}

func createTestUser(th *TestHelper, username string) iuser.User {
	//mctx := getModelContext(th.Ctx)
	err := testBack.CreateUser(&testbackend.TestUser{
		Name: username,
		//		FullName:     username,
		PasswordHash: []byte("test"),
	}, true)
	assert.Nil(th.T, err)
	u, err := testBack.GetUserByName(username)
	assert.Nil(th.T, err)
	return u
}

func callPostGroups(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/groups",
		strings.NewReader(`{"name": "testgroup", "fullname": "Test Group"}`))
	return GroupsResource{}.Post(th.Ctx, r)
}

func callPostGroupsWithGroupName(th *TestHelper, groupname string) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/groups",
		strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, groupname)))
	return GroupsResource{}.Post(th.Ctx, r)
}

func createTestUserAndCallPostGroups(th *TestHelper) (int, interface{}) {
	createTestUser(th, "testuser")
	th.loginWithScope("testuser", "write:group")
	code, resp := callPostGroups(th)
	th.logout()
	return code, resp
}

func callGetGroup(th *TestHelper, groupname string) (int, interface{}) {
	r, _ := http.NewRequest("GET",
		fmt.Sprintf("http://sso.example.com/api/groups/%s", groupname), nil)

	aMock := mockParams(th, map[string]string{"groupname": groupname})
	defer aMock.restore()

	return GroupResource{}.Get(th.Ctx, r)
}

func callDeleteGroup(th *TestHelper, groupname string) (int, interface{}) {
	r, _ := http.NewRequest("DELETE",
		fmt.Sprintf("http//sso.example.com/api/groups/%s", groupname), nil)

	aMock := mockParams(th, map[string]string{"groupname": groupname})
	defer aMock.restore()

	return GroupResource{}.Delete(th.Ctx, r)
}

func callPutGroupMember(th *TestHelper, groupname, username string) (int, interface{}) {
	r, _ := http.NewRequest("PUT",
		fmt.Sprintf("http://sso.example.com/api/groups/%s/members/%s",
			groupname, username),
		strings.NewReader(`{"role": "normal"}`))
	aMock := mockParams(th, map[string]string{
		"groupname": groupname,
		"username":  username})
	defer aMock.restore()

	return MemberResource{}.Put(th.Ctx, r)
}

func callDeleteGroupMember(th *TestHelper, groupname, username string) (int, interface{}) {
	r, _ := http.NewRequest("DELETE",
		fmt.Sprintf("http://sso.example.com/api/groups/%s/members/%s",
			groupname, username),
		nil)
	aMock := mockParams(th, map[string]string{
		"groupname": groupname,
		"username":  username})
	defer aMock.restore()

	return MemberResource{}.Delete(th.Ctx, r)
}
