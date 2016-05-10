package oauth2

import (
	"testing"

	"github.com/RangelReale/osin"
	"github.com/stretchr/testify/assert"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/app"
)

func init() {
	getApp = func(ctx *models.Context, id int) (*app.App, error) {
		return &app.App{Id: 1}, nil
	}
}

var (
	testOsinAuthorizeData = &osin.AuthorizeData{
		Client:   &app.App{Id: 1},
		Code:     "testcode",
		UserData: AuthorizeUserData{UserId: 1},
	}

	testOsinAccessData = &osin.AccessData{
		Client:        &app.App{Id: 1},
		AccessToken:   "testtoken",
		AuthorizeData: testOsinAuthorizeData,
		UserData:      AccessUserData{UserId: 1},
	}

	testOsinAccessDataWithAccessData = &osin.AccessData{
		Client:        &app.App{Id: 1},
		AccessToken:   "testtoken2",
		AuthorizeData: testOsinAuthorizeData,
		AccessData:    testOsinAccessData,
		UserData:      AccessUserData{UserId: 1},
	}
)

func TestLoadAuthorizeShouldLoadWhatSavedPreviously(t *testing.T) {
	th := NewTestHelper(t)
	s := NewOAuth2Storage(th.Ctx)

	ad := testOsinAuthorizeData
	err := s.SaveAuthorize(ad)
	assert.Nil(t, err)

	ad, err = s.LoadAuthorize("testcode")
	assert.Nil(t, err)
	assert.Equal(t, 1, ad.UserData.(AuthorizeUserData).UserId)
}

func TestSaveAccessShouldHandleAccessDataWithNilUserData(t *testing.T) {
	th := NewTestHelper(t)
	s := NewOAuth2Storage(th.Ctx)
	err := s.SaveAccess(testOsinAccessData)
	assert.Nil(t, err)
}

func TestLoadAccessShouldLoadWhatSavedPreviously(t *testing.T) {
	th := NewTestHelper(t)
	s := NewOAuth2Storage(th.Ctx)

	err := s.SaveAccess(testOsinAccessData)
	assert.Nil(t, err)
	ad, err := s.LoadAccess("testtoken")
	assert.Nil(t, err)
	assert.Equal(t, 1, ad.UserData.(AccessUserData).UserId)
	t.Log(ad)

	// osin 的方式
	t.Log(testOsinAccessDataWithAccessData)
	testOsinAccessDataWithAccessData.UserData = AccessUserData{
		UserId:       ad.UserData.(AccessUserData).UserId,
		AccessDataId: ad.UserData.(AccessUserData).AccessDataId}
	t.Log(testOsinAccessDataWithAccessData)
	err = s.SaveAccess(testOsinAccessDataWithAccessData)
	assert.Nil(t, err)
	ad2, err := s.LoadAccess("testtoken2")
	assert.Nil(t, err)
	t.Log(ad2)
	assert.Equal(t, 1, ad2.UserData.(AccessUserData).UserId)
	assert.Equal(t, "testtoken", ad2.AccessData.AccessToken)
}
