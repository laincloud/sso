package ssolib

import (
	"testing"
	"net/http"
	"github.com/laincloud/sso/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestRolesResource_Get(t *testing.T) {
	th := NewTestHelper(t)
	createTestUserWithEmail(th, "testuser")
	th.loginWithScope("testuser", "write:app read:role")
	createApp(th)
	createRole(th)
	code ,resp := getRole(th)
	assert.Equal(t, http.StatusOK, code)
	a, ok := resp.([]RoleMembers)
	assert.True(t, ok)
	assert.Equal(t, 1, len(a))
}


func getRole(th *TestHelper) (int, interface{}){
	r, _ := http.NewRequest("GET", "http://sso.example.com/api/roles?app_id=1&all=true", nil)
	return RolesResource{}.Get(th.Ctx, r)
}

