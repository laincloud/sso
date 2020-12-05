package ssolib

import (
	"net/http"
	"strconv"

	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/mijia/sweb/form"
	"github.com/laincloud/sso/ssolib/models/application"
	"github.com/laincloud/sso/ssolib/models/role"
	"errors"
	"github.com/laincloud/sso/ssolib/models/app"
)

type ApplicationsResource struct {
	server.BaseResource
}

func createApplication(ctx context.Context, id int, newApp *application.Application)(*application.Application, error) {
	mctx := getModelContext(ctx)
	adminIds, err := group.GetAdminIdsOfGroup(mctx, id)
	if err != nil {
		return nil, err
	}
	if adminIds == nil {
		return nil, errors.New("no available admins")
	}
	if !(newApp.TargetType == "group" || newApp.TargetType == "role") {
		return nil, errors.New("no such target type")
	}
	adminEmails := []string{}
	batchEmails := ""
	//send emails and create applications
	for _, id := range adminIds {
		admin, err := mctx.Back.GetUser(id)
		if err != nil {
			return nil, err
		}
		adminEmail := admin.GetProfile().GetEmail()
		adminEmails = append(adminEmails, adminEmail)
		batchEmails = batchEmails + adminEmail + " "
	}
	res, err := application.CreateApplication(mctx, newApp, adminEmails)
	if err != nil {
		return nil, err
	}
	var name string
	if newApp.TargetType == "group" {
		name = "group: " + res.TargetContent.Name
	} else {
		name = "role " + res.TargetContent.AppName + " " + res.TargetContent.Name
	}
	user := getCurrentUser(ctx)
	content := user.GetName() + "applies to join" + name + "\n" + "https://text/api/applications/" + strconv.Itoa(res.Id)
	err = SendTo("new application", content, batchEmails)
	if err != nil {
		log.Debug(err)
	}
	return res, nil
}


type ApplicationReq struct {
	Target      []application.TargetContent `json:"target"`
	TargetType  string               `json:"target_type"`
	Reason      string               `json:"reason"`
}


