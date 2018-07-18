package ssolib

import (
	"net/http"
	"strconv"

	"github.com/mijia/sweb/log"
	"github.com/mijia/sweb/server"
	"golang.org/x/net/context"

	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/models/role"
	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/group"
	"github.com/mijia/sweb/form"
)



type Apply struct {
	server.BaseResource
}

func applyEnterGroup(mctx *models.Context, id int, target role.TargetContent, email string, reason string)(*role.Application, error) {
	res, err := role.CreateApplication(mctx, email, "group", &target, reason)
	if err != nil {
		return nil, err
	}
	rows, err := mctx.DB.Query("SELECT user_id FROM user_group WHERE group_id=? AND role=?", id, 1)
	if err != nil {
		return nil, err
	}
	log.Debug(rows)
	back := mctx.Back
	admin_num := 0
	for rows.Next() {
		var userId int
		if err = rows.Scan(&userId); err == nil {
			log.Debug(userId)
			user, err := back.GetUser(userId)
			if err == nil {
				adminEmail := user.GetProfile().GetEmail()
				log.Debug(adminEmail)
				admin_num++
				role.CreatePendingApplication(mctx, res.Id, adminEmail)
			}
		}
	}
	if admin_num > 0 {
		resp, err := role.UpdateApplication(mctx, res.Id,"emails sent", "NULL")
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return res, nil
}



type Application struct {
	Description string               `json:"description"`
	Target      []role.TargetContent `json:"target"`
	TargetType  string               `json:"target_type"`
	Reason      string               `json:"reason"`
}



func GetTypeOfUser(ctx *models.Context, id int, groupId int) (int, error){
	role := []int{}
	err := ctx.DB.Get(&role, "SELECT role FROM user_group WHERE user_id=? AND group_id=?", id, groupId)
	log.Debug(err)
	if err != nil {
		return -1, err
	}
	return role[0], nil
}


func CheckIfInGroup(group_ids []int, id int) (bool) {
	left := 0;
	right := len(group_ids)
	Exist := false
	for left <= right {
		mid := left + (right-left)/2
		if group_ids[mid] == id {
			Exist = true
			break
		} else if group_ids[mid] > id {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return Exist
}

func CheckIfQualified(ctx *models.Context, group_ids []int, id int, target role.TargetContent, u iuser.User) (bool, error) {
	if CheckIfInGroup(group_ids, id) {
		role, err := GetTypeOfUser(ctx, u.GetId(), id)
		if err != nil {
			return false, err
		}
		if role == 0 && target.MemberType == "admin" {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

type ApplicationResp struct {
	ApplicationList []*role.Application `json:"application_list"`
	Alreadyin       []string            `json:"alreadyin"`
	Err             error               `json:"err"`
}


func (ay Apply) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		log.Debug("apply starting")
		req := Application{}
		if err := form.ParamBodyJson(r, &req); err != nil {
			return http.StatusBadRequest, err
		}
		email := u.GetProfile().GetEmail()
		if req.TargetType == "group" {
			groupIds, err := getDirectGroupsOfUser(mctx, u)
			if err != nil {
				return http.StatusBadRequest, err
			}
			if len(req.Target) == 1 {
				group, err := group.GetGroupByName(mctx, req.Target[0].GroupName)
				if err != nil {
					return http.StatusBadRequest, err
				}
				id := group.Id
				qualified, err := CheckIfQualified(mctx, groupIds, id, req.Target[0], u)
				if err != nil {
					return http.StatusBadRequest, err
				}
				if qualified {
					resp, err := applyEnterGroup(mctx, id, req.Target[0], email, req.Reason)
					if err != nil {
						return http.StatusBadRequest, err
					}
					return http.StatusOK, resp
				}
				return http.StatusBadRequest, "already in the group!"
			} else if len(req.Target) > 1 {
				var application_list []*role.Application
				var alreadyin []string
				var err1 error
				for i := 0; i < len(req.Target); i++ {
					group, err := group.GetGroupByName(mctx, req.Target[i].GroupName)
					if err != nil {
						err1 = err
						break
					}
					id := group.Id
					qualified, err := CheckIfQualified(mctx, groupIds, id, req.Target[i], u)
					if err != nil {
						err1 = err
						break
					}
					if qualified {
						resp, err := applyEnterGroup(mctx, id, req.Target[i], email, req.Reason)
						if err != nil {
							err1 = err
							break
						}
						application_list = append(application_list, resp)
					} else {
						alreadyin = append(alreadyin, req.Target[i].GroupName)
					}
				}
				temp := ApplicationResp{
					application_list ,
					alreadyin ,
					err1 ,
				}
				if err1 != nil {
					return http.StatusBadRequest, temp
				}
				return http.StatusOK, temp
			}
		}
		return http.StatusBadRequest, "no such type"
	})
}


func getDirectGroupsOfUser(ctx *models.Context, user iuser.User) ([]int, error) {
	groupIds := []int{}
	log.Debug("getgroupsofuser")
	err := ctx.DB.Select(&groupIds,
		"SELECT group_id FROM user_group WHERE user_id=?",
		user.GetId())
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	for _, gid := range groupIds {
		_, err := group.GetGroup(ctx, gid)
		if err != nil {
			if err != group.ErrGroupNotFound {
				return nil, err
			} else {
				log.Warnf("User %s in a non-exist group %d?", user.GetName(), gid)
			}
		}
	}
	return groupIds, nil
}


type ApplicationStatus struct {
	server.BaseResource
}
type applicationStatus struct {
	Id           int                 `json:"id"`
	TargetType   string              `json:"target_type"`
	Target       *role.TargetContent `json:"target"`
	Status       string              `json:"status"`
	OperatorList []string            `json:"opr_emails"`
}

func (as ApplicationStatus) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		r.ParseForm()
		From := r.Form.Get("from")
		To := r.Form.Get("to")
		from := 0
		to := 50
		if From != "" && To != "" {
			tfrom, err := strconv.Atoi(From)
			if err != nil {
				return http.StatusBadRequest, err
			}
			tto, err := strconv.Atoi(To)
			if err != nil {
				return http.StatusBadRequest, err
			}
			if !(tfrom >= 0 && tto > 0 && tfrom <= tto) {
				from = 0
				to = 50
			} else {
				from = tfrom
				to = tto
			}
		}

		Status := r.Form.Get("status")
		applications, err := role.GetApplications(mctx, u.GetProfile().GetEmail(), Status, from, to)
		log.Debug(applications)
		if err != nil {
			return http.StatusBadRequest, err
		}
		resp := []applicationStatus{}
		for _, a := range applications {
			operatorList := []string{}
			if a.Status == "emails sent" {
				pend_applications, err := role.GetPendingApplicationByApplicationId(mctx, a.Id)
				log.Debug(pend_applications)
				if err != nil {
					return http.StatusBadRequest, err
				}
				for _, p := range pend_applications {
					operatorList = append(operatorList, p.OperatorEmail)
				}
			} else {
				operatorList = append(operatorList, a.CommitorEmail)
			}
			temp := applicationStatus{
				Id:           a.Id,
				TargetType:   a.TargetType,
				Target:       a.Target,
				Status:       a.Status,
				OperatorList: operatorList,
			}
			resp = append(resp, temp)
		}
		return http.StatusOK, resp
	})
}




