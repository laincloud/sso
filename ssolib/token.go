package ssolib

import (
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/oauth2"
)

type TokenEndpoint struct {
	server.BaseResource
}

func (te TokenEndpoint) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	oauth2p := getOAuth2Provider(ctx)
	resp := oauth2p.NewResponse()
	defer resp.Close()

	if ar := oauth2p.HandleAccessRequest(resp, r); ar != nil {

		ar.Authorized = true

		r.ParseForm()
		grantType := osin.AccessRequestType(r.Form.Get("grant_type"))

		var userId, acId int
		switch grantType {
		case osin.AUTHORIZATION_CODE:
			userId = ar.AuthorizeData.UserData.(oauth2.AuthorizeUserData).UserId
			ar.UserData = oauth2.AccessUserData{UserId: userId}
		case osin.REFRESH_TOKEN:
			userId = ar.AccessData.UserData.(oauth2.AccessUserData).UserId
			acId = ar.UserData.(oauth2.AccessUserData).AccessDataId
			ar.UserData = oauth2.AccessUserData{AccessDataId: acId, UserId: userId}
		}
		oauth2p.FinishAccessRequest(resp, r, ar)

		if hasOpenidScope(resp.Output) {
			setIDTokenInResponseOutput(ctx, resp, r.Form.Get("client_id"), userId, "", "")
		}
	}
	if resp.IsError && resp.InternalError != nil {
		log.Error(resp.InternalError)
	}
	return resp.StatusCode, resp.Output
}

func (te TokenEndpoint) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return te.Post(ctx, r)
}
