package ssolib

import (
	"testing"
	"net/http"
	"strings"
	"github.com/laincloud/sso/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/laincloud/sso/ssolib/models/application"
)

func TestApply_Post(t *testing.T) {
	th := NewTestHelper(t)
	code ,resp := createTestUserAndApplication(th)
	assert.Equal(t, code, http.StatusOK)
	a, ok := resp.(*Application)
	assert.True(t, ok)
	assert.Equal(t, "testing", a.Reason)
	assert.Equal(t, application.TargetContent{Name:"group1",Role:"normal",}, a.Target)

}

func createTestUserAndApplication(th *TestHelper) (int, interface{}) {
	createTestUser(th, "testuser")
	th.loginWithScope("testuser", "")
	return callPostApplication(th)
}

func callPostApplication(th *TestHelper) (int, interface{}) {
	r, _ := http.NewRequest("POST", "http://sso.example.com/api/applications",
		strings.NewReader(`{"target_type": "group", "reason":"testing", "target": [{"name": "group1","role":"normal"}]}`))
	return Apply{}.Post(th.Ctx, r)
}