type ApplicationQuery struct {
	server.BaseResource
}

func (aq ApplicationQuery) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		laingroup, err := group.GetGroupByName(mctx, "lain")
		islain := false
		if laingroup != nil && err == nil {
			id := laingroup.Id
			usergroup, err := getDirectGroupsOfUser(mctx, u)
			if err == nil && usergroup != nil {
				islain = CheckIfInGroup(usergroup, id)
			}
		}
		if islain {
			r.ParseForm()
			email := r.Form.Get("applicant_email")
			Status := r.Form.Get("status")
			From := r.Form.Get("from")
			To := r.Form.Get("to")
			from := 0
			to := 50
			if From != "" && To != "" {
				tfrom, err := strconv.Atoi(From)
				if err != nil {
					return http.StatusBadRequest, err
				}
				tto, err := strconv.Atoi(To)
				if err != nil {
					return http.StatusBadRequest, err
				}
				if !(tfrom >= 0 && tto > 0 && tfrom <= tto) {
					from = 0
					to = 50
				} else {
					from = tfrom
					to = tto
				}
			}
			Applications := []role.Application{}
			if email == "" {
				applications , err := role.GetAllApplications(mctx, Status, from, to)
				if err != nil {
					return http.StatusBadRequest, err
				}
				Applications = applications
			}else {
				applications , err := role.GetApplications(mctx, email, Status, from, to)
				if err != nil {
					return http.StatusBadRequest, err
				}
				Applications = applications
			}
			resp := []role.Application{}
			for _,a := range Applications {
				operatorList := []string{}
				if a.Status == "emails sent" {
					pendApplications, err := role.GetPendingApplicationByApplicationId(mctx, a.Id)
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
			return http.StatusOK, resp
		}
		return http.StatusBadRequest, "only member of lain group can assess"
	})
}


