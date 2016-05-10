package ssolib

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/RangelReale/osin"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

const (
	OPENID_SCOPE = "openid"
	ID_TOKEN     = "id_token"
)

const (
	TOKEN_IDTOKEN osin.AuthorizeRequestType = "token id_token"
)

// TODO 完整的 openid-configuration
type OIDC_Configuration struct {
	Issuer                string                      `json:"issuer"`
	AuthEnd               string                      `json:"authorization_endpoint"`
	TokenEnd              string                      `json:"token_endpoint"`
	UserInfoEnd           string                      `json:"userinfo_endpoint"`
	JwksUri               string                      `json:"jwks_uri"`
	ResponseTypeSupported []osin.AuthorizeRequestType `json:"response_types_supported"`
}

func (s *Server) OidcConfig(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	log.Debug("openid connect config request")
	ret := s.NewOidcConfig(ctx)
	retjson, _ := json.Marshal(ret)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(retjson)
	return ctx
}

func (s *Server) NewOidcConfig(ctx context.Context) *OIDC_Configuration {
	mctx := getModelContext(ctx)
	ssoname := mctx.SSOSiteURL.String()
	ret := &OIDC_Configuration{}
	ret.Issuer = ssoname

	// TODO 去掉这些硬编码以及 server.go 里的
	ret.AuthEnd = ssoname + "/oauth2/auth"
	ret.TokenEnd = ssoname + "/oauth2/token"
	ret.JwksUri = ssoname + "/oauth2/certs"
	ret.UserInfoEnd = ssoname + "/oauth2/userinfo"
	ret.ResponseTypeSupported = getOAuth2Provider(ctx).Config.AllowedAuthorizeTypes
	return ret
}

func IsAuthorizeRequestTypeEqual(t1, t2 osin.AuthorizeRequestType) bool {
	t1s := strings.Split(string(t1), " ")
	sort.Strings(t1s)
	t2s := strings.Split(string(t2), " ")
	sort.Strings(t2s)
	return reflect.DeepEqual(t1s, t2s)
}
