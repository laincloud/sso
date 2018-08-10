package ssolib

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/RangelReale/osin"
	"github.com/go-sql-driver/mysql"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/oauth2"
)

var loginTemplate *template.Template

func init() {
	os.Setenv("TEMPLATES_PATH","/Users/yixinf-o/go/src/github.com/laincloud/sso/templates")
	templatesPath := os.Getenv("TEMPLATES_PATH")
	if templatesPath == "" {
		templatesPath = "./templates"
	}
	loginTemplate = template.Must(template.ParseFiles(templatesPath + "/login.html"))
}

// var loginTemplate = template.Must(template.New("login").Parse(`
// <html><body>
// <p>{{.ClientName}} is requesting your permission to access your information</p>
// {{if .Scopes}}
// Requested permissions:
// <ul>
//   {{range .Scopes}}
//     <li>{{.}}</li>
//   {{end}}
// </ul>
// {{end}}
// <form action="{{.FormURL}}" method="POST">
// Login: <input type="text" name="login" /><br/>
// Password: <input type="password" name="password" /><br/>
// <input type="submit" />
// {{if .Err }}<div class="error">{{.Err}}</div>{{end}}
// </form>
// </body></html>
// `))

// Why not use strings.Split directly? Because it can only handle single space delimeter.
var (
	reverse = func(s *Server, route string, params ...interface{}) string {
		return s.Reverse(route, params...)
	}
)

type loginTemplateContext struct {
	ClientName string
	FormURL    string
	Err        error
	Scopes     []string
}

func (s *Server) AuthorizationEndpoint(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	log.Debugf("sso_debug: sso_oauth_auth api begin.")

	oauth2p := getOAuth2Provider(ctx)
	resp := oauth2p.NewResponse()
	defer func() {
		if resp.IsError && resp.InternalError != nil {
			log.Error(resp.InternalError)
		}
		resp.Close()
	}()

	// response_type 必须带 code, 或者 token 之一，当 response_type 带有 token 时，可以有 id_token
	// 换句话说，当前的 response_type 只有三种情况
	// "token" "code" "token id_token"
	r.ParseForm()
	oidc := false
	res_type := osin.AuthorizeRequestType(r.Form.Get("response_type"))
	if IsAuthorizeRequestTypeEqual(res_type, TOKEN_IDTOKEN) {
		oidc = true
		// 因为 TOKEN_IDTOKEN 的处理逻辑和 token 的处理逻辑有相同之处，这里使用 token 传入 osin 减少重复代码
		r.Form.Set("response_type", string(osin.TOKEN))
		res_type = osin.TOKEN
	}

	ar := oauth2p.HandleAuthorizeRequest(resp, r)
	if ar == nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return ctx
	}

	mctx := getModelContext(ctx)

	var tmplContextErr error

	if r.Method == "POST" {
		login, password := r.FormValue("login"), r.FormValue("password")
		log.Debugf("sso_debug: sso_oauth_auth api load info from db begin.")
		ub := getUserBackend(ctx)
		if s.queryUser { // for detail errors of login, i.e. "no such user"
			u, err := ub.GetUserByFeature(login)
			log.Debug(u)
			if err != nil {
				log.Debugf("sso_debug: sso_oauth_auth api load user info from db fail.")
				if err == iuser.ErrUserNotFound {
					tmplContextErr = errors.New("No such user")
				} else {
					if mysqlError, ok := err.(*mysql.MySQLError); ok {
						if mysqlError.Number == 1267 {
							// for "Illegal mix of collations (latin1_swedish_ci,IMPLICIT) and (utf8_general_ci,COERCIBLE) for operation '='"
							log.Info(err.Error())
							tmplContextErr = errors.New("No such user")
						} else {
							panic(err)
						}
					} else {
						panic(err)
					}
				}
			} else if ok, _ := ub.AuthPassword(u.GetSub(), password); ok {
				if res_type == osin.CODE {
					ar.UserData = oauth2.AuthorizeUserData{UserId: u.GetId()}
				} else if res_type == osin.TOKEN {
					ar.UserData = oauth2.AccessUserData{UserId: u.GetId()}
				} else {
					panic("unknown response_type and osin didn't handle it")
				}
				ar.Authorized = true
				oauth2p.FinishAuthorizeRequest(resp, r, ar)
				if oidc {
					setIDTokenInResponseOutput(ctx, resp, r.Form.Get("client_id"),
						u.GetId(), r.Form.Get("nonce"), resp.Output["access_token"].(string))
				}
				osin.OutputJSON(resp, w, r)
				log.Debugf("sso_debug: sso_oauth_auth api load info from db end.")
				return ctx
			} else {
				tmplContextErr = errors.New("incorrect password")
			}
			log.Debugf("sso_debug: sso_oauth_auth api load info from db end.")
		} else { // only gives "no such user or incorrect password "
			if ok, u, err := ub.AuthPasswordByFeature(login, password); ok {
				if res_type == osin.CODE {
					ar.UserData = oauth2.AuthorizeUserData{UserId: u.GetId()}
				} else if res_type == osin.TOKEN {
					ar.UserData = oauth2.AccessUserData{UserId: u.GetId()}
				} else {
					panic("unknown response_type and osin didn't handle it")
				}
				ar.Authorized = true
				oauth2p.FinishAuthorizeRequest(resp, r, ar)
				if oidc {
					setIDTokenInResponseOutput(ctx, resp, r.Form.Get("client_id"),
						u.GetId(), r.Form.Get("nonce"), resp.Output["access_token"].(string))
				}
				osin.OutputJSON(resp, w, r)
				log.Debugf("sso_debug: sso_oauth_auth api load info from db end.")
				return ctx
			} else {
				log.Debug(err)
				tmplContextErr = errors.New("user does not exist or password is incorrect")
			}
		}
	}

	appIdString := ar.Client.GetId()
	appId, err := strconv.Atoi(appIdString)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return ctx
	}
	log.Debugf("sso_debug: sso_oauth_auth api begin load app info from db.")
	ap, err := app.GetApp(mctx, appId)
	if err != nil {
		if err == app.ErrAppNotFound {
			http.Error(w, "Client not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error when verify client", http.StatusInternalServerError)
		}
		return ctx
	}
	log.Debugf("sso_debug: sso_oauth_auth api end load app info from db.")

	tmplContext := loginTemplateContext{
		ClientName: ap.FullName,
		FormURL:    reverse(s, "PostLoginForm") + "?" + r.URL.RawQuery,
		Err:        tmplContextErr,
		Scopes:     split(ar.Scope),
	}
	s.renderHtmlTemplate(w, loginTemplate, tmplContext)
	log.Debugf("sso_debug: sso_oauth_auth api end.")

	return ctx
}

