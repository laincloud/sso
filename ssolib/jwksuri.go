package ssolib

import (
	"encoding/json"
	"net/http"

	"github.com/mendsley/gojwk"

	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

type JWKs struct {
	Keys []*gojwk.Key `json:"keys"`
}

func (s *Server) Jwks_uriEndpoint(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	log.Debug("jwks request")
	jwk, _ := gojwk.PublicKey(tokenConfig.publicKey.CryptoPublicKey())
	jwk.Kid = tokenConfig.publicKey.KeyID()
	jwk.Alg = "RS256"
	ret := &JWKs{
		make([]*gojwk.Key, 0, 1),
	}
	ret.Keys = append(ret.Keys, jwk)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	retjson, _ := json.Marshal(ret)
	w.Write(retjson)
	return ctx
}
