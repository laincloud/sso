package ssolib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mijia/sweb/log"
)

// registryV2token 里的 JWT 实现用到了 github.com/docker/distribution/registry/auth/token，
// 而这个 JWT 的 Claims 写死了很多东西，缺少一些可以自定义的 field
// 以下的实现中，JWT 的 签名利用的是 docker 库 github.com/docker/libtrust 中的签名方法

type JWT struct {
	Header map[string]interface{}
	Claims map[string]interface{}
	Sig    string
	Token  string
}

func (jwt *JWT) Init() {
	jwt.Header = make(map[string]interface{})
	jwt.Claims = make(map[string]interface{})
}

func (jwt *JWT) HeaderJson() []byte {
	return objectJson(jwt.Header)
}

func (jwt *JWT) ClaimsJson() []byte {
	return objectJson(jwt.Claims)
}

func objectJson(o interface{}) []byte {
	j, err := json.Marshal(o)
	if err != nil {
		log.Error(err)
	}
	return j
}

func (jwt *JWT) Sign() error {
	bHeaderJSON := jwt.HeaderJson()
	bClaimsJSON := jwt.ClaimsJson()
	payload := fmt.Sprintf("%s.%s", joseBase64UrlEncode(bHeaderJSON), joseBase64UrlEncode(bClaimsJSON))

	sig, sigAlg2, err := tokenConfig.privateKey.Sign(strings.NewReader(payload), 0)

	sigAlg := jwt.Header["alg"]

	if err != nil {
		log.Error("signing failed")
		return err
	}

	if sigAlg2 != sigAlg {
		log.Error("jwt header is wrong")
		return errors.New("jwt header is wrong since private key is probably changed")
	}

	jwt.Sig = joseBase64UrlEncode(sig)

	idtoken := fmt.Sprintf("%s.%s", payload, jwt.Sig)
	jwt.Token = idtoken
	return nil
}
