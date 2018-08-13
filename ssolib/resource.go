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

func getResourceByClient(secret string, mctx *models.Context, app *app.App, retType string) (int, interface{}) {
	if app.GetSecret() != secret {
		return http.StatusForbidden, "authorization is required"
	}
	switch retType {
	case "", "byrole":
		{
			rrs, err := role.GetResourcesForRole(mctx, app.Id)
			if err != nil {
				return http.StatusBadRequest, err
			}
			return http.StatusOK, rrs
		}
	case "raw" :
		{
			rs, err := role.GetAllResources(mctx, app.Id)
			log.Debug(rs, err)
			if err != nil {
				panic(err)
				return http.StatusBadRequest, err
			}
			return http.StatusOK, rs
		}
	default : return http.StatusBadRequest, "type is not defined"
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
	theApp, err := app.GetApp(mctx,appId)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if theApp == nil {
		return http.StatusBadRequest, "invaild app_id"
	}
	retType := r.Form.Get("type")
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		err := requireScope(ctx, "read:resource")
		if err != nil {
			return http.StatusUnauthorized, err
		}
		u := getCurrentUser(ctx)
		switch retType {
		case "", "byapp":
			{
				rs, err := role.GetResources(mctx, appId, u)
				if err != nil {
					return http.StatusBadRequest, err
				}
				return http.StatusOK, rs
			}
		case "byrole":
			{
				rrs, err := role.GetResourcesForRole(mctx, appId)
				if err != nil {
					return http.StatusBadRequest, err
				}
				return http.StatusOK, rrs
			}
		case "raw":
			{
				rs, err := role.GetAllResources(mctx, appId)
				log.Debug(rs, err)
				if err != nil {
					panic(err)
					return http.StatusBadRequest, err
				}
				return http.StatusOK, rs
			}
		default : return http.StatusBadRequest, "type is not defined"
		}
	} else {
		return getResourceByClient(secret, mctx, theApp, retType)
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
	//if using secret, resource owner name is app name, otherwise the name is user's name.
	var name string
	if secret == "" {
		err := requireScope(ctx, "write:resource")
		if err != nil {
			return http.StatusUnauthorized, err
		}
		u := getCurrentUser(ctx)
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		name = u.GetName()
		if !(ok && mType == group.ADMIN) {
			return http.StatusForbidden, "only the admin of the root role can create resource"
		}
	} else {
		theApp, err := app.GetApp(mctx, appId)
		if err != nil {
			return http.StatusUnauthorized, err
		}
		name = theApp.FullName
		if theApp.GetSecret() != secret {
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
	rId := params(ctx, "id")
	if rId == "" {
		return http.StatusBadRequest, "resource id not given"
	}
	id, err := strconv.Atoi(rId)
	if err != nil {
		return http.StatusBadRequest, err
	}
	resource, err := role.GetResource(mctx, id)
	if err != nil {
		return http.StatusBadRequest, err
	}
	appId := resource.AppId
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		err := requireScope(ctx, "write:resource")
		if err != nil {
			return http.StatusUnauthorized, err
		}
		u := getCurrentUser(ctx)
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		if !(ok && mType == group.ADMIN) {
			return http.StatusForbidden, "only the admin of the root role can modify resource"
		}
	} else {
		theApp, err := app.GetApp(mctx, appId)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if theApp.GetSecret() != secret {
			return http.StatusForbidden, "only the admin of the root role can modify resource"
		}
	}
	resp, err := role.UpdateResource(mctx, id, resourceReq.Name, resourceReq.Description, resourceReq.Data)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, resp
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
	resource, err := role.GetResource(mctx, id)
	if err != nil {
		return http.StatusBadRequest, err
	}
	appId := resource.AppId
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		err := requireScope(ctx, "write:resource")
		if err != nil {
			return http.StatusUnauthorized, err
		}
		u := getCurrentUser(ctx)
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		if !(ok && mType == group.ADMIN) {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}
	} else {
		theApp, err := app.GetApp(mctx, appId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		if theApp.GetSecret() != secret {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}
	}
	err = role.DeleteResource(mctx, id)
	if err != nil {
		return http.StatusInternalServerError, err
	} else {
		return http.StatusNoContent, "delete succeed"
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
		err := requireScope(ctx, "write:role")
		if err != nil {
			return http.StatusUnauthorized, err
		}
		u := getCurrentUser(ctx)
		ok, mType := role.IsUserInAppAdminRole(mctx, u, appId)
		if !(ok && mType == group.ADMIN) {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}
	} else {
		theApp, err := app.GetApp(mctx, appId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		if theApp.GetSecret() != secret {
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
	case "add":
		if err := role.AddRoleResource(mctx, id, utils.ToInts(reqSet.ToSlice())); err!= nil {
			return http.StatusBadRequest, err
		}
		return  http.StatusOK, nil
	case "delete":
		if err := role.RemoveRoleResource(mctx, id, utils.ToInts(reqSet.ToSlice())); err!= nil {
			return http.StatusBadRequest, err
		}
		return http.StatusNoContent, nil
	case "update":
		if err := role.UpdateRoleResource(mctx, id, utils.ToInts(reqSet.ToSlice())); err!= nil {
			return http.StatusBadRequest, err
		}
		return  http.StatusOK, nil
	default:
		return http.StatusBadRequest, "action should be add, delete or update"
	}
}

func deleteResources(ctx context.Context,r *http.Request) (int, string) {
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
			return http.StatusForbidden, "resources must belong to the same app"
		}
	}
	secret := r.Header.Get("secret")
	//two ways of authorization. If secret is not given, check user, otherwise check secret.
	if secret == "" {
		err := requireScope(ctx, "write:role")
		if err != nil {
			return http.StatusUnauthorized, err.Error()
		}
		u := getCurrentUser(ctx)
		ok, mType := role.IsUserInAppAdminRole(mctx, u, AppId)
		if !(ok && mType == group.ADMIN) {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}
	} else {
		theApp, err := app.GetApp(mctx, AppId)
		if err != nil {
			return http.StatusBadRequest, err.Error()
		}
		if theApp.GetSecret() != secret {
			return http.StatusForbidden, "only the admin member of the root role can delete resources"
		}
	}
	deleted := ""
	for _, resource := range resources {
		err = role.DeleteResource(mctx, resource.Id)
		if err != nil {
			return http.StatusBadRequest, deleted
		}
		deleted = deleted + " " + strconv.Itoa(resource.Id)
	}
	return http.StatusNoContent, "all resources deleted"
}

func (s *Server) ResourcesDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	status, v := deleteResources(ctx, r)
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
