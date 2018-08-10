// OpenID Connect Server

package ssolib

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"

	"github.com/RangelReale/osin"
	"github.com/getsentry/raven-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/render"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/lock"
	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/app"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/oauth2"
	"github.com/laincloud/sso/ssolib/models/role"
	"github.com/laincloud/sso/ssolib/utils"
)

// 添加一些与后端相关的 handler, 如用户注册相关等
type AddHandles func(*Server)

type Server struct {
	*server.Server
	render *render.Render

	mysqlDSN          string
	siteURL           string
	smtpAddr          string
	emailFrom         string
	emailFromPassword string
	emailTLS          bool
	emailSuffix       string
	isDebug           bool

	prikeyfile string
	pubkeyfile string
	sentryDSN  string

	queryUser bool

	userBackend iuser.UserBackend
}

func NewServer(
	mysqlDSN, siteURL, smtpAddr, emailFrom, emailFromPassword, emailSuffix string,
	emailTLS, isDebug bool, prikeyfile string, pubkeyfile string, sentryDSN string, queryUser bool) *Server {
	srv := &Server{
		mysqlDSN:          mysqlDSN,
		siteURL:           siteURL,
		smtpAddr:          smtpAddr,
		emailFrom:         emailFrom,
		emailFromPassword: emailFromPassword,
		emailTLS:          emailTLS,
		emailSuffix:       emailSuffix,
		isDebug:           isDebug,
		prikeyfile:        prikeyfile,
		pubkeyfile:        pubkeyfile,
		sentryDSN:         sentryDSN,
		queryUser:         queryUser,
	}
	return srv
}

func (s *Server) SetUserBackend(ub iuser.UserBackend) {
	s.userBackend = ub
}

