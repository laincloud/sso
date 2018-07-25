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
)



type Apply struct {
	server.BaseResource
}

func applyEnterGroup(mctx *models.Context, id int, target application.TargetContent, user iuser.User, reason string)(*application.Application, error) {
	res, err := application.CreateApplication(mctx, user.GetProfile().GetEmail(), "group", &target, reason)
	if err != nil {
		return nil, err
	}
	rows, err := mctx.DB.Query("SELECT user_id FROM user_group WHERE group_id=? AND role=?", id, 1)
	if err != nil {
		return nil, err
	}
	log.Debug(rows)
	back := mctx.Back
	for rows.Next() {
		var userId int
		if err = rows.Scan(&userId); err == nil {
			log.Debug(userId)
			user, err := back.GetUser(userId)
			if err == nil {
				adminEmail := user.GetProfile().GetEmail()
				log.Debug(adminEmail)
				g, err := group.GetGroup(mctx, id)
				if err != nil {
					log.Debug(err)
				}
				content := user.GetName() + "applies join" + g.Name + "\n" + "please log in to handle it"
				err1 := SendTo("new application", content, adminEmail)
				if err1 != nil{
					log.Debug(err1)
				}
				application.CreatePendingApplication(mctx, res.Id, adminEmail)
			}
		}
	}
	return res, nil
}



type Application struct {
	Description string               `json:"description"`
	Target      []application.TargetContent `json:"target"`
	TargetType  string               `json:"target_type"`
	Reason      string               `json:"reason"`
}



func GetTypeOfUser(ctx *models.Context, id int, groupId int) (int, error){
	role := []int{}
	err := ctx.DB.Select(&role, "SELECT role FROM user_group WHERE user_id=? AND group_id=?", id, groupId)
	log.Debug(err)
	if err != nil {
		return -1, err
	}
	return role[0], nil
}


