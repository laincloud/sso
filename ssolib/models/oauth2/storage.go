package oauth2

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/RangelReale/osin"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/app"
)

var (
	getApp                   = app.GetApp
	ErrAuthorizeDataNotFound = errors.New("AuthorizeData not found")
	ErrAccessDataNotFound    = errors.New("AccessData not found")
)

const createAuthorizeDataTableSQL = `
CREATE TABLE IF NOT EXISTS oauth2_authorize_data (
	id INT NOT NULL AUTO_INCREMENT,
	app_id INT NOT NULL,
	code VARCHAR(22) NOT NULL,
	expires_in INT NOT NULL DEFAULT 3600,
	scope VARCHAR(128) NOT NULL DEFAULT '',
	redirect_uri VARCHAR(256) NOT NULL DEFAULT '',
	state VARCHAR(256) NOT NULL DEFAULT '',
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	user_id INT NOT NULL,
	PRIMARY KEY (id),
	UNIQUE KEY (code)
) DEFAULT CHARSET=latin1`

type AuthorizeData struct {
	Id          int
	AppId       int `db:"app_id"`
	Code        string
	ExpiresIn   int32 `db:"expires_in"`
	Scope       string
	RedirectUri string `db:"redirect_uri"`
	State       string
	Created     string
	UserId      int `db:"user_id"`
}

func (ad AuthorizeData) CreatedTime() time.Time {
	t, err := parseMySQLTimeString(ad.Created)
	if err != nil {
		panic(err)
	}
	return t
}

type AuthorizeUserData struct {
	AuthorizeDataId int
	UserId          int
}

const createAccessDataTableSQL = `
CREATE TABLE IF NOT EXISTS oauth2_access_data (
	id INT NOT NULL AUTO_INCREMENT,
	app_id INT NOT NULL,
	authorize_data_id INT NOT NULL,
	access_data_id INT NOT NULL,
	access_token CHAR(22) NOT NULL,
	refresh_token CHAR(22) NOT NULL,
	expires_in INT NOT NULL,
	scope VARCHAR(128) NOT NULL,
	redirect_uri VARCHAR(256) NOT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	user_id INT NOT NULL,
	PRIMARY KEY (id),
	UNIQUE KEY (access_token)
) DEFAULT CHARSET=latin1`

type AccessData struct {
	Id              int
	AppId           int    `db:"app_id"`
	AuthorizeDataId int    `db:"authorize_data_id"`
	AccessDataId    int    `db:"access_data_id"`
	AccessToken     string `db:"access_token"`
	RefreshToken    string `db:"refresh_token"`
	ExpiresIn       int32  `db:"expires_in"`
	Scope           string
	RedirectUri     string `db:"redirect_uri"`
	Created         string
	UserId          int `db:"user_id"`
}

func (ad AccessData) CreatedTime() time.Time {
	t, err := parseMySQLTimeString(ad.Created)
	if err != nil {
		panic(err)
	}
	return t
}

type AccessUserData struct {
	AccessDataId int
	UserId       int
}

const createRefreshTableSQL = `
CREATE TABLE IF NOT EXISTS oauth2_refresh (
	id INT NOT NULL AUTO_INCREMENT,
	refresh_token CHAR(22) NOT NULL,
	access_token CHAR(22) NOT NULL,
	PRIMARY KEY (id),
	UNIQUE KEY (refresh_token)
) DEFAULT CHARSET=latin1`

type Refresh struct {
	Id           int
	RefreshToken string `db:"refresh_token"`
	AccessToken  string `db:"access_token"`
}