func checkIfQualified(ctx *models.Context, gId int, targetRole string, u iuser.User) (bool, error) {
	if group.CheckIfInGroup(ctx, gId, u.GetId()) {
		role, err := group.GetRoleOfUser(ctx, u.GetId(), gId)
		if err != nil {
			return false, err
		}
		if role == group.NORMAL && targetRole == "admin" {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

type RespWithLen struct {
	Resp []application.Application `json:"applications"`
	Total int `json:"total"`
}

func getApplicationsForCommiting(ctx context.Context, from int, to int) (int, interface{}) {
	Applications := []application.Application{}
	mctx := getModelContext(ctx)
	currentEmail := getCurrentUser(ctx).GetProfile().GetEmail()
	pendingApplications, num, err := application.GetPendingApplicationByEmail(mctx, currentEmail, from, to)
	if err != nil {
		return http.StatusBadRequest, err
	}
	for _, pa := range pendingApplications {
		a, err := application.GetApplication(mctx, pa.ApplicationId)
		if err != nil {
			return http.StatusBadRequest, err
		}
		Applications = append(Applications, *a)
	}
	respWithLen := RespWithLen{
		Resp: Applications,
		Total: num,
	}
	return http.StatusOK, respWithLen
}

func parseCommitEmails(mctx * models.Context, applications []application.Application, total int) (int, interface{}) {
	resp := []application.Application{}
	for _,a := range applications {
		operatorList := []string{}
		if a.Status == "initialled" {
			pendApplications, err := application.GetPendingApplicationByApplicationId(mctx, a.Id)
			if err != nil {
				return http.StatusBadRequest, err
			}
			for _, p := range pendApplications {
				operatorList = append(operatorList, p.OperatorEmail)
			}
			a.ParseOprEmail(operatorList)
		}
		resp = append(resp, a)
	}
	respWithLen := RespWithLen{
		Resp: resp,
		Total: total,
	}
	return http.StatusOK, respWithLen
}


func (ay ApplicationsResource) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	err := requireLogin(ctx)
	if err != nil {
		return http.StatusUnauthorized, err
	}
	currentUser := getCurrentUser(ctx)
	r.ParseForm()
	mctx := getModelContext(ctx)
	applicantEmail := r.Form.Get("applicant_email")
	status := r.Form.Get("status")
	fromStr := r.Form.Get("from")
	toStr := r.Form.Get("to")
	from := 0
	to := 50
	applications := []application.Application{}
	var total int
	if fromStr != "" && toStr != "" {
		start, err := strconv.Atoi(fromStr)
		if err != nil {
			return http.StatusBadRequest, err
		}
		end, err := strconv.Atoi(toStr)
		if err != nil {
			return http.StatusBadRequest, err
		}
		if !(start >= 0 && start <= end) {
			from = 0
			to = 50
		} else {
			from = start
			to = end
		}
	}
	currentEmail := currentUser.GetProfile().GetEmail()
	commitEmail := r.Form.Get("commit_email")
	if commitEmail == currentEmail {
		log.Debug(commitEmail)
		return getApplicationsForCommiting(ctx, from, to)
	} else if commitEmail == "" {
		islain := false
		laingroup, err := group.GetGroupByName(mctx, "lain")
		if laingroup != nil && err == nil {
			islain = group.CheckIfInGroup(mctx, laingroup.Id, currentUser.GetId())
		}
		//lain member checks all applications
		if islain && applicantEmail == "" {
			apps, num, err := application.GetAllApplications(mctx, status, from, to)
			if err != nil {
				return http.StatusBadRequest, err
			}
			applications = apps
			total = num
			//user checks his applications
		} else if applicantEmail == "" || applicantEmail == currentEmail{
			apps, num, err := application.GetApplications(mctx, currentEmail, status, from, to)
			if err != nil {
				return http.StatusBadRequest, err
			}
			applications = apps
			total = num
			//lain member checks some user's applications
		} else if islain && applicantEmail != currentEmail {
			apps, num, err := application.GetApplications(mctx, applicantEmail, status, from, to)
			if err != nil {
				return http.StatusBadRequest, err
			}
			applications = apps
			total = num
		} else {
			return http.StatusBadRequest, "not qualified for the operation"
		}
		return parseCommitEmails(mctx, applications, total)
	}
	return http.StatusBadRequest, "commitEmail is not vaild"
}

func createApplicationForGroup(ctx context.Context, req ApplicationReq) (int, interface{}) {
	mctx := getModelContext(ctx)
	if len(req.Target) < 1 {
		return http.StatusBadRequest, "no enough target"
	}
	groupIds := []int{}
	for i := 0; i < len(req.Target); i++ {
		g, err := group.GetGroup(mctx, req.Target[i].Id)
		if err != nil {
			if err == group.ErrGroupNotFound {
				return http.StatusBadRequest, err
			}
			return http.StatusInternalServerError, err
		}
		groupIds = append(groupIds,g.Id)
		req.Target[i].Name = g.Name
	}
	return checkQualifiedAndCreateApplications(ctx, req, groupIds)
}

func createApplicationForRole(ctx context.Context, req ApplicationReq) (int, interface{}) {
	mctx := getModelContext(ctx)
	if len(req.Target) < 1 {
		return http.StatusBadRequest, "no enough target"
	}
	roleIds := []int{}
	for i := 0; i < len(req.Target); i++ {
		theRole, err := role.GetRole(mctx, req.Target[i].Id)
		if err != nil {
			if err == role.ErrRoleNotFound {
				return http.StatusBadRequest, err
			}
			return http.StatusInternalServerError, err
		}
		_, err = group.GetGroup(mctx, req.Target[i].Id)
		if err != nil {
			if err == group.ErrGroupNotFound {
				return http.StatusBadRequest, err
			}
			return http.StatusInternalServerError, err
		}
		theApp, err := app.GetApp(mctx, theRole.AppId)
		if err != nil {
			if err == app.ErrAppNotFound {
				return http.StatusBadRequest, err
			}
			return http.StatusInternalServerError, err
		}
		req.Target[i].Name = theRole.Name
		req.Target[i].AppName = theApp.FullName
		roleIds = append(roleIds, req.Target[i].Id)
	}
	return checkQualifiedAndCreateApplications(ctx, req, roleIds)
}

func checkQualifiedAndCreateApplications(ctx context.Context, req ApplicationReq, targets []int) (int, interface{}) {
	mctx := getModelContext(ctx)
	applicant := getCurrentUser(ctx)
	var applications []application.Application
	for i := 0; i < len(req.Target); i++ {
		qualified, err := checkIfQualified(mctx, targets[i], req.Target[i].Role, applicant)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if !qualified {
			temp := application.Application{
				TargetType:    req.TargetType,
				TargetContent: &req.Target[i],
				Status:        "existed",
			}
			applications = append(applications, temp)
		} else {
			newApp :=&application.Application{
				Reason: req.Reason,
				TargetContent: &req.Target[i],
				TargetType: req.TargetType,
				ApplicantEmail: applicant.GetProfile().GetEmail(),
			}
			resp, err := createApplication(ctx, targets[i], newApp)
			if err == nil && resp != nil{
				applications = append(applications, *resp)
			} else {
				log.Debug(err)
				return http.StatusBadRequest, applications
			}
		}
	}
	return http.StatusOK, applications
}



func (ay ApplicationsResource) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	err := requireLogin(ctx)
	if err != nil {
		return http.StatusUnauthorized, err
	}
	req := ApplicationReq{}
	if err := form.ParamBodyJson(r, &req); err != nil {
		return http.StatusBadRequest, err
	}
	if req.TargetType == "role" {
		return createApplicationForRole(ctx, req)
	}
	if req.TargetType == "group" {
		return createApplicationForGroup(ctx, req)
	}
	return http.StatusBadRequest, "no such type"
}



type ApplicationResource struct {
	server.BaseResource
}

func (ah ApplicationResource) Put(ctx context.Context, r *http.Request) (int, interface{}) {
	err := requireLogin(ctx)
	if err != nil {
		return http.StatusUnauthorized, err
	}
	currentUser := getCurrentUser(ctx)
	mctx := getModelContext(ctx)
	aid := params(ctx, "id")
	r.ParseForm()
	action := r.Form.Get("action")
	aId, err := strconv.Atoi(aid)
	if err != nil {
		return http.StatusBadRequest, err
	}
	currentApp, err := application.GetApplication(mctx, aId)
	if err != nil {
		log.Debug(err)
		return http.StatusBadRequest, err
	}
	switch action {
	case "recall":
		{
			if currentApp.ApplicantEmail != currentUser.GetProfile().GetEmail() {
				return http.StatusBadRequest, "only applicant can delete application"
			}
			err = application.RecallApplication(mctx, aId)
			if err != nil {
				return http.StatusBadRequest, err
			}
			return http.StatusNoContent, "application deleted"
		}
	case "approve", "reject":
		{
			targetGroup, err := group.GetGroup(mctx, currentApp.TargetContent.Id)
			if err != nil {
				if err == group.ErrGroupNotFound {
					return http.StatusBadRequest, err
				}
				return http.StatusInternalServerError, err
			}
			commitRole, err := group.GetRoleOfUser(mctx, currentUser.GetId(), targetGroup.Id)
			if err != nil {
				log.Debug(err)
				return http.StatusBadRequest, err
			}
			if commitRole != group.ADMIN {
				return http.StatusBadRequest, "not qualified for the operation"
			}
			applicant, err := mctx.Back.GetUserByEmail(currentApp.ApplicantEmail)
			if err != nil {
				return http.StatusBadRequest, err
			}
			if action == "approve" {
				var role group.MemberRole
				if currentApp.TargetContent.Role == "admin" {
					role = group.ADMIN
				} else {
					role = group.NORMAL
				}
				log.Debug("adding member")
				err := targetGroup.AddMember(mctx, applicant, role)
				if err != nil {
					return http.StatusBadRequest, err
				}
				log.Debug("finishing handing application")
				resp, err := application.FinishApplication(mctx, currentApp.Id, "approved", currentUser.GetProfile().GetEmail())
				if err != nil {
					return http.StatusBadRequest, err
				}
				return http.StatusOK, resp
			} else if action == "reject" {
				log.Debug("finishing handing application")
				resp, err := application.FinishApplication(mctx, currentApp.Id, "rejected", currentUser.GetProfile().GetEmail())
				if err != nil {
					return http.StatusBadRequest, err
				}
				return http.StatusOK, resp
			}

		}
	default:
		return http.StatusBadRequest, "no such operation"
	}
	return http.StatusInternalServerError, "dummy message"
}