func (s *Server) ListenAndServe(addr string, addHandlers AddHandles) error {

	group.EnableNestedGroup()
	//	group.SetMaxDepth(4)

	db, err := utils.InitMysql(s.mysqlDSN)
	if err != nil {
		return err
	}

	// since there are many databases to modify
	// it will be deleted in the next version
	log.Debug(s.siteURL)

	siteURL, err := url.Parse(s.siteURL)
	if err != nil {
		return err
	}

	dLock := lock.New(s.mysqlDSN, "groupdag")

	mctx := &models.Context{
		DB:                db,
		SSOSiteURL:        siteURL,
		SMTPAddr:          s.smtpAddr,
		EmailFrom:         s.emailFrom,
		EmailFromPassword: s.emailFromPassword,
		EmailTLS:          s.emailTLS,
		Back:              s.userBackend,
		Lock:              dLock,
	}

	s.initDatabase(mctx)
	s.userBackend.InitModel(mctx)

	oauth2Provider, err := s.initOAuth2Provider(mctx)
	if err != nil {
		return err
	}

	sentryClient, err := raven.NewClient(s.sentryDSN, nil)
	if err != nil {
		log.Error(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "oauth2", oauth2Provider)
	ctx = context.WithValue(ctx, "db", db)
	ctx = context.WithValue(ctx, "mctx", mctx)
	ctx = context.WithValue(ctx, "emailSuffix", s.emailSuffix)
	ctx = context.WithValue(ctx, "userBackend", s.userBackend)
	s.Server = server.New(ctx, s.isDebug)

	s.render = s.initRender()
	s.EnableExtraAssetsJson("assets_map.json")

	s.Middleware(server.NewSentryRecoveryWare(sentryClient, s.isDebug))
	s.Middleware(NewAuthenticateMiddleware())

	s.RestfulHandlerAdapter(s.adaptResourceHandler)
	s.Get("/", "Home", s.Home)
	s.Get("/spa/*lochash", "SsoSpa", s.PageApplication)

	s.Get("/oauth2/auth", "AuthorizationEndpoint", s.AuthorizationEndpoint)
	s.Post("/oauth2/auth", "PostLoginForm", s.AuthorizationEndpoint)
	s.AddRestfulResource("/oauth2/token", "TokenEndpoint", TokenEndpoint{})

	s.Get("/.well-known/openid-configuration", "OidcConfig", s.OidcConfig)
	s.Get("/oauth2/certs", "Jwks_uriEndpoint", s.Jwks_uriEndpoint)
	s.Get("/oauth2/userinfo", "UserInfo", s.UserInfo)
	s.Handle("OPTIONS", "/oauth2/userinfo", "UserInfo", s.UserInfo)

	s.Get("/api/users", "UsersList", s.UsersList)

	s.Post("/api/resources/delete", "ResourcesDelete", s.ResourcesDelete)
	s.Post("/api/rolemembers", "RoleMembers", s.RoleMembers)
	s.Get("/api/batch-users", "BatchUsers", s.BatchUsers)

	addHandlers(s)

	s.AddRestfulResource("/api/users/:username", "UserResource", UserResource{})
	s.AddRestfulResource("/api/me", "MeResource", MeResource{})
	s.AddRestfulResource("/api/apps", "AppsResource", AppsResource{})
	s.AddRestfulResource("/api/app/:id", "AppResource", AppResource{})
	s.AddRestfulResource("/api/groups", "GroupsResource", GroupsResource{})
	s.AddRestfulResource("/api/groups/:groupname", "GroupResource", GroupResource{})
	s.AddRestfulResource("/api/groups/:groupname/members/:username",
		"MemberResource", MemberResource{})
	s.AddRestfulResource("/api/groups/:groupname/group-members/:sonname", "GroupMemberResource", GroupMemberResource{})
	s.AddRestfulResource("/api/app_roles", "AppRoleResource", AppRoleResource{})
	s.AddRestfulResource("/api/resources/:id", "ResourceResource", ResourceResource{})
	s.AddRestfulResource("/api/resources", "ResourcesResource", ResourcesResource{})
	s.AddRestfulResource("/api/roles", "RolesResource", RolesResource{})
	s.AddRestfulResource("/api/roles/:id", "RoleResource", RoleResource{})
	s.AddRestfulResource("/api/roles/:id/members/:username", "RoleMemberResource", RoleMemberResource{})
	s.AddRestfulResource("/api/roles/:id/resources", "RoleResourceResource", RoleResourceResource{})
	s.AddRestfulResource("/api/applications", "Apply", Apply{})
	s.AddRestfulResource("/api/applications/:application_id", "ApplicationHandle", ApplicationHandle{})
	s.AddRestfulResource("/api/app_info", "AppInformation", AppInformation{})

	puk, prk, err := loadCertAndKey(s.pubkeyfile, s.prikeyfile)
	if err != nil {
		panic(err)
	}

	tokenConfig.privateKey = prk
	tokenConfig.publicKey = puk

	s.Files("/static/*filepath", http.Dir("public"))
	s.Files("/apidoc/*filepath", http.Dir("apidoc"))
	s.Files("/assets/*filepath", http.Dir("public"))

	s.NotFound(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		s.renderError(w, http.StatusNotFound, "API not found", "")
		return ctx
	})
	s.MethodNotAllowed(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		s.renderError(w, http.StatusMethodNotAllowed, "Method is not allowed", "")
		return ctx
	})

	// 如果 ssoserver 和 mysql 中间的防火墙会丢弃长时间 idle 的连接上的数据包，导致查询超时的话，可以去掉下一行的注释。
	// 详情见 HeartbeatToDB 的文档。
	// go HeartbeatToDB(ctx)

	return s.Run(addr)
}

func (s *Server) initRender() *render.Render {
	tSets := []*render.TemplateSet{
		render.NewTemplateSet("spa", "spa.html", "spa.html"),
	}
	r := render.New(render.Options{
		Directory:     "templates",
		Funcs:         s.renderFuncMaps(),
		Delims:        render.Delims{"{{", "}}"},
		IndentJson:    true,
		UseBufPool:    true,
		IsDevelopment: s.isDebug,
	}, tSets)
	return r
}

func (s *Server) renderFuncMaps() []template.FuncMap {
	funcs := make([]template.FuncMap, 0)
	funcs = append(funcs, s.DefaultRouteFuncs())
	return funcs
}

func (s *Server) initDatabase(mctx *models.Context) {
	oauth2.InitDatabase(mctx)
	group.InitDatabase(mctx)
	app.InitDatabase(mctx)
	role.InitDatabase(mctx)
}

