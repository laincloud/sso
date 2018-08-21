package ssoldap

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/laincloud/sso/ssoldap/user"
	"github.com/laincloud/sso/ssolib"
	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/mijia/sweb/form"
	"golang.org/x/net/context"
)

func UsersPost(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	// only admin can post
	status, v := func() (int, string) {
		info := user.UserInfo{}
		if err := form.ParamBodyJson(r, &info); err != nil {
			return http.StatusBadRequest, err.Error()
		}

		if err := info.Validate(ctx); err != nil {
			return http.StatusBadRequest, err.Error()
		}

		fullname := info.FullName
		if fullname == "" {
			info.FullName = info.Name
		}

		if info.Email[0:strings.Index(info.Email, "@")] != info.Name {
			return http.StatusBadRequest, "user name should be prefix of email"
		}

		mctx := ctx.Value("mctx").(*models.Context)
		u := ctx.Value("user")
		if u == nil {
			return http.StatusUnauthorized, "only admin can create or update users"
		}
		currentUser := u.(iuser.User)
		isAdmin := false
		if currentUser != nil {
			adminsGroup, err := group.GetGroupByName(mctx, "admins")
			if err != nil {
				panic(err)
			}
			isAdmin, _, _ = adminsGroup.GetMember(mctx, currentUser)
		}

		if currentUser == nil || !isAdmin {
			return http.StatusUnauthorized, "only admin can create or update users"
		} else {
			ub := ctx.Value("userBackend").(*user.UserBack)
			err := ub.AddUser(info)
			if err != nil {
				return http.StatusBadRequest, err.Error()
			}
			return http.StatusAccepted, ""
		}

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