func CheckIfInGroup(group_ids []int, id int) (bool) {
	left := 0;
	right := len(group_ids) - 1
	if len(group_ids) == 0 {
		return false
	}
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

func CheckIfQualified(ctx *models.Context, group_ids []int, id int, target application.TargetContent, u iuser.User) (bool, error) {
	if CheckIfInGroup(group_ids, id) {
		role, err := GetTypeOfUser(ctx, u.GetId(), id)
		if err != nil {
			return false, err
		}
		if role == 0 && target.Role == "admin" {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}


func (ay Apply) Get(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		r.ParseForm()
		mctx := getModelContext(ctx)
		applicantEmail := r.Form.Get("applicant_email")
		Status := r.Form.Get("status")
		From := r.Form.Get("from")
		To := r.Form.Get("to")
		from := 0
		to := 50
		Applications := []application.Application{}
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
		currentEmail := u.GetProfile().GetEmail()
		commitEmail := r.Form.Get("commit_email")
		if commitEmail == currentEmail {
			pendingApplications, err := application.GetPendingApplicationByEmail(mctx, currentEmail)
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
			return http.StatusOK, Applications
		} else if commitEmail == "" {
			islain := false
			laingroup, err := group.GetGroupByName(mctx, "lain")
			if laingroup != nil && err == nil {
				id := laingroup.Id
				usergroup, err := getDirectGroupsOfUser(mctx, u)
				if err == nil && usergroup != nil {
					islain = CheckIfInGroup(usergroup, id)
				}
			}
			if islain && applicantEmail == "" {
				applications, err := application.GetAllApplications(mctx, Status, from, to)
				if err != nil {
					return http.StatusBadRequest, err
				}
				Applications = applications
			} else if applicantEmail == "" || applicantEmail == currentEmail{
				applications, err := application.GetApplications(mctx, currentEmail, Status, from, to)
				if err != nil {
					return http.StatusBadRequest, err
				}
				Applications = applications
				log.Debug(applications)
			} else if islain && applicantEmail != currentEmail {
				applications, err := application.GetApplications(mctx, applicantEmail, Status, from, to)
				if err != nil {
					return http.StatusBadRequest, err
				}
				Applications = applications
			} else {
				return http.StatusBadRequest, "not qualified for the operation"
			}
			resp := []application.Application{}
			for _,a := range Applications {
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
			return http.StatusOK, resp
		}
		return http.StatusBadRequest, "commitEmail is not vaild"
	})
}



func (ay Apply) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		log.Debug("apply starting")
		req := Application{}
		if err := form.ParamBodyJson(r, &req); err != nil {
			return http.StatusBadRequest, err
		}
		log.Debug(u)
		if req.TargetType == "group" {
			groupIds, err := getDirectGroupsOfUser(mctx, u)
			if err != nil {
				return http.StatusBadRequest, err
			}
			if len(req.Target) == 1 {
				group, err := group.GetGroupByName(mctx, req.Target[0].Name)
				if err != nil {
					return http.StatusBadRequest, err
				}
				id := group.Id
				qualified, err := CheckIfQualified(mctx, groupIds, id, req.Target[0], u)
				if err != nil {
					log.Debug(err)
					return http.StatusInternalServerError , err
				}
				if qualified {
					resp, err := applyEnterGroup(mctx, id, req.Target[0], u, req.Reason)
					if err != nil {
						return http.StatusBadRequest, err
					}
					return http.StatusOK, resp
				}
				return http.StatusBadRequest, "already in the group!"
			} else if len(req.Target) > 1 {
				for i := 0; i < len(req.Target); i++ {
					_, err := group.GetGroupByName(mctx, req.Target[i].Name)
					if err != nil {
						if err == group.ErrGroupNotFound {
							return http.StatusBadRequest, err
						}

					}
				}
				var applications []application.Application
				for i := 0; i < len(req.Target); i++ {
					g, err := group.GetGroupByName(mctx, req.Target[i].Name)
					qualified, err := CheckIfQualified(mctx, groupIds, g.Id, req.Target[i], u)
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
						resp, err := applyEnterGroup(mctx, g.Id, req.Target[i], u, req.Reason)
						if err == nil {
							if resp != nil {
								applications = append(applications, *resp)
							}
						} else {
							return http.StatusBadRequest, applications
						}
					}
				}
				return http.StatusOK, applications
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


type ApplicationHandle struct {
	server.BaseResource
}


func (ah ApplicationHandle) Post(ctx context.Context, r *http.Request) (int, interface{}) {
	return requireLogin(ctx, func(u iuser.User) (int, interface{}) {
		mctx := getModelContext(ctx)
		aid := params(ctx, "application_id")
		r.ParseForm()
		action := r.Form.Get("action")
		log.Debug(action)
		aId, err := strconv.Atoi(aid)
		if err != nil {
			return http.StatusBadRequest, err
		}
		application1, err := application.GetApplication(mctx, aId)
		if err != nil {
			log.Debug(err)
			return http.StatusBadRequest, err
		}
		Ttype := application1.TargetType
		if Ttype == "group" {
			if action == "recall" {
				if application1.ApplicantEmail != u.GetProfile().GetEmail() {
					return http.StatusBadRequest, "only applicant can delete application"
				}
				err = application.RecallApplication(mctx, aId)
				if err != nil {
					return http.StatusBadRequest, err
				}
				return http.StatusNoContent, "application deleted"
			}
			name := application1.TargetContent.Name
			group, err := group.GetGroupByName(mctx, name)
			log.Debug(group)
			if err != nil {
				return http.StatusBadRequest, err
			}
			R := []int{}
			err = mctx.DB.Select(&R, "SELECT role FROM user_group WHERE user_id =? AND group_id=?",
				u.GetId(), group.Id)
			if err != nil {
				log.Debug(err)
				return http.StatusBadRequest, err
			}
			if R[0] != 1 {
				return http.StatusBadRequest, "not qualified for the operation"
			}
			back := mctx.Back
			user, err := back.GetUserByFeature(application1.ApplicantEmail)
			log.Debug(user)
			if action == "approve" {
				if application1.TargetContent.Role == "admin" {
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
				resp, err1 := application.FinishApplication(mctx, application1.Id, "approved", u.GetProfile().GetEmail())
				if err1 != nil {
					return http.StatusBadRequest, err1
				}
				return http.StatusOK, resp
			} else if action == "reject" {
				log.Debug("finishing handing application")
				resp, err1 := application.FinishApplication(mctx, application1.Id, "rejected", u.GetProfile().GetEmail())
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

