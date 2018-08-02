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

	"github.com/laincloud/sso/ssolib/models"
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

func GetResourceByClient (secret string, mctx *models.Context, appId int, retType string) (int, interface{}) {
	client, err := app.GetApp(mctx, appId)
	if err != nil || client.GetSecret() != secret {
		return http.StatusForbidden, "authorization is required"
	}
	if retType == "" {
		retType = "byrole"
	} else if retType != "byrole" && retType != "raw" {
		return http.StatusBadRequest, "type is not defined"
	}
	if retType == "byrole" {
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
	retType := r.Form.Get("type")
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		return requireScope(ctx, "read:resource", func(u iuser.User) (int, interface{}) {
			if retType == "" {
				retType = "byapp"
			} else if retType != "byrole" && retType != "byapp" && retType != "raw" {
				return http.StatusBadRequest, "type is not defined"
			}

			if retType == "byapp" {
				rs, err := role.GetResources(mctx, appId, u)
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
	} else {
		return GetResourceByClient(secret, mctx, appId, retType)
	}
}


func (rsr ResourcesResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
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
	mctx := getModelContext(ctx)
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	//if using secret, resource owner name is client name, otherwise the name is user's name.
	var name string
	if secret == "" {
		auth, msg:= requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
			ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
			name = u.GetName()
			if !(ok && mType == group.ADMIN) {
				return http.StatusForbidden, "only the admin of the root role can create resource"
			}
			return http.StatusOK, "authorized"
		})
		if auth != http.StatusOK {
			return auth, msg
		}
	} else {
		client, err := app.GetApp(mctx, appId)
		name = client.FullName
		if err != nil || client.GetSecret() != secret {
			return http.StatusForbidden, "only the admin of the root role can create resource"
		}
	}
	tmp := role.Resource{
		Name:        resourceReq.Name,
		Description: resourceReq.Description,
		AppId:       appId,
		Data:        resourceReq.Data,
		Owner:       name,
	}
	resp, err := role.CreateResource(mctx, &tmp)
	if err != nil {
		panic(err)
	}
	return http.StatusOK, resp
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
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		auth, msg:= requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
			ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
			if !(ok && mType == group.ADMIN) {
				return http.StatusForbidden, "only the admin of the root role can modify resource"
			}
			return http.StatusOK, "authorized"
		})
		if auth != http.StatusOK {
			return auth, msg
		}
	} else {
		client, err := app.GetApp(mctx, appId)
		if err != nil || client.GetSecret() != secret {
			return http.StatusForbidden, "only the admin of the root role can modify resource"
		}
	}
	currentId := []int{id}
	Resources,err := role.GetResourcesByIds(mctx, currentId)
	resource := Resources[0]
	resource_appId := resource.AppId
	if err == nil && resource_appId == appId {
		resp, err := role.UpdateResource(mctx, id, resourceReq.Name, resourceReq.Description, resourceReq.Data)
		if err != nil {
			panic(err)
		}
		return http.StatusOK, resp
	} else {
		return http.StatusBadRequest, "cannot update other app's resource"
	}
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
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		auth, msg := requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
			ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
			if !(ok && mType == group.ADMIN) {
				return http.StatusForbidden, "only the admin member of the root role can delete resources"
			}
			return http.StatusOK, "authorized"
		})
		if auth != http.StatusOK {
			return auth, msg
		}
	} else {
		client, err := app.GetApp(mctx, appId)
		if err != nil || client.GetSecret() != secret {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}
	}
	currentId := []int{id}
	Resources,err := role.GetResourcesByIds(mctx, currentId)
	resource := Resources[0]
	resource_appId := resource.AppId
	if err == nil && resource_appId == appId {
		err = role.DeleteResource(mctx, id)
		if err != nil {
			panic(err)
		}
		return http.StatusNoContent, "delete succeed"
	} else {
		return http.StatusBadRequest, "cannot update other app's resource"
	}
}

type RoleResourceResource struct {
	server.BaseResource
}

type RoleResourceReq struct {
	Action string
	Resources []int `json:"resource_list"`
}

func (rrr RoleResourceResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {

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
	appId := modifyRole.AppId
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		auth, msg := requireScope(ctx, "write:role", func(u iuser.User) (int, interface{}) {
			if ok, roleType := role.IsUserInAppAdminRole(mctx, u, modifyRole.AppId); ok {
				if roleType != group.ADMIN {
					return http.StatusForbidden, ErrNotAdmin
				}
			} else {
				return http.StatusForbidden, ErrNotAdmin
			}
			return http.StatusOK, "authorized"
		})
		if auth != http.StatusOK {
			return auth, msg
		}
	} else {
		client, err := app.GetApp(mctx, appId)
		if err != nil || client.GetSecret() != secret {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}
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
}

func (s *Server) ResourcesDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	status, v := func() (int, string) {
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
		AppId := resources[0].AppId
		// check if the resources to be deleted belong to the app
		for _, resource := range resources {
			appId := resource.AppId
			if appId != AppId {
				return http.StatusForbidden, "cannot delete other app's resource"
			}
		}
		secret := r.Header.Get("secret")
		//two ways of authorization. If secret is not given, check user, otherwise check secret.
		if secret == "" {
			auth, _ := requireScope(ctx, "write:resource", func(u iuser.User) (int, interface{}) {
				if ok, roleType := role.IsUserInAppAdminRole(mctx, u, AppId); ok {
					if roleType != group.ADMIN {
						return http.StatusForbidden, ErrNotAdmin
					}
				} else {
					return http.StatusForbidden, ErrNotAdmin
				}
				return http.StatusOK, "authorized"
			})
			if auth != http.StatusOK {
				return auth, "authorization is required"
			}
		} else {
			client, err := app.GetApp(mctx, AppId)
			if err != nil || client.GetSecret() != secret {
				return http.StatusForbidden, "only the admin member of the root role can delete resources"
			}
		}
		deleted := ""
		for _, resource := range resources {
			err = role.DeleteResource(mctx, resource.Id)
			if err != nil {
				return http.StatusBadRequest, deleted
				panic(err)
			}
			deleted = deleted + " " + strconv.Itoa(resource.Id)
		}
		return http.StatusNoContent, "all resources deleted"
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
