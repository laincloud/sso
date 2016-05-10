package ssolib

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"strings"

	"github.com/docker/libtrust"
)

type TokenConfig struct {
	Issuer     string
	CertFile   string
	KeyFile    string
	Expiration int64
	publicKey  libtrust.PublicKey
	privateKey libtrust.PrivateKey
}

// 关于token相关的一些配置
var tokenConfig = TokenConfig{
	Issuer:     "auth server",
	Expiration: 900,
}

func loadCertAndKey(certFile, keyFile string) (pk libtrust.PublicKey, prk libtrust.PrivateKey, err error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return
	}
	pk, err = libtrust.FromCryptoPublicKey(x509Cert.PublicKey)
	if err != nil {
		return
	}
	prk, err = libtrust.FromCryptoPrivateKey(cert.PrivateKey)
	return
}

func joseBase64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
