package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

var urlPrefix = flag.String("url", "http://localhost:8888/_lain/auth", "url prefix")
var authEndpoint = flag.String("auth", "http://localhost:14000/oauth2/auth", "authorize endpoint")
var tokenEndpoint = flag.String("token", "http://localhost:14000/oauth2/token", "token endpoint")
var clientId = flag.String("clientid", "1", "client id")
var clientSecret = flag.String("clientSecret", "admin", "client secret")

func main() {
	flag.Parse()

	u, err := url.Parse(*urlPrefix)
	if err != nil {
		log.Fatal("urlPrefix is not a valid url!")
	}
	pathPrefix := u.Path
	redirectUri := *urlPrefix + "/code"

	conf := &oauth2.Config{
		ClientID:     *clientId,
		ClientSecret: *clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  *authEndpoint,
			TokenURL: *tokenEndpoint,
		},
		RedirectURL: redirectUri,
		Scopes:      []string{"openid", "email", "groups"},
	}

	http.HandleFunc(pathPrefix+"/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		state := "next=" + r.Form.Get("next")
		// TODO: put anti CSRF session token in state

		url := conf.AuthCodeURL(state)
		http.Redirect(w, r, url, 302)
	})

	http.HandleFunc(pathPrefix+"/code", func(w http.ResponseWriter, r *http.Request) {
		// TODO: validate state for CSRF attack

		var err error = nil

		defer func() {
			if err != nil {
				// TODO: correct status code
				w.Write([]byte("<html><body>"))
				w.Write([]byte(err.Error()))
				w.Write([]byte("</body></html>"))
			}
		}()

		// Exchange code for access token and ID token
		token, err := conf.Exchange(oauth2.NoContext, r.FormValue("code"))
		if err != nil {
			return
		}

		w.Write([]byte(fmt.Sprintf("Access Token: %s", token.AccessToken)))
	})

	http.ListenAndServe(":8888", nil)
}
