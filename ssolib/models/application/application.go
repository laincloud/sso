package application

import (
	"github.com/mijia/sweb/log"
	"github.com/laincloud/sso/ssolib/models"
	"encoding/json"
)

const createApplicationTableSQL = `
CREATE TABLE IF NOT EXISTS application (
	id INT NOT NULL AUTO_INCREMENT,
	applicant_email VARCHAR(64) NULL DEFAULT NULL,
	target_type VARCHAR(64) NULL DEFAULT NULL,
	target VARCHAR(128) CHARACTER SET utf8 NOT NULL,
	reason VARCHAR(1024) CHARACTER SET utf8 NOT NULL,
    status VARCHAR(64) NULL DEFAULT NULL,
	commit_email VARCHAR(64) NULL DEFAULT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY (applicant_email, target_type, target)
) DEFAULT CHARSET=latin1`


const createPENDING_ApplicationStatusTableSQL = `
CREATE TABLE IF NOT EXISTS pending_application (
	id INT NOT NULL AUTO_INCREMENT,
	application_id INT NOT NULL,
	operator_email VARCHAR(64) NULL DEFAULT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY (application_id,operator_email )
) DEFAULT CHARSET=latin1`


func InitDatabase(ctx *models.Context) {
	ctx.DB.MustExec(createApplicationTableSQL)
	ctx.DB.MustExec(createPENDING_ApplicationStatusTableSQL)
}

type Application struct {
	Id             int            `json:"id"`
	ApplicantEmail string         `db:"applicant_email" json:"applicant_email"`
	TargetType     string         `db:"target_type" json:"target_type"`
	TargetStr      string         `db:"target"`
	TargetContent  *TargetContent `json:"target"`
	Reason         string         `json:"reason"`
	Status         string         `json:"status"`
	CommitEmail    string         `db:"commit_email" json:"commit_email"`
	Created        string         `json:"created"`
	Updated        string         `json:"updated"`
}

func (a *Application) MarshalJson() ([]byte, error) {
	ret := map[string]interface{}{
		"id": a.Id,
	}
	return json.Marshal(ret)
}

type TargetContent struct {
	Name string `json:"name"`
	Role string `json:"role"`
}


func (a *Application) ParseTarget() {
	if a.TargetContent == nil && a.TargetStr !="" {
		json.Unmarshal([]byte(a.TargetStr), &(a.TargetContent))
	}
	a.TargetStr = ""
}

func (a *Application) ParseOprEmail(emails []string) {
	if a.Status == "initialled" {
		t, err:= json.Marshal(emails)
		if err != nil {
			log.Debug(err)
		}
		a.CommitEmail = string(t)
	}
}


type PendingApplication struct {
	Id            int    `json:"id"`
	ApplicationId int    `db:"application_id" json:"application_id"`
	OperatorEmail string `db:"operator_email" json:"operator_email"`
	Created       string `json:"created"`
	Updated       string `json:"updated"`
}


func CreatePendingApplication(ctx *models.Context, applicationId int, oprEmail string) (*PendingApplication, error) {
	tx := ctx.DB.MustBegin()
	result, err := tx.Exec(
		"INSERT INTO pending_application (application_id, operator_email) VALUES(?, ?)",applicationId, oprEmail)
	if err2 := tx.Commit(); err2 != nil {
		log.Debug(err2)
		return nil, err2
	}
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	return GetPendingApplication(ctx, int(id))
}


func GetPendingApplicationByEmail(ctx *models.Context, email string) ([]PendingApplication, error) {
	applicationstatus := []PendingApplication{}
	err := ctx.DB.Select(&applicationstatus, "SELECT * FROM pending_application WHERE operator_email=?", email)
	if err != nil {
		return nil, err
	}
	return applicationstatus, nil
}


