package application

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var groupName = "group1"
var applicantEmail = "testapplicant@example.com"
var adminEmail = "testadmin@example.com"
var admin2Email = "testadmin2@example.com"

func TestCreateApplication(t *testing.T) {
	th := NewTestHelper(t)
	application := &Application{ApplicantEmail:applicantEmail, TargetType: "group", TargetContent: &TargetContent{groupName, "normal", 0}, Reason: "testreason"}
	adminEmails := []string{adminEmail}
	a, err := CreateApplication(th.Ctx, application, adminEmails)
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{groupName, "normal", 0}, a.TargetContent )
}



func TestFinishApplication(t *testing.T) {
	th := NewTestHelper(t)
	application := &Application{ApplicantEmail:applicantEmail, TargetType: "group", TargetContent: &TargetContent{groupName, "normal", 0}, Reason: "testreason"}
	adminEmails := []string{adminEmail,admin2Email}
	a, err := CreateApplication(th.Ctx, application, adminEmails)
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{groupName, "normal", 0}, a.TargetContent)
	d, err := FinishApplication(th.Ctx, a.Id, "approved", adminEmail)
	assert.Nil(t, err)
	assert.Equal(t, "approved", d.Status)
	assert.Equal(t, adminEmail, d.CommitEmail)
	e, _ := GetPendingApplicationByApplicationId(th.Ctx, a.Id)
	assert.Equal(t, 0, len(e))
}

func TestRecallApplication(t *testing.T) {
	th := NewTestHelper(t)
	application := &Application{ApplicantEmail:applicantEmail, TargetType: "group", TargetContent: &TargetContent{groupName, "normal", 0}, Reason: "testreason"}
	adminEmails := []string{adminEmail}
	a, err := CreateApplication(th.Ctx, application, adminEmails)
	assert.Nil(t, err)
	assert.Equal(t, &TargetContent{groupName, "normal", 0}, a.TargetContent )
	err = RecallApplication(th.Ctx, a.Id)
	assert.Nil(t, err)
	d, err := GetApplication(th.Ctx, a.Id)
	assert.Nil(t, d)
}

func TestGetApplications(t *testing.T) {
	th := NewTestHelper(t)
	for i := 0; i < 10; i++ {
		groupName := "group" + string(i)
		application := &Application{ApplicantEmail:applicantEmail, TargetType: "group", TargetContent: &TargetContent{groupName, "normal", 0}, Reason: "testreason"}
		adminEmails := []string{adminEmail}
		CreateApplication(th.Ctx, application, adminEmails)
	}
	as, _, err := GetApplications(th.Ctx, applicantEmail, "initialled", 2, 7 )
	assert.Nil(t, err)
	assert.Equal(t, 6, len(as))
	assert.Equal(t,8, as[0].Id)
	as, _, err = GetApplications(th.Ctx, applicantEmail, "approved", 2, 7 )
	assert.Nil(t, err)
	assert.Equal(t, 0, len(as))
}

func TestGetAllApplications(t *testing.T) {
	th := NewTestHelper(t)
	for i := 0; i < 10; i++ {
		groupName := "group" + string(i)
		application := &Application{ApplicantEmail:applicantEmail, TargetType: "group", TargetContent: &TargetContent{groupName, "normal", 0}, Reason: "testreason"}
		adminEmails := []string{adminEmail}
		CreateApplication(th.Ctx, application, adminEmails)
	}
	as, _, err := GetAllApplications(th.Ctx, "", 2, 8)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(as))
}
