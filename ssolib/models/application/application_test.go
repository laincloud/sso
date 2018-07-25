package application

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
)

var groupName = "group1"
var applicantEmail = "testapplicant@creditease.cn"
var adminEmail = "testadmin@creditease.cn"
var admin2Email = "testadmin2@creditease.cn"

func TestCreateApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, applicantEmail, "group", &TargetContent{groupName, "normal"},"testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{groupName, "normal"}, a.TargetContent )
}

func TestCreatePendingApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreatePendingApplication(th.Ctx, 1, adminEmail)
	assert.Nil(t, err)
	assert.Equal(t, adminEmail, a.OperatorEmail )
}

func TestUpdateApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, applicantEmail, "group", &TargetContent{groupName, "normal"},"testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{groupName, "normal"}, a.TargetContent )
	b, err := UpdateApplication(th.Ctx, a.Id, "approved", adminEmail)
	assert.Nil(t, err)
	assert.Equal(t, "approved", b.Status)
	assert.Equal(t, adminEmail, b.CommitEmail)
}

func TestFinishApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, applicantEmail, "group", &TargetContent{groupName, "normal"}, "testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{groupName, "normal"}, a.TargetContent)
	b, err := CreatePendingApplication(th.Ctx, a.Id, adminEmail)
	assert.Nil(t, err)
	assert.Equal(t, adminEmail, b.OperatorEmail)
	c, err := CreatePendingApplication(th.Ctx, a.Id, admin2Email)
	assert.Nil(t, err)
	assert.Equal(t, admin2Email, c.OperatorEmail)
	d, err := FinishApplication(th.Ctx, a.Id, "approved", adminEmail)
	assert.Nil(t, err)
	assert.Equal(t, "approved", d.Status)
	assert.Equal(t, adminEmail, d.CommitEmail)
	e, _ := GetPendingApplicationByApplicationId(th.Ctx, a.Id)
	assert.Equal(t, 0, len(e))
}

func TestRecallApplication(t *testing.T) {
	th := NewTestHelper(t)
	a, err := CreateApplication(th.Ctx, applicantEmail, "group", &TargetContent{groupName, "normal"},"testreason")
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{groupName, "normal"}, a.TargetContent )
	b, err := CreatePendingApplication(th.Ctx, a.Id, adminEmail)
	assert.Nil(t, err)
	assert.Equal(t, adminEmail, b.OperatorEmail )
	c, err := CreatePendingApplication(th.Ctx, a.Id, admin2Email)
	assert.Nil(t, err)
	assert.Equal(t, admin2Email, c.OperatorEmail )
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
		CreateApplication(th.Ctx, applicantEmail, "group", &TargetContent{groupName, "normal"},"testreason")
	}
	as, err := GetApplications(th.Ctx, applicantEmail, "initialled", 2, 7 )
	assert.Nil(t, err)
	assert.Equal(t, 5, len(as))
	assert.Equal(t,8, as[0].Id)
	as, err = GetApplications(th.Ctx, applicantEmail, "approved", 2, 7 )
	assert.Nil(t, err)
	assert.Equal(t, 0, len(as))
}

func TestGetAllApplications(t *testing.T) {
	th := NewTestHelper(t)
	for i := 0; i < 10; i++ {
		groupName := "group" + string(i)
		time.Sleep(1 * time.Second)
		CreateApplication(th.Ctx, applicantEmail, "group", &TargetContent{groupName, "normal"},"testreason")
	}
	as, err := GetAllApplications(th.Ctx, "", 2, 8)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(as))
}
