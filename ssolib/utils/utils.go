package utils

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"
)

func InitMysql(mysqlDSN string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", mysqlDSN)
	if err != nil {
		return nil, err
	}
	db.MustExec(`SET time_zone = '+00:00';`)
	return db, nil
}

func AlterTable(db *sqlx.DB) {
	alterGroup(db)
}

func alterGroup(db *sqlx.DB) {
	groupName := "admins"
	if rows, err := db.Query("SELECT backend FROM `group` WHERE name=?", groupName); err != nil {
		log.Debug(rows)
		log.Debug(err)
		if strings.Index(err.Error(), `Unknown column 'backend' in 'field list'`) != -1 {
			alterSql := "AlTER TABLE `group` ADD COLUMN backend TINYINT NOT NULL DEFAULT 0"
			log.Info("begin:", alterSql)
			db.MustExec(alterSql)
			log.Info("end: ", alterSql)
		}
	} else {
		log.Debug(rows)
	}
}