func GetPendingApplication(ctx *models.Context, id int) (*PendingApplication, error) {
	applications := PendingApplication{}
	err := ctx.DB.Get(&applications, "SELECT * FROM pending_application WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	return &applications, nil
}

func GetPendingApplicationByApplicationId(ctx *models.Context, id int) ([]PendingApplication, error) {
	log.Debug("start getting pending_application")
	applications := []PendingApplication{}
	err := ctx.DB.Select(&applications, "SELECT * FROM pending_application WHERE application_id=?", id)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	return applications, nil
}

func CreateApplication (ctx *models.Context, applicantEmail string, targetType string, target *TargetContent, Reason string) (*Application, error) {
	log.Debug("CreateApplication")
	tx := ctx.DB.MustBegin()
	t, err:= json.Marshal(target)
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(
		"INSERT INTO application (applicant_email, target_type, target, reason, status, commit_email) VALUES(?, ?, ?, ?, ?, ?)",
		applicantEmail, targetType, string(t), Reason, "initialled", "NULL")
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	if err2 := tx.Commit(); err2 != nil {
		log.Debug(err2)
		return nil, err2
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	log.Debug("finish creating application")
	return GetApplication(ctx, int(id))
}


func UpdateApplication (ctx *models.Context, id int, Status string, CommitEmail string) (*Application, error) {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec(
		"UPDATE application SET status=?, commit_email=? WHERE id=?",Status, CommitEmail, id)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	if err1 := tx.Commit(); err1 != nil {
		log.Debug(err1)
		return nil, err1
	}
	return GetApplication(ctx, id)
}

func FinishApplication (ctx *models.Context, id int, Status string, CommitEmail string) (*Application, error) {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec(
		"UPDATE application SET status=?, commit_email=? WHERE id=?",Status, CommitEmail, id)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	_, err1 := tx.Exec("DELETE FROM pending_application WHERE application_id=?", id)
	if err1 != nil {
		log.Debug(err1)
		return nil, err1
	}
	if err2 := tx.Commit(); err2 != nil {
		log.Debug(err2)
		return nil, err2
	}
	return GetApplication(ctx, id)
}



func GetApplications(ctx *models.Context, email string, status string, from int, to int) ([]Application, error) {
	applications := []Application{}
	if status != "" {
		err := ctx.DB.Select(&applications, "SELECT * FROM application WHERE applicant_email=? AND status=? ORDER BY created DESC LIMIT ?, ?", email, status, from, to - from + 1)
		if err != nil {
			return nil, err
		}
	}else {
		err := ctx.DB.Select(&applications, "SELECT * FROM application WHERE applicant_email=? ORDER BY created DESC LIMIT ?, ?", email, from, to - from + 1)
		if err != nil {
			return nil, err
		}
	}
	Applications := []Application{}
	for _,a := range applications {
		a.ParseTarget()
		Applications = append(Applications, a)
	}

	return Applications, nil
}

func GetAllApplications(ctx *models.Context, status string, from int, to int) ([]Application, error) {
	applications := []Application{}
	if status != "" {
		err := ctx.DB.Select(&applications, "SELECT * FROM application WHERE status=? ORDER BY created DESC LIMIT ?, ?",status, from, to - from + 1)
		if err != nil {
			return nil, err
		}
	}else {
		err := ctx.DB.Select(&applications, "SELECT * FROM application ORDER BY created DESC LIMIT ?, ?", from, to - from + 1)
		if err != nil {
			return nil, err
		}
	}
	Applications := []Application{}
	for _,a := range applications {
		a.ParseTarget()
		Applications = append(Applications, a)
	}

	return Applications, nil
}


func GetApplication(ctx *models.Context, id int) (*Application, error) {
	application := Application{}
	err := ctx.DB.Get(&application, "SELECT * FROM application WHERE id=?", id)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	application.ParseTarget()
	return &application, nil
}

func RecallApplication(ctx *models.Context, id int) (error) {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec("DELETE FROM application WHERE id=?", id)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err1 := tx.Exec("DELETE FROM pending_application WHERE application_id=?", id)
	if err1 != nil {
		return err1
	}
	if err2 := tx.Commit(); err2 != nil {
		return err2
	}
	return nil
}
