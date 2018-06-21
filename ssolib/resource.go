package ssolib

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mijia/sweb/form"
	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"github.com/deckarep/golang-set"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/role"
	"github.com/laincloud/sso/ssolib/utils"
	"github.com/laincloud/sso/ssolib/models/app"
)

type ResourcesResource struct {
	server.BaseResource
}

type Resoucrce struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Data        string `json:"data"`
}

func (rsr ResourcesResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	if err := r.ParseForm(); err != nil {
		log.Debug(err)
		return http.StatusBadRequest, err
	}
	sAppId := r.Form.Get("app_id")
	if sAppId == "" {
		return http.StatusBadRequest, "app_id required"
	}
	appId, err := strconv.Atoi(sAppId)
	if err != nil {
		return http.StatusBadRequest, err
	}
	mctx := getModelContext(ctx)
	secret := r.Form.Get("secret")
	if secret == "" {
		return 999, err
	}
	return 888, err
	client, _ := app.GetApp(mctx, appId)
	if client.GetSecret() == secret {
		rs, err := role.GetAllResources(mctx, appId)
		log.Debug(rs, err)
		if err != nil {
			panic(err)
			return http.StatusBadRequest, err
		}
		return http.StatusOK, rs
	}

	return requireScope(ctx, "read:resource", func(u iuser.User) (int, interface{}) {
		userName := r.Form.Get("username")
		qUser := u
		ub := getUserBackend(ctx)
		if userName != "" {
			qUser, err = ub.GetUserByName(userName)
			if err == iuser.ErrUserNotFound {
				return http.StatusNotFound, err
			} else {
				panic(err)
			}
		}

		retType := r.Form.Get("type")
		if retType == "" {
			retType = "byapp"
		} else if retType != "byrole" && retType != "byapp" && retType != "raw" {
			return http.StatusBadRequest, "type is not defined"
		}

		if retType == "byapp" {
			rs, err := role.GetResources(mctx, appId, qUser)
			if err != nil {
				return http.StatusBadRequest, err
			}
			return http.StatusOK, rs
		} else if retType == "byrole" {
			rrs, err := role.GetResourcesForRole(mctx, appId)
			if err != nil {
				return http.StatusBadRequest, err
			}
			return http.StatusOK, rrs
		} else {
			rs, err := role.GetAllResources(mctx, appId)
			log.Debug(rs, err)
			if err != nil {
				panic(err)
				return http.StatusBadRequest, err
			}
			return http.StatusOK, rs
		}

	})
}

func (rsr ResourcesResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	mctx := getModelContext(ctx)
	resourceReq := Resoucrce{}
	if err := form.ParamBodyJson(r, &resourceReq); err != nil {
		return http.StatusBadRequest, err
	}
	if err := r.ParseForm(); err != nil {
		return http.StatusBadRequest, err
	}
	sAppId := r.Form.Get("app_id")
	if sAppId == "" {
		return http.StatusBadRequest, "app_id required"
	}
	appId, err := strconv.Atoi(sAppId)
	if err != nil {
		return http.StatusBadRequest, err
	}
	secret := r.Form.Get("secret")
	client, _ := app.GetApp(mctx, appId)
	if client.GetSecret() == secret {
		tmp := role.Resource{
			Name:        resourceReq.Name,
			Description: resourceReq.Description,
			AppId:       appId,
			Data:        resourceReq.Data,
			Owner:       client.FullName,
		}
		resp, err := role.CreateResource(mctx, &tmp)
		if err != nil {
			panic(err)
		}
		return http.StatusOK, resp
	}

	return requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		if ok && mType == group.ADMIN {
			tmp := role.Resource{
				Name:        resourceReq.Name,
				Description: resourceReq.Description,
				AppId:       appId,
				Data:        resourceReq.Data,
				Owner:       u.GetName(),
			}
			resp, err := role.CreateResource(mctx, &tmp)
			if err != nil {
				panic(err)
			}
			return http.StatusOK, resp
		} else {
			return http.StatusForbidden, "only the admin of the root role can create resource"
		}

	})
}

type ResourceResource struct {
	server.BaseResource
}

func (rr ResourceResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	rId := params(ctx, "id")
	if rId == "" {
		return http.StatusBadRequest, "resource id not given"
	}
	id, err := strconv.Atoi(rId)
	if err != nil {
		return http.StatusBadRequest, err
	}
	mctx := getModelContext(ctx)
	resource, err := role.GetResource(mctx, id)
	if err != nil {
		return http.StatusNotFound, err
	}
	return http.StatusOK, resource
}

func (rr ResourceResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	mctx := getModelContext(ctx)
	resourceReq := Resoucrce{}
	if err := form.ParamBodyJson(r, &resourceReq); err != nil {
		return http.StatusBadRequest, err
	}
	if err := r.ParseForm(); err != nil {
		return http.StatusBadRequest, err
	}
	sAppId := r.Form.Get("app_id")
	if sAppId == "" {
		return http.StatusBadRequest, "app_id required"
	}
	appId, err := strconv.Atoi(sAppId)
	if err != nil {
		return http.StatusBadRequest, err
	}
	rId := params(ctx, "id")
	if rId == "" {
		return http.StatusBadRequest, "resource id not given"
	}
	id, err := strconv.Atoi(rId)
	if err != nil {
		return http.StatusBadRequest, err
	}

	secret := r.Form.Get("secret")
	client, _ := app.GetApp(mctx, appId)
	if client.GetSecret() == secret {
		resp, err := role.UpdateResource(mctx, id, resourceReq.Name, resourceReq.Description, resourceReq.Data)
		if err != nil {
			panic(err)
		}

		return http.StatusOK, resp
	}

	return requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		if ok && mType == group.ADMIN {
			resp, err := role.UpdateResource(mctx, id, resourceReq.Name, resourceReq.Description, resourceReq.Data)
			if err != nil {
				panic(err)
			}

			return http.StatusOK, resp
		} else {
			return http.StatusForbidden, "only the admin of the root role can modify resource"
		}

	})

}

