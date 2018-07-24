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
)

func TestApply_Post(t *testing.T) {
	th := NewTestHelper2(t)
	createTestUserAndGroup(th)
	code ,resp := callPostApplication(th)
	assert.Equal(t, code, http.StatusOK)
	a, ok := resp.([]application.Application)
	assert.True(t, ok)
	assert.Equal(t, "testuser@creditease.cn", a[0].ApplicantEmail)
}

func createTestUserAndGroup(th *TestHelper)  {
	mctx := getModelContext(th.Ctx)
	admin := createTestUserWithEmail(th, "testadmin")
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testadmin", "write:group")
	callPostGroupsWithGroupName(th, "group1")
	g1 ,_:= group.GetGroupByName(mctx, "group1")
	g1.AddMember(mctx, admin, 1)
	callPostGroupsWithGroupName(th, "group2")
	g2 ,_:= group.GetGroupByName(mctx, "group2")
	g2.AddMember(mctx, admin, 1)
	th.logout()
	th.login("testuser")
}

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
	th := NewTestHelper2(t)
	createTestUserAndGroup(th)
	callPostApplication(th)
	code, resp := callGetApplication(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]application.Application)
	assert.True(t, ok)
	assert.Equal(t, "testuser@creditease.cn", a[0].ApplicantEmail)
	th.logout()
	th.login("testadmin")
	code, resp = callGetApplication(th)
	assert.Equal(t, http.StatusBadRequest, code)
	code, resp = callGetApplicationByCommit(th)
	assert.Equal(t, http.StatusOK, code)
	b, ok2 := resp.([]application.Application)
	assert.True(t, ok2)
	assert.Equal(t, 2, len(b))
}

func callGetApplication(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/applications?"+"applicant_email=testuser@creditease.cn", nil)
	return Apply{}.Get(th.Ctx, r)
}

func callGetApplicationByCommit(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/applications?"+"commit_email=testadmin@creditease.cn", nil)
	return Apply{}.Get(th.Ctx, r)
}
/*
func TestApplicationHandle_Post(t *testing.T) {
	th := NewTestHelper2(t)
	createTestUserAndGroup(th)
	callPostApplication(th)
	code, resp := callPostApplicationHandle(th)
}

func callPostApplicationHandle (th *TestHelper) (int, interface{}) {
	url := "http://sso.example.com/api/applications/1"
	r, _ := http.NewRequest("POST", url, nil)
	return Apply{}.Get(th.Ctx, r)
}*/