type AuthenticateWare struct {
}

func NewAuthenticateMiddleware() server.Middleware {
	return &AuthenticateWare{}
}

func (aw *AuthenticateWare) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, next server.Handler) context.Context {
	log.Debugf("sso_debug: sso_server_http api begin")
	if _, ok := r.Header["Authorization"]; !ok && r.FormValue("access_token") == "" {
		// No auth info
		return next(ctx, w, r)
	}

	user, scope := func() (iuser.User, string) {
		oauth2p := getOAuth2Provider(ctx)
		resp := oauth2p.NewResponse()
		defer resp.Close()

		log.Debugf("sso_debug: sso_server_http parse request begin")
		ir := oauth2p.HandleInfoRequest(resp, r)
		if ir == nil {
			return nil, ""
		}

		userData, ok := ir.AccessData.UserData.(oauth2.AccessUserData)
		if !ok {
			panic("Load userinfo failed")
		}
		log.Debugf("sso_debug: sso_server_http parse request end")

		log.Debugf("sso_debug: sso_server_http get user from db begin")
		ub := getUserBackend(ctx)
		user, err := ub.GetUser(userData.UserId)
		if err != nil {
			panic(err)
		}
		log.Debugf("sso_debug: sso_server_http get user from db end")
		return user, ir.AccessData.Scope
	}()

	if user != nil {
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "scope", split(scope))
	}
	log.Debugf("sso_debug: sso_server_http api end")
	return next(ctx, w, r)
}

func requireLogin(ctx context.Context) error {
	user := getCurrentUser(ctx)
	if user == nil {
		return errors.New("require login in")
	}
	return nil
}

func requireScope(ctx context.Context, scope string) error {
	err := requireLogin(ctx)
	if err != nil {
		return err
	}
	scopes := getScope(ctx)
	if scopes != nil {
		for _, s := range scopes {
			if s == scope {
				return nil
			}
		}
	}
	return errors.New("require scope")
}