func (s *Server) adaptResourceHandler(handler server.ResourceHandler) server.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		code, data := handler(ctx, r)
		if code < 400 {
			s.renderJsonOr500(w, code, data)
		} else {
			errMessage := ""
			if msg, ok := data.(string); ok {
				errMessage = msg
			} else if msg, ok := data.(error); ok {
				errMessage = msg.Error()
			}
			switch code {
			case http.StatusMethodNotAllowed:
				if errMessage == "" {
					errMessage = fmt.Sprintf("Method %q is not allowed", r.Method)
				}
				s.renderError(w, code, errMessage, data)
			case http.StatusNotFound:
				if errMessage == "" {
					errMessage = "Cannot find the resource"
				}
				s.renderError(w, code, errMessage, data)
			case http.StatusBadRequest:
				if errMessage == "" {
					errMessage = "Invalid request get or post params"
				}
				s.renderError(w, code, errMessage, data)
			default:
				if errMessage == "" {
					errMessage = fmt.Sprintf("HTTP Error Code: %d", code)
				}
				s.renderError(w, code, errMessage, data)
			}
		}
		return ctx
	}
}

const (
	kContentCharset = "; charset=UTF-8"
	kContentJson    = "application/json"
)

func (s *Server) renderJson(w http.ResponseWriter, status int, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	data = append(data, '\n')
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", kContentJson+kContentCharset)
	w.WriteHeader(status)
	if status != http.StatusNoContent {
		_, err = w.Write(data)
	}
	return err
}

func (s *Server) renderJsonOr500(w http.ResponseWriter, status int, v interface{}) {
	if err := s.renderJson(w, status, v); err != nil {
		s.renderError(w, http.StatusInternalServerError, err.Error(), "")
	}
}

func (s *Server) renderHtmlTemplate(w http.ResponseWriter, tmpl *template.Template, v interface{}) {
	err := tmpl.Execute(w, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderError(w http.ResponseWriter, status int, msg string, data interface{}) {
	apiError := ApiError{msg, data}
	if err := s.renderJson(w, status, apiError); err != nil {
		log.Errorf("Server got a json rendering error, %s", err)
		// we fallback to the http.Error instead return a json formatted error
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderHtmlOr500(w http.ResponseWriter, status int, name string, binding interface{}) {
	if err := s.render.Html(w, status, name, binding); err != nil {
		log.Errorf("Server got a rendering error, %q", err)
		if s.isDebug {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "500, Internal Server Error", http.StatusInternalServerError)
		}
	}
}

type ApiError struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (s *Server) initOAuth2Provider(mctx *models.Context) (*osin.Server, error) {
	cfg := osin.NewServerConfig()
	cfg.AllowGetAccessRequest = true
	cfg.AllowClientSecretInParams = true
	cfg.AccessExpiration = 3600 * 24 * 15
	cfg.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN, TOKEN_IDTOKEN}
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN}

	storage := oauth2.NewOAuth2Storage(mctx)
	return osin.NewServer(cfg, storage), nil
}

func getOAuth2Provider(ctx context.Context) *osin.Server {
	return ctx.Value("oauth2").(*osin.Server)
}

func getDB(ctx context.Context) *sqlx.DB {
	return ctx.Value("db").(*sqlx.DB)
}

func getCurrentUser(ctx context.Context) iuser.User {
	u := ctx.Value("user")
	if u == nil {
		return nil
	}
	return u.(iuser.User)
}

func getScope(ctx context.Context) []string {
	scope := ctx.Value("scope")
	if scope == nil {
		return nil
	}
	return scope.([]string)
}

func getModelContext(ctx context.Context) *models.Context {
	return ctx.Value("mctx").(*models.Context)
}

func getEmailSuffix(ctx context.Context) string {
	return ctx.Value("emailSuffix").(string)
}

func getLegalNets(ctx context.Context) []*net.IPNet {
	return ctx.Value("legalNets").([]*net.IPNet)
}

func getUserBackend(ctx context.Context) iuser.UserBackend {
	return ctx.Value("userBackend").(iuser.UserBackend)
}