func InitDatabase(ctx *models.Context) {
	log.Debug("oauth2.InitDatabase")
	tx := ctx.DB.MustBegin()
	tx.MustExec(createAuthorizeDataTableSQL)
	tx.MustExec(createAccessDataTableSQL)
	tx.MustExec(createRefreshTableSQL)
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

type Storage struct {
	ctx *models.Context
}

func NewOAuth2Storage(ctx *models.Context) *Storage {
	s := &Storage{ctx: ctx}
	return s
}

func (s *Storage) Clone() osin.Storage {
	return s
}

func (s *Storage) Close() {
}

func (s *Storage) GetClient(id string) (osin.Client, error) {
	log.Debugf("GetClient %s", id)
	appId, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	app, err := getApp(s.ctx, appId)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (s *Storage) SetClient(id string, client osin.Client) error {
	return nil
}

func (s *Storage) SaveAuthorize(data *osin.AuthorizeData) error {
	log.Debugf("SaveAuthorize: %s", data.Code)

	tx := s.ctx.DB.MustBegin()
	_, err1 := tx.Exec(
		"INSERT INTO oauth2_authorize_data (app_id, code, expires_in, scope, redirect_uri, state, user_id) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		data.Client.GetId(), data.Code, data.ExpiresIn, data.Scope,
		data.RedirectUri, data.State, data.UserData.(AuthorizeUserData).UserId)

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}

	if err1 != nil {
		return err1
	}

	return nil
}

func (s *Storage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	log.Debugf("LoadAuthorize: %s", code)
	data := AuthorizeData{}
	if err := s.ctx.DB.Get(&data, "SELECT * FROM oauth2_authorize_data WHERE code=?", code); err != nil {
		return nil, err
	}

	a, err := getApp(s.ctx, data.AppId)
	if err != nil {
		return nil, err
	}

	return &osin.AuthorizeData{
		Client:      a,
		Code:        data.Code,
		ExpiresIn:   data.ExpiresIn,
		Scope:       data.Scope,
		RedirectUri: data.RedirectUri,
		State:       data.State,
		CreatedAt:   data.CreatedTime(),
		UserData:    AuthorizeUserData{AuthorizeDataId: data.Id, UserId: data.UserId},
	}, nil
}

func (s *Storage) getAuthorizeData(id int) (*osin.AuthorizeData, error) {
	log.Debugf("getAuthorizeData: %d", id)
	data := AuthorizeData{}
	if err := s.ctx.DB.Get(&data, "SELECT * FROM oauth2_authorize_data WHERE id=?", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAuthorizeDataNotFound
		}
		return nil, err
	}
	log.Debugf("getAuthorizeData: %s finish", id)

	a, err := getApp(s.ctx, data.AppId)
	if err != nil {
		return nil, err
	}

	return &osin.AuthorizeData{
		Client:      a,
		Code:        data.Code,
		ExpiresIn:   data.ExpiresIn,
		Scope:       data.Scope,
		RedirectUri: data.RedirectUri,
		State:       data.State,
		CreatedAt:   data.CreatedTime(),
		UserData:    AuthorizeUserData{AuthorizeDataId: data.Id, UserId: data.UserId},
	}, nil
}

func (s *Storage) RemoveAuthorize(code string) error {
	log.Debugf("RemoveAuthorize: %s", code)
	tx := s.ctx.DB.MustBegin()
	_, err1 := tx.Exec("DELETE FROM oauth2_authorize_data WHERE code=?", code)

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}

	if err1 != nil {
		return err1
	}

	return nil
}