type ApplicationDelete struct {
	server.BaseResource
}


func (ad ApplicationDelete) Delete(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		id := params(ctx, "application_id")
		Id, err := strconv.Atoi(id)
		if err != nil {
			return http.StatusBadRequest, "application_id required"
		}
		application, err := role.GetApplication(mctx, Id)
		if err != nil {
			return http.StatusBadRequest, err
		}
		if application.ApplicantEmail != u.GetProfile().GetEmail() {
			return http.StatusBadRequest, "only applicant can delete application"
		}
		err = role.RecallApplication(mctx, Id)
		if err != nil {
			return http.StatusBadRequest, err
		}
		return http.StatusNoContent, "application deleted"
	})
}


type ApplicationApprove struct {
	server.BaseResource
}


func (aa ApplicationApprove) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		email := u.GetProfile().GetEmail()
		application_status_list, err := role.GetPendingApplicationByEmail(mctx, email)
		if err != nil {
			return http.StatusBadRequest, err
		}
		return http.StatusOK, application_status_list
	})
}


type ApplicationHandle struct {
	server.BaseResource
}


func (ah ApplicationHandle) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		aid := params(ctx, "application_id")
		log.Debug(aid)
		r.ParseForm()
		action := r.Form.Get("action")
		log.Debug(action)
		aId, err := strconv.Atoi(aid)
		if err != nil {
			return http.StatusBadRequest, err
		}
		log.Debug(aId)
		application, err := role.GetApplication(mctx, aId)
		log.Debug(application)
		if err != nil {
			log.Debug(err)
			return http.StatusBadRequest, err
		}
		Ttype := application.TargetType
		if Ttype == "group" {
			name := application.Target.GroupName
			group, err := group.GetGroupByName(mctx, name)
			if err != nil {
				return http.StatusBadRequest, err
			}
			R := 10
			err = mctx.DB.Get(&R, "SELECT role FROM user_group WHERE user_id =? AND group_id=?",
				u.GetId(), group.Id)
			if err != nil {
				return http.StatusBadRequest, err
			}
			if R != 1 {
				return http.StatusBadRequest, "not qualified for the operation"
			}
			back := mctx.Back
			user, err := back.GetUserByFeature(application.ApplicantEmail)
			log.Debug(user)
			if action == "approve" {
				if application.Target.MemberType == "admin" {
					log.Debug("adding member")
					err := group.AddMember(mctx, user, 1)
					if err != nil {
						return http.StatusBadRequest, err
					}
				} else {
					log.Debug("adding member")
					err := group.AddMember(mctx, user, 0)
					if err != nil {
						return http.StatusBadRequest, err
					}
				}
				log.Debug("finishing handing application")
				resp, err1 := role.FinishApplication(mctx, application.Id, "approved", u.GetProfile().GetEmail())
				if err1 != nil {
					return http.StatusBadRequest, err1
				}
				return http.StatusOK, resp
			} else if action == "reject" {
				log.Debug("finishing handing application")
				resp, err1 := role.FinishApplication(mctx, application.Id, "rejected", u.GetProfile().GetEmail())
				if err1 != nil {
					return http.StatusBadRequest, err1
				}
				return http.StatusOK, resp
			} else {
				return http.StatusBadRequest, "no such operation"
			}
		}
		return http.StatusBadRequest, "no such type"
	})
}

