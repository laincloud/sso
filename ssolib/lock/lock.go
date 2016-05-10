package lock

import (
	"database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"
	"github.com/laincloud/sso/ssolib/utils"
)

const LockTableName = "distributedlock"

const createLockTableSQL = `
CREATE TABLE IF NOT EXISTS ` + LockTableName + ` (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(32) NOT NULL,
	status INT NOT NULL DEFAULT 0,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY (name)
) DEFAULT CHARSET=latin1`

type DistributedLock struct {
	Id      int
	Name    string
	Status  int
	Created string
	Updated string
	db      *sqlx.DB
	tx      *sqlx.Tx
	lock    sync.Mutex
}

func New(mysqlDSN string, lockName string) *DistributedLock {
	db, err := utils.InitMysql(mysqlDSN)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	lock := DistributedLock{db: db}
	lock.initTable(lockName)
	return &lock
}

func (l *DistributedLock) initTable(lockName string) {
	l.db.MustExec(createLockTableSQL)
	tx := l.db.MustBegin()
	err := tx.Get(l, ("SELECT * FROM " + LockTableName + " WHERE name=? FOR UPDATE"), lockName)
	if err == sql.ErrNoRows {
		_, err := tx.Exec(("INSERT INTO " + LockTableName + " (name) VALUES (?)"), lockName)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		log.Info("created distributedlock:", lockName)
	} else if err != nil {
		panic(err)
	} else {
		err = tx.Commit()
		if err != nil {
			panic(err)
		}
		return
	}
	err = tx.Get(l, ("SELECT * FROM " + LockTableName + " WHERE name=? FOR UPDATE"), lockName)
	if err != nil {
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	log.Debug(l.Name)
}

func (l *DistributedLock) Lock() {
	l.lock.Lock()
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			l.tx.Commit()
			l.lock.Unlock()
		}
	}()
	l.tx = l.db.MustBegin()
	err := l.tx.Get(l, ("SELECT * FROM " + LockTableName + " WHERE name=? FOR UPDATE"), l.Name)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	if l.Status != 0 {
		panic("unexpected status of lock")
	}
	_, err = l.tx.Exec(("UPDATE " + LockTableName + " SET status=? WHERE name=?"), 1, l.Name)
	if err != nil {
		panic(err)
	}
}

func (l *DistributedLock) Unlock() {
	defer l.lock.Unlock()
	_, err := l.tx.Exec(("UPDATE " + LockTableName + " SET status=? WHERE name=?"), 0, l.Name)
	if err != nil {
		panic(err)
	}
	err = l.tx.Commit()
	if err != nil {
		panic(err)
	}
}
