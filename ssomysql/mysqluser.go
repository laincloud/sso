package ssomysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mijia/sweb/form"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib"
	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssomysql/user"
)

const (
	//	ActivationString    string = "Please check activation email. \n 如果没有找到邮件，请查阅垃圾箱。"
	//	ResetPasswordString string = "please check email to continue.  找不到就去垃圾箱。"
	ActivationString    string = "请去邮箱查看激活邮件，如果没有找到邮件，请查阅垃圾箱。"
	ResetPasswordString string = "请去邮箱并继续，找不到邮件就去垃圾箱。"
)

//TODO 是否要将 ssolib 的 getUserBackend 和 getModelContext 变为包外可见？
func getModelContext(ctx context.Context) *models.Context {
	return ctx.Value("mctx").(*models.Context)
}

func getUserBackend(ctx context.Context) iuser.UserBackend {
	return ctx.Value("userBackend").(iuser.UserBackend)
}

type InactiveUsersResource struct {
	server.BaseResource
}

func (iur InactiveUsersResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {

	// FIXME 由于 ssolib 中的 requireScope 不可见，所以为了实现简单，
	// 这里对不符合 scope 的情况直接返回 403 而非重定向。
	scope := ctx.Value("scope")
	find := false
	if scope != nil {
		scopes := scope.([]string)
		for _, s := range scopes {
			if s == "write:user" {
				find = true
			}
		}
	}
	if !find {
		return http.StatusForbidden, `the scope "write:user" is required`
	}

	mctx := getModelContext(ctx)
	ub := getUserBackend(ctx)

	isCurrentUserAdmin := false
	adminsGroup, err := group.GetGroupByName(mctx, "admins")
	if err != nil {
		panic(err)
	}

	// bug FIXME should be all the group members, not only the direct Members.
	admins, err := adminsGroup.ListMembers(mctx)
	if err != nil {
		panic(err)
	}

	currentUser := ctx.Value("user").(iuser.User)

	for _, admin := range admins {
		if admin.GetId() == currentUser.GetId() {
			isCurrentUserAdmin = true
			break
		}
	}

	if !isCurrentUserAdmin {
		return http.StatusForbidden, "have no permission"
	}

	InactiveUsers, err := ub.(*user.UserBack).ListInactiveUsers(ctx)
	if err != nil {
		panic(err)
	}

	return http.StatusOK, InactiveUsers
}

func (iur InactiveUsersResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {

	// FIXME 由于 ssolib 中的 requireScope 不可见，所以为了实现简单，
	// 这里对不符合 scope 的情况直接返回 403 而非重定向。
	scope := ctx.Value("scope")
	find := false
	if scope != nil {
		scopes := scope.([]string)
		for _, s := range scopes {
			if s == "write:user" {
				find = true
			}
		}
	}
	if !find {
		return http.StatusForbidden, `the scope "write:user" is required`
	}

	mctx := getModelContext(ctx)
	ub := getUserBackend(ctx)

	isCurrentUserAdmin := false
	adminsGroup, err := group.GetGroupByName(mctx, "admins")
	if err != nil {
		panic(err)
	}
	admins, err := adminsGroup.ListMembers(mctx)
	if err != nil {
		panic(err)
	}

	currentUser := ctx.Value("user").(iuser.User)

	for _, admin := range admins {
		if admin.GetId() == currentUser.GetId() {
			isCurrentUserAdmin = true
			break
		}
	}

	if !isCurrentUserAdmin {
		return http.StatusForbidden, "have no permission"
	}

	err = ub.(*user.UserBack).DeleteAllActivationCodes(ctx)
	if err != nil {
		panic(err)
	}

	return http.StatusNoContent, "All Activation Codes have been deleted"
}

func UsersPost(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {

	status, v := func() (int, string) {
		reg := UserRegistration{}
		if err := form.ParamBodyJson(r, &reg); err != nil {
			return http.StatusBadRequest, err.Error()
		}

		if err := reg.Validate(ctx); err != nil {
			return http.StatusBadRequest, err.Error()
		}

		fullname := string(reg.FullName)
		if fullname == "" {
			fullname = string(reg.Name)
		}

		if _, err := user.RegisterUser(getModelContext(ctx), user.UserRegistration{
			Name:     string(reg.Name),
			FullName: fullname,
			Email:    sql.NullString{String: string(reg.Email), Valid: true},
			Password: reg.Password,
			Mobile:   sql.NullString{String: reg.Mobile, Valid: reg.Mobile != ""},
		}, getUserBackend(ctx)); err != nil {
			if err == user.ErrUserExists {
				return http.StatusConflict, err.Error()
			}
			log.Error(err)
			panic(err)
		}
		return http.StatusAccepted, ActivationString
	}()

	var data []byte
	var err error
	if status != http.StatusBadRequest && status != http.StatusConflict {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		apiError := ssolib.ApiError{v, v}
		data, err = json.MarshalIndent(apiError, "", "  ")
	}
	if err != nil {
		status = http.StatusInternalServerError
		data = []byte(err.Error())
	}
	data = append(data, '\n')
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	w.Write(data)
	return ctx
}

type UserRegistration struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Mobile   string `json:"mobile"`
}

func (ur *UserRegistration) Validate(ctx context.Context) error {
	if ur.Name == "" {
		return errors.New("Empty name")
	}
	if err := ssolib.ValidateSlug(ur.Name); err != nil {
		return err
	}

	if ur.FullName != "" {
		if err := ssolib.ValidateFullName(ur.FullName); err != nil {
			return err
		}
	}

	if ur.Email == "" {
		return errors.New("Empty email")
	}
	if err := ssolib.ValidateUserEmail(ur.Email, ctx); err != nil {
		return err
	}

	if ur.Password == "" {
		return errors.New("Empty password")
	}
	if len(ur.Password) < 4 {
		return errors.New("Password too short")
	}

	return nil
}

type ActivateUserResource struct {
	server.BaseResource
}

func (aur ActivateUserResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	code := form.ParamString(r, "code", "")
	if code == "" {
		return http.StatusBadRequest, "no code given"
	}

	u, err := user.ActivateUser(getModelContext(ctx), code, getUserBackend(ctx))
	if err != nil {
		if err == user.ErrCodeNotFound {
			return http.StatusNotFound, err.Error()
		}
		panic(err)
	}

	return http.StatusCreated, u.GetProfile()
}

type RequestResetPasswordResource struct {
	server.BaseResource
}

func (rrpr RequestResetPasswordResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	username := server.Params(ctx, "username")
	if username == "" {
		return http.StatusBadRequest, "username not given"
	}
	mctx := getModelContext(ctx)

	ub := getUserBackend(ctx)
	u, err := ub.GetUserByName(username)
	if err != nil {
		if err == iuser.ErrUserNotFound {
			return http.StatusNotFound, "no such user"
		}
		panic(err)
	}

	if err = user.RequestResetPassword(mctx, u.(*user.User)); err != nil {
		panic(err)
	}

	return http.StatusAccepted, ResetPasswordString
}

type RequestResetPasswordResourceByEmail struct {
	server.BaseResource
}

type RequestResetPasswordByEmail struct {
	Email string `json:"email"`
}

func (rrpr RequestResetPasswordResourceByEmail) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	postData := RequestResetPasswordByEmail{}
	if err := form.ParamBodyJson(r, &postData); err != nil {
		log.Debug(err)
		return http.StatusBadRequest, err.Error()
	}
	log.Debug(postData)
	email := postData.Email
	if email == "" {
		return http.StatusBadRequest, "email not given"
	}
	mctx := getModelContext(ctx)

	ub := getUserBackend(ctx)
	u, err := ub.(*user.UserBack).GetUserByEmail(email)
	if err != nil {
		if err == iuser.ErrUserNotFound {
			return http.StatusNotFound, "no such user"
		}
		panic(err)
	}

	if err = user.RequestResetPassword(mctx, u.(*user.User)); err != nil {
		panic(err)
	}

	return http.StatusAccepted, ResetPasswordString
}

type ResetPasswordResource struct {
	server.BaseResource
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
	Code     string `json:"code"`
}

func (rpr ResetPasswordResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {

	req := ResetPasswordRequest{}
	if err := form.ParamBodyJson(r, &req); err != nil {
		return http.StatusBadRequest, err
	}

	if req.Password == "" {
		return http.StatusBadRequest, "password is empty"
	}
	if len(req.Password) < 4 {
		return http.StatusBadRequest, "password too short"
	}

	if req.Code == "" {
		return http.StatusBadRequest, "no code given"
	}

	username := server.Params(ctx, "username")
	if username == "" {
		return http.StatusBadRequest, "no username"
	}

	mctx := getModelContext(ctx)

	ub := getUserBackend(ctx)
	u, err := ub.GetUserByName(username)
	if err != nil {
		if err == iuser.ErrUserNotFound {
			return http.StatusNotFound, "no such user"
		}
		panic(err)
	}

	if err = user.ResetPassword(mctx, u.(*user.User), req.Code, req.Password); err != nil {
		if err == user.ErrCodeNotFound {
			return http.StatusNotFound, "code not found"
		}
		panic(err)
	}

	return http.StatusNoContent, ""
}