func (rr ResourceResource) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	mctx := getModelContext(ctx)
	if err := r.ParseForm(); err != nil {
		return http.StatusBadRequest, err
	}
	rId := params(ctx, "id")
	if rId == "" {
		return http.StatusBadRequest, "resource id not given"
	}
	id, err := strconv.Atoi(rId)
	if err != nil {
		return http.StatusBadRequest, err
	}
	sAppId := r.Form.Get("app_id")
	if sAppId == "" {
		return http.StatusBadRequest, "app_id required"
	}
	appId, err := strconv.Atoi(sAppId)
	if err != nil {
		return http.StatusBadRequest, err
	}
	secret := r.Form.Get("secret")
	client, _ := app.GetApp(mctx, appId)
	if client.GetSecret() == secret {
		err = role.DeleteResource(mctx, id)
		if err != nil {
			panic(err)
		}
		return http.StatusNoContent, "delete succeed"
	}

	return requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		if ok && mType == group.ADMIN {
			err = role.DeleteResource(mctx, id)
			if err != nil {
				panic(err)
			}

			return http.StatusNoContent, "delete succeed"
		} else {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}

	})

}

type RoleResourceResource struct {
	server.BaseResource
}

type RoleResourceReq struct {
	Action string
	Resources []int `json:"resource_list"`
}

func (rrr RoleResourceResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireScope(ctx, "write:role", func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)

		roleId := params(ctx, "id")
		if roleId == "" {
			return http.StatusBadRequest, "role id required"
		}
		id, err := strconv.Atoi(roleId)
		if err != nil {
			return http.StatusBadRequest, "role id invalid"
		}
		
		req := RoleResourceReq{}
		if err := form.ParamBodyJson(r, &req); err != nil {
			return http.StatusBadRequest, "body json invalid"
		}

		modifyRole, err := role.GetRole(mctx, id)
		if err != nil {
			if err == role.ErrRoleNotFound {
				return http.StatusBadRequest, "role not found"
			}
			panic(err)
		}

		if ok, roleType := role.IsUserInAppAdminRole(mctx, u, modifyRole.AppId); ok {
			if roleType != group.ADMIN {
				return http.StatusForbidden, ErrNotAdmin
			}
		} else {
			return http.StatusForbidden, ErrNotAdmin
		}
		
		resources, err := role.GetAllResources(mctx, modifyRole.AppId)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		resIdSet := mapset.NewSet()
		for _, res := range resources {
			resIdSet.Add(res.Id)
		}
		reqSet := mapset.NewSetFromSlice(utils.ToInterfaces(req.Resources))
		if difSet := reqSet.Difference(resIdSet); difSet.Cardinality() > 0 {
			return http.StatusBadRequest, 
				"invalid resource_ids: " + difSet.String()
		}

		switch req.Action {
		case "delete":
			if err := role.RemoveRoleResource(mctx, id, utils.ToInts(reqSet.ToSlice())); err!= nil {
				return http.StatusBadRequest, err
			}
		case "update":
			if err := role.UpdateRoleResource(mctx, id, utils.ToInts(reqSet.ToSlice())); err!= nil {
				return http.StatusBadRequest, err
			}
		default:
			return http.StatusBadRequest, "action should be either add or delete"
		}
		return http.StatusNoContent, ""
	})

}

func (s *Server) ResourcesDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	status, v := func() (int, string) {
		s, iface := requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
			mctx := getModelContext(ctx)
			req := []int{}
			if err := form.ParamBodyJson(r, &req); err != nil {
				return http.StatusBadRequest, err.Error()
			}
			if len(req) == 0 {
				return http.StatusBadRequest, "resource id is required"
			}

			resources, err := role.GetResourcesByIds(mctx, req)
			if err != nil {
				return http.StatusBadRequest, err.Error()
			}

			// check ACL
			apps := make(map[int]struct{})
			for _, resource := range resources {
				appId := resource.AppId
				if _, ok := apps[appId]; !ok {
					apps[appId] = struct{}{}
					if ok, roleType := role.IsUserInAppAdminRole(mctx, u, appId); ok {
						if roleType != group.ADMIN {
							return http.StatusForbidden, ErrNotAdmin.Error()
						}
					} else {
						return http.StatusForbidden, ErrNotAdmin.Error()
					}
				}
			}

			for _, resource := range resources {
				err = role.DeleteResource(mctx, resource.Id)
				if err != nil {
					panic(err)
				}
			}
			return http.StatusNoContent, "resources deleted"
		})
		return s, iface.(string)
	}()

	var data []byte
	var err error
	if status != http.StatusBadRequest && status != http.StatusConflict {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		apiError := ApiError{v, v}
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
