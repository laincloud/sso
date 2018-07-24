package application

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestCreateApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, "testapplicant@creditease.cn", "group", &TargetContent{"group1", "normal"},"testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{"group1", "normal"}, a.TargetContent )
}

func TestCreatePendingApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, "testapplicant@creditease.cn", "group", &TargetContent{"group1", "normal"},"testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{"group1", "normal"}, a.TargetContent )
	b, err := CreatePendingApplication(th.Ctx, a.Id, "testadmin@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "testadmin@creditease.cn", b.OperatorEmail )
}

func TestUpdateApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, "testapplicant@creditease.cn", "group", &TargetContent{"group1", "normal"},"testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{"group1", "normal"}, a.TargetContent )
	b, err := CreatePendingApplication(th.Ctx, a.Id, "testadmin@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "testadmin@creditease.cn", b.OperatorEmail )
	c, err := CreatePendingApplication(th.Ctx, a.Id, "testadmin2@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "testadmin2@creditease.cn", c.OperatorEmail )
	d, err := UpdateApplication(th.Ctx, a.Id, "approved", "testadmin@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "approved", d.Status)
	assert.Equal(t, "testadmin@creditease.cn", d.CommitEmail)
}

func TestFinishApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, "testapplicant@creditease.cn", "group", &TargetContent{"group1", "normal"}, "testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{"group1", "normal"}, a.TargetContent)
	b, err := CreatePendingApplication(th.Ctx, a.Id, "testadmin@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "testadmin@creditease.cn", b.OperatorEmail)
	c, err := CreatePendingApplication(th.Ctx, a.Id, "testadmin2@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "testadmin2@creditease.cn", c.OperatorEmail)
	d, err := FinishApplication(th.Ctx, a.Id, "approved", "testadmin@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "approved", d.Status)
	assert.Equal(t, "testadmin@creditease.cn", d.CommitEmail)
	e, _ := GetPendingApplicationByApplicationId(th.Ctx, a.Id)
	assert.Equal(t, 0, len(e))
}

func TestRecallApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, "testapplicant@creditease.cn", "group", &TargetContent{"group1", "normal"},"testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{"group1", "normal"}, a.TargetContent )
	b, err := CreatePendingApplication(th.Ctx, a.Id, "testadmin@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "testadmin@creditease.cn", b.OperatorEmail )
	c, err := CreatePendingApplication(th.Ctx, a.Id, "testadmin2@creditease.cn")
	assert.Nil(t, err)
	assert.Equal(t, "testadmin2@creditease.cn", c.OperatorEmail )
	err = RecallApplication(th.Ctx, a.Id)
	assert.Nil(t, err)
	d, err := GetApplication(th.Ctx, a.Id)
	assert.Nil(t, d)
}

func TestGetApplications(t *testing.T) {
	th := NewTestHelper(t)
	for i := 0; i < 10; i++ {
		groupName := "group" + string(i)
		time.Sleep(1 * time.Second)
		CreateApplication(th.Ctx, "testapplicant@creditease.cn", "group", &TargetContent{groupName, "normal"},"testreason")
	}
	as, err := GetApplications(th.Ctx, "testapplicant@creditease.cn", "initialled", 2, 7 )
	assert.Nil(t, err)
	assert.Equal(t, 5, len(as))
	assert.Equal(t,8, as[0].Id)
	as, err = GetApplications(th.Ctx, "testapplicant@creditease.cn", "approved", 2, 7 )
	assert.Nil(t, err)
	assert.Equal(t, 0, len(as))
}

func TestGetAllApplications(t *testing.T) {
	th := NewTestHelper(t)
	for i := 0; i < 10; i++ {
		groupName := "group" + string(i)
		time.Sleep(1 * time.Second)
		CreateApplication(th.Ctx, "testapplicant@creditease.cn", "group", &TargetContent{groupName, "normal"},"testreason")
	}
	as, err := GetAllApplications(th.Ctx, "", 2, 8)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(as))
}
