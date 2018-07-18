package role

import (
	"github.com/mijia/sweb/log"
	"github.com/laincloud/sso/ssolib/models"
	"database/sql"
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
	commitor_email VARCHAR(64) NULL DEFAULT NULL,
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


type Application struct {
	Id               int    `json:"id"`
	Applicant_email   string `json:"applicant_email"`
	Target_type       string `json:"target_type"`
	Target           TargetContent `json:"target"`
	Reason           string `db:"reason" json:"reason"`
	Status           string `json:"status"`
	Commitor_email   string `json:"commitor_email"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
}

type tempApplication struct {
	Id               int    `json:"id"`
	Applicant_email   string `db:"applicant_email" json:"applicant_email"`
	Target_type       string `db: "target_type" json:"target_type"`
	Target           string `json:"target"`
	Reason           string `db:"reason" json:"reason"`
	Status           string `json:"status"`
	Commitor_email   string `json:"commitor_email"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
}

type TargetContent struct {
	GroupName string `json:"name"`
	MemberType string `json:"role"`
}

type Pending_Application struct {
	Id               int    `json:"id"`
	Application_Id    int `json:"application_id"`
	Operator_email   string `json:"operator_email"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
}


func CreatePending_Application (ctx *models.Context, application_id int, opr_email string) (*Pending_Application, error) {
	tx := ctx.DB.MustBegin()
	result, err := tx.Exec(
		"INSERT INTO pending_application (application_id, operator_email) VALUES(?, ?)",application_id, opr_email)
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
	return GetPending_Application(ctx, int(id))
}


func GetPending_ApplicationByEmail (ctx *models.Context, email string) ([]Pending_Application, error) {
	applicationstatus := []Pending_Application{}
	err := ctx.DB.Select(&applicationstatus, "SELECT * FROM pending_application WHERE operator_email=?", email)
	if err == sql.ErrNoRows {
		return nil, ErrResourceNotFound
	} else if err != nil {
		return nil, err
	}
	return applicationstatus, nil
}


func GetPending_Application(ctx *models.Context, id int) (*Pending_Application, error) {
	applications := Pending_Application{}
	err := ctx.DB.Get(&applications, "SELECT * FROM pending_application WHERE id=?", id)
	if err == sql.ErrNoRows {
		return nil, ErrResourceNotFound
	} else if err != nil {
		return nil, err
	}
	return &applications, nil
}

func GetPending_ApplicationByApplicationId(ctx *models.Context, id int) ([]Pending_Application, error) {
	log.Debug("start getting pending_application")
	applications := []Pending_Application{}
	err := ctx.DB.Select(&applications, "SELECT * FROM pending_application WHERE application_id=?", id)
	if err == sql.ErrNoRows {
		log.Debug(err)
		return nil, ErrResourceNotFound
	} else if err != nil {
		log.Debug(err)
		return nil, err
	}
	return applications, nil
}

func CreateApplication (ctx *models.Context, applicant_email string, target_type string, target *TargetContent, Reason string) (*Application, error) {
	log.Debug("CreateApplication")
	tx := ctx.DB.MustBegin()
	t, err:= json.Marshal(target)
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(
		"INSERT INTO application (applicant_email, target_type, target, reason, status, commitor_email) VALUES(?, ?, ?, ?, ?, ?)",
		applicant_email, target_type, string(t), Reason, "None email sent", "NULL")
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


func UpdateApplication (ctx *models.Context, id int, Status string, Commitor_email string) (*Application, error) {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec(
		"UPDATE application SET status=?, commitor_email=? WHERE id=?",Status, Commitor_email, id)
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

func FinishApplication (ctx *models.Context, id int, Status string, Commitor_email string) (*Application, error) {
	tx := ctx.DB.MustBegin()
	_, err := tx.Exec(
		"UPDATE application SET status=?, commitor_email=? WHERE id=?",Status, Commitor_email, id)
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
	applications := []tempApplication{}
	if status != "" {
		err := ctx.DB.Select(&applications, "SELECT * FROM application WHERE applicant_email=? AND status=? ORDER BY created DESC LIMIT ?, ?", email, status, from, to)
		if err != nil {
			return nil, err
		}
	}else {
		err := ctx.DB.Select(&applications, "SELECT * FROM application WHERE applicant_email=? ORDER BY created DESC LIMIT ?, ?", email, from, to)
		if err != nil {
			return nil, err
		}
	}
	Applications := []Application{}
	for _,a := range applications {
		t := TargetContent{}
		json.Unmarshal([]byte(a.Target), &t)
		temp := Application{
			Id: a.Id,
			Applicant_email: a.Applicant_email,
			Target_type: a.Target_type,
			Target: t,
			Reason: a.Reason,
			Status: a.Status,
			Commitor_email: a.Commitor_email,
			Created: a.Created,
			Updated: a.Updated,
		}
		Applications = append(Applications, temp)
	}

	return Applications, nil
}

func GetAllApplications(ctx *models.Context, status string, from int, to int) ([]Application, error) {
	applications := []tempApplication{}
	if status != "" {
		err := ctx.DB.Select(&applications, "SELECT * FROM application WHERE status=? ORDER BY created DESC LIMIT ?, ?",status, from, to)
		if err != nil {
			return nil, err
		}
	}else {
		err := ctx.DB.Select(&applications, "SELECT * FROM application ORDER BY created DESC LIMIT ?, ?", from, to)
		if err != nil {
			return nil, err
		}
	}
	Applications := []Application{}
	for _,a := range applications {
		t := TargetContent{}
		json.Unmarshal([]byte(a.Target), &t)
		temp := Application{
			Id: a.Id,
			Applicant_email: a.Applicant_email,
			Target_type: a.Target_type,
			Target: t,
			Reason: a.Reason,
			Status: a.Status,
			Commitor_email: a.Commitor_email,
			Created: a.Created,
			Updated: a.Updated,
		}
		Applications = append(Applications, temp)
	}

	return Applications, nil
}


func GetApplication(ctx *models.Context, id int) (*Application, error) {
	application := tempApplication{}
	err := ctx.DB.Get(&application, "SELECT * FROM application WHERE id=?", id)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	t := TargetContent{}
	json.Unmarshal([]byte(application.Target), &t)
	temp := Application{
		Id: application.Id,
		Applicant_email: application.Applicant_email,
		Target_type: application.Target_type,
		Target: t,
		Reason: application.Reason,
		Status: application.Status,
		Commitor_email: application.Commitor_email,
		Created: application.Created,
		Updated: application.Updated,
	}
	return &temp, nil
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
