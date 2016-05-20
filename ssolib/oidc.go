package ssolib

import (
	"net/http"
	"strings"
	"time"

	"github.com/RangelReale/osin"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

func CreateIDToken(issuer string, client_id string, sub string, userInfo *UserWithGroups, nonce string, access_token string) (string, error) {

	// get signing method
	_, sigAlg, err := tokenConfig.privateKey.Sign(strings.NewReader("dummy"), 0)
	if err != nil {
		log.Error(err)
		return "", err
	}

	var jwt JWT
	jwt.Init()

	jwt.Header["alg"] = sigAlg
	jwt.Header["typ"] = "JWT"
	jwt.Header["kid"] = tokenConfig.publicKey.KeyID()

	now := time.Now().Unix()

	jwt.Claims["iss"] = issuer
	jwt.Claims["sub"] = sub
	jwt.Claims["aud"] = client_id
	jwt.Claims["iat"] = now
	jwt.Claims["exp"] = now + tokenConfig.Expiration
	jwt.Claims["user_info"] = userInfo.User
	if nonce != "" {
		jwt.Claims["nonce"] = nonce
	}
	if access_token != "" {
		ac_hash, _, _ := tokenConfig.privateKey.Sign(strings.NewReader(access_token), 0)
		jwt.Claims["at_hash"] = string(joseBase64UrlEncode(ac_hash[0:16]))
	}

	err = jwt.Sign()

	return jwt.Token, err
}

func hasOpenidScope(output osin.ResponseData) bool {
	if elem, ok := output["scope"]; ok {
		scopes := strings.Split(elem.(string), " ")
		for _, s := range scopes {
			log.Debug(s)
			if s == OPENID_SCOPE {
				return true
			}
		}
	} else {
		log.Error("can not find field scope")
	}
	return false
}

func setIDTokenInResponseOutput(ctx context.Context, resp *osin.Response, client_id string, userid int, nonce string, access_token string) {
	mctx := getModelContext(ctx)
	ub := getUserBackend(ctx)
	userOb, err := ub.GetUser(userid)
	if err != nil {
		log.Error(err)
		resp.StatusCode = http.StatusInternalServerError
		return
	}
	userInfo := GetUserWithGroups(ctx, userOb)
	issuer := mctx.SSOSiteURL.String()
	sub := userOb.GetSub()
	idtoken, err := CreateIDToken(issuer, client_id, sub, userInfo, nonce, access_token)
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		resp.Output["id_token"] = ""
	} else {
		resp.Output["id_token"] = idtoken
	}
}