func (s *Storage) SaveAccess(data *osin.AccessData) error {
	log.Debugf("SaveAccess: %s %v", data.AccessToken, data.AccessData)
	userdata, ok := data.UserData.(AccessUserData)
	if !ok {
		userdata = AccessUserData{}
	}
	tx := s.ctx.DB.MustBegin()

	defer func() {
		tx.Commit()
	}()

	var someID int
	if data.AuthorizeData != nil {
		// for code flow
		someID = data.AuthorizeData.UserData.(AuthorizeUserData).AuthorizeDataId
	} else {
		// for implicit flow
		someID = -1
	}

	log.Debug(userdata)
	if _, err := tx.Exec(
		"INSERT INTO oauth2_access_data (app_id, authorize_data_id, "+
			"access_data_id, access_token, refresh_token, "+
			"expires_in, scope, redirect_uri, user_id) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		data.Client.GetId(),
		someID,
		userdata.AccessDataId,
		data.AccessToken, data.RefreshToken, data.ExpiresIn, data.Scope,
		data.RedirectUri, userdata.UserId); err != nil {
		return err
	}

	if data.RefreshToken != "" {
		if _, err := tx.Exec(
			"INSERT INTO oauth2_refresh (refresh_token, access_token) "+
				"VALUES (?, ?)", data.RefreshToken, data.AccessToken); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) LoadAccess(code string) (*osin.AccessData, error) {
	log.Debugf("LoadAccess: %s", code)
	ad := AccessData{}
	if err := s.ctx.DB.Get(&ad, "SELECT * FROM oauth2_access_data WHERE access_token=?", code); err != nil {
		log.Errorf("LoadAccess: find access_token failed: %s", err)
		return nil, err
	}
	log.Debugf("LoadAccess: %s, finish select oauth2_access_data", code)

	a, err := getApp(s.ctx, ad.AppId)
	if err != nil {
		log.Errorf("LoadAccess: getApp failed: %s", err)
		return nil, err
	}

	authd, err := s.getAuthorizeData(ad.AuthorizeDataId)
	if err != nil && err != ErrAuthorizeDataNotFound {
		log.Errorf("LoadAccess: getAuthorizeData failed: %s", err)
		return nil, err
	}
	log.Debugf("LoadAccess: getAuthorizeData: %d finish", ad.AuthorizeDataId)

	accd, err := s.getAccessData(ad.AccessDataId)
	if err != nil && err != ErrAccessDataNotFound {
		log.Errorf("LoadAccess: getAccessData failed: %s", err)
		return nil, err
	}

	data := &osin.AccessData{
		Client:        a,
		AuthorizeData: authd,
		AccessData:    accd,
		AccessToken:   ad.AccessToken,
		RefreshToken:  ad.RefreshToken,
		ExpiresIn:     ad.ExpiresIn,
		Scope:         ad.Scope,
		RedirectUri:   ad.RedirectUri,
		CreatedAt:     ad.CreatedTime(),
		UserData:      AccessUserData{AccessDataId: ad.Id, UserId: ad.UserId},
	}
	log.Debugf("AccessData loaded: %v", data)
	return data, nil
}

func (s *Storage) getAccessData(id int) (*osin.AccessData, error) {
	log.Debugf("getAccessData: %d", id)
	ad := AccessData{}
	if err := s.ctx.DB.Get(&ad, "SELECT * FROM oauth2_access_data WHERE id=?", id); err != nil {
		if err == sql.ErrNoRows {
			log.Debug("cannot find access data: ", id)
			return nil, ErrAccessDataNotFound
		}
		return nil, err
	}
	a, err := getApp(s.ctx, ad.AppId)
	if err != nil {
		return nil, err
	}

	data := &osin.AccessData{
		Client:        a,
		AuthorizeData: nil,
		AccessData:    nil,
		AccessToken:   ad.AccessToken,
		RefreshToken:  ad.RefreshToken,
		ExpiresIn:     ad.ExpiresIn,
		Scope:         ad.Scope,
		RedirectUri:   ad.RedirectUri,
		CreatedAt:     ad.CreatedTime(),
		UserData:      AccessUserData{AccessDataId: ad.Id, UserId: ad.UserId},
	}
	return data, nil
}

func (s *Storage) RemoveAccess(code string) error {
	log.Debugf("RemoveAccess: %s", code)
	tx := s.ctx.DB.MustBegin()
	_, err1 := tx.Exec("DELETE FROM oauth2_access_data WHERE access_token=?", code)

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}

	if err1 != nil {
		return err1
	}

	return nil
}

func (s *Storage) LoadRefresh(code string) (*osin.AccessData, error) {
	log.Debugf("LoadRefresh: %s", code)
	r := Refresh{}
	if err := s.ctx.DB.Get(&r, "SELECT * FROM oauth2_refresh WHERE refresh_token=?", code); err != nil {
		return nil, err
	}

	return s.LoadAccess(r.AccessToken)
}

func (s *Storage) RemoveRefresh(code string) error {
	log.Debugf("RemoveRefresh: %s", code)
	tx := s.ctx.DB.MustBegin()
	_, err1 := tx.Exec("DELETE FROM oauth2_refresh WHERE refresh_token=?", code)

	if err2 := tx.Commit(); err2 != nil {
		return err2
	}
	if err1 != nil {
		return err1
	}
	return nil
}

func parseMySQLTimeString(s string) (time.Time, error) {
	ss, err := time.Parse("2006-01-02 15:04:05", s)
	log.Debugf("Parse %s to %v", s, ss)
	return ss, err
}
