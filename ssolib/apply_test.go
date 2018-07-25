package ssolib

import (
	"testing"
	"net/http"
	"strings"
	"github.com/laincloud/sso/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/testbackend"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/application"
	"strconv"
)

func TestApply_Post(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserAndGroup(th)
	th.login("testuser")
	code ,resp := callPostApplication(th)
	assert.Equal(t, code, http.StatusOK)
	a, ok := resp.([]application.Application)
	assert.True(t, ok)
	assert.Equal(t, "testuser@creditease.cn", a[0].ApplicantEmail)
}

func createTestUserAndGroup(th *TestHelper) iuser.User {
	mctx := getModelContext(th.Ctx)
	admin := createTestUserWithEmail(th, "testadmin")
	u := createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testadmin", "write:group")
	callPostGroupsWithGroupName(th, "group1")
	g1 ,_:= group.GetGroupByName(mctx, "group1")
	g1.AddMember(mctx, admin, 1)
	callPostGroupsWithGroupName(th, "group2")
	g2 ,_:= group.GetGroupByName(mctx, "group2")
	g2.AddMember(mctx, admin, 1)
	callPostGroupsWithGroupName(th, "lain")
	g3 ,_:= group.GetGroupByName(mctx, "lain")
	g3.AddMember(mctx, admin, 1)
	return u
}


//create two users and three groups. one user is testuser, and the other is testadmin.
//Two normal groups and one lain groups.
//testadmin is the admin of three groups.
//testuser will apply to join in two normal groups.
func createTestUserWithEmail(th *TestHelper, username string) iuser.User {
	err := testBack.CreateUser(&testbackend.TestUser{
		Name: username,
		PasswordHash: []byte("test"),
		Email: username+"@creditease.cn",
	}, true)
	assert.Nil(th.T, err)
	u, err := testBack.GetUserByName(username)
	assert.Nil(th.T, err)
	return u
}


func callPostApplication(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/applications",
		strings.NewReader(`{"target_type": "group", "reason":"testing", "target": [{"name": "group1","role":"normal"}, {"name": "group2","role":"normal"}]}`))
	return Apply{}.Post(th.Ctx, r)
}

func TestApply_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserAndGroup(th)
	th.login("testuser")
	callPostApplication(th)
	code1, resp1 := callGetApplication(th)
	assert.Equal(t, http.StatusOK, code1)
	a, ok := resp1.([]application.Application)
	assert.True(t, ok)
	assert.Equal(t, "testuser@creditease.cn", a[0].ApplicantEmail)
	th.logout()
	th.login("testadmin")
	//lain member can see others' application
	code2, resp2 :=callGetApplication(th)
	assert.Equal(t, http.StatusOK, code2)
	b, ok2 := resp2.([]application.Application)
	assert.True(t, ok2)
	assert.Equal(t, 2 , len(b))
	//test admin's getting applications
	code3, resp3 := callGetApplicationByCommit(th)
	assert.Equal(t, http.StatusOK, code3)
	c, ok3 := resp3.([]application.Application)
	assert.True(t, ok3)
	assert.Equal(t, 2, len(c))
	code4, resp4 := callPostApplicationHandle(th, strconv.Itoa(1),"approve")
	assert.Equal(t, http.StatusOK, code4)
	d, ok4 := resp4.(*application.Application)
	assert.True(t, ok4)
	assert.Equal(t, "approved", d.Status)
	//test getting by status
	code5, resp5 := callGetApplicationByStatus(th)
	assert.Equal(t, http.StatusOK, code5)
	e, ok5 := resp5.([]application.Application)
	assert.True(t, ok5)
	assert.Equal(t, 1 , len(e))
	assert.Equal(t, "approved", e[0].Status)
	//test getting by time
	//get 2nd application which has a initialled status
	code6,resp6 := callGetApplicationByTime(th)
	assert.Equal(t, http.StatusOK, code6)
	f, ok6 := resp6.([]application.Application)
	assert.True(t, ok6)
	assert.Equal(t, 1 , len(f))
	assert.Equal(t, "initialled", f[0].Status)
}

func callGetApplication(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/applications?"+"applicant_email=testuser@creditease.cn", nil)
	return Apply{}.Get(th.Ctx, r)
}

func callGetApplicationByTime(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/applications?"+"from=1&to=1", nil)
	return Apply{}.Get(th.Ctx, r)
}

func callGetApplicationByCommit(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/applications?"+"commit_email=testadmin@creditease.cn", nil)
	return Apply{}.Get(th.Ctx, r)
}

func callGetApplicationByStatus(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/applications?"+"applicant_email=testuser@creditease.cn&status=approved", nil)
	return Apply{}.Get(th.Ctx, r)
}

func TestApplicationHandle_Post(t *testing.T) {
	th := NewTestHelper(t)
	mctx := getModelContext(th.Ctx)
	user := createTestUserAndGroup(th)
	callPostApplication(th)
	th.logout()
	th.login("testadmin")
	code, resp := callPostApplicationHandle(th, strconv.Itoa(1),"approve")
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.(*application.Application)
	assert.True(t, ok)
	assert.Equal(t, "approved", a.Status)
	gs ,_ := group.GetGroupRolesOfUser(mctx, user)
	assert.Equal(t, 1, len(gs))
	th.logout()
	th.login("testuser")
	code, resp = callPostApplicationHandle(th, strconv.Itoa(2),"recall")
	assert.Equal(t, http.StatusNoContent, code)
	code, resp = callGetApplication(th)
	assert.Equal(t, http.StatusOK, code)
	b, ok := resp.([]application.Application)
	assert.True(t, ok)
	assert.Equal(t, 1, len(b))

}

func callPostApplicationHandle (th *TestHelper, id string, action string) (int, interface{}) {
	url := "http://sso.example.com/api/applications/" + id + "?action=" + action
	r, _ := http.NewRequest("POST", url, nil)
	aMock := mockParams(th, map[string]string{
		"application_id": id})
	defer aMock.restore()
	return ApplicationHandle{}.Post(th.Ctx, r)
}