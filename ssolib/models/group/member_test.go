package group

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/laincloud/sso/ssolib/models/testbackend"
)

var (
	testuser1 = &testbackend.TestUser{Id: 1}
	testback1 = &testbackend.TestBackend{}
)

func TestGetGroupsOfUserShouldReturnEmptyListWhenHeIsNotInAnyGroup(t *testing.T) {
	th := NewTestHelper(t)

	gs, err := GetGroupsOfUser(th.Ctx, testuser1)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(gs))
}

func TestGroupGetMemberShouldReturnMemberRole(t *testing.T) {
	th := NewTestHelper(t)

	g, err := CreateGroup(th.Ctx, &Group{
		Name:      "testgroup",
		FullName:  "Test Group",
		GroupType: 0,
	})
	assert.Nil(t, err)

	user := &testbackend.TestUser{
		Name:         "testuser",
		PasswordHash: []byte("test"),
	}
	err = testback1.CreateUser(user, true)
	assert.Nil(t, err)

	u, err := testback1.GetUserByName("testuser")
	assert.Nil(t, err)

	err = g.AddMember(th.Ctx, u, ADMIN)
	assert.Nil(t, err)

	ok, role, err := g.GetMember(th.Ctx, u)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, ADMIN, role)
}

func TestGroupGetMemberShouldReturnMemberRoleInNestedGroup(t *testing.T) {
	EnableNestedGroup()
	th := NewTestHelper(t)

	g, err := CreateGroup(th.Ctx, &Group{
		Name:      "testgroup",
		FullName:  "Test Group",
		GroupType: 0,
	})
	assert.Nil(t, err)

	user := &testbackend.TestUser{
		Name:         "testuser",
		PasswordHash: []byte("test"),
	}
	err = testback1.CreateUser(user, true)
	assert.Nil(t, err)

	u, err := testback1.GetUserByName("testuser")
	assert.Nil(t, err)

	err = g.AddMember(th.Ctx, u, ADMIN)
	assert.Nil(t, err)

	ok, role, err := g.GetMember(th.Ctx, u)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, ADMIN, role)

	ga, err := CreateGroup(th.Ctx, &Group{
		Name: "A",
	})
	assert.Nil(t, err)

	err = ga.AddGroupMember(th.Ctx, g, ADMIN)
	assert.Nil(t, err)

	ok, role, err = ga.GetMember(th.Ctx, u)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, ADMIN, role)

	gb, err := CreateGroup(th.Ctx, &Group{
		Name: "B",
	})
	assert.Nil(t, err)

	gb.AddGroupMember(th.Ctx, ga, NORMAL)
	ok, role, err = gb.GetMember(th.Ctx, u)
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, NORMAL, role)

}
