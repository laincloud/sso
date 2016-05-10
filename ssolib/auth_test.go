package ssolib

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/laincloud/sso/ssolib/models/app"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestAuthorizeEndpointShouldReturnLoginFormWhenRequestIsGet(t *testing.T) {
	th := NewTestHelper(t)

	r, _ := http.NewRequest("GET",
		"http://sso.example.com/oauth2/auth?"+
			"client_id=1&response_type=code&redirect_uri=http%3A//example.com",
		nil)
	w := httptest.NewRecorder()
	s := &Server{}

	mock := mockReverse(th, "http://sso.example.com/oauth2/auth")
	defer mock.restore()

	mctx := getModelContext(th.Ctx)
	oauth2Provider, err := s.initOAuth2Provider(mctx)
	assert.Nil(t, err)
	th.Ctx = context.WithValue(th.Ctx, "oauth2", oauth2Provider)

	app.InitDatabase(mctx)
	u := createTestUser(th, "testuser")
	t.Log(u)
	appSpec := &app.App{
		Secret:      "secret",
		FullName:    "Test App",
		RedirectUri: "http://example.com",
	}
	_, err = app.CreateApp(mctx, appSpec, u)
	assert.Nil(t, err)

	s.AuthorizationEndpoint(th.Ctx, w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	// assert form in returned HTML
	assert.Contains(t, w.Body.String(), `action="http://sso.example.com/oauth2/auth`)
}
