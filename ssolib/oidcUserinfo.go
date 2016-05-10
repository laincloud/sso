package ssolib

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
)

func (s *Server) UserInfo(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Authorization")
		w.WriteHeader(http.StatusOK)
		return ctx
	}
	log.Debug("userinfo request")
	ret := make(map[string]string)
	u := getCurrentUser(ctx)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", strings.Join([]string{"origin", "Custom-Headers", "Authorization", "authorization", "custom-headers"}, ","))
	if u != nil {
		w.WriteHeader(http.StatusOK)
		ret["sub"] = u.GetSub()
		ret["name"] = u.GetName()
		retjson, _ := json.Marshal(ret)
		w.Write(retjson)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return ctx
}
