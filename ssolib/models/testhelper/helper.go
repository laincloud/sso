package testhelper

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/lock"
	"github.com/laincloud/sso/ssolib/models"
	"github.com/laincloud/sso/ssolib/models/testbackend"
)

func clearDatabase(ctx *models.Context) {
	log.Debug("testhelper.clearDatabase")
	tables := []string{}
	tx := ctx.DB.MustBegin()
	tx.Select(&tables, "SHOW TABLES")
	for _, table := range tables {
		tx.MustExec(fmt.Sprintf("TRUNCATE TABLE `%s`", table))
	}
	tx.Commit()
}

type TestHelper struct {
	T   *testing.T
	Ctx *models.Context
}

func GetTestMysqlDSN() string {
	mysqlDSN := os.Getenv("TEST_MYSQL_DSN")
	if mysqlDSN == "" {
		mysqlDSN = "test:test@tcp(127.0.0.1:3306)/sso_test"
	} else {
		if !strings.HasSuffix(mysqlDSN, "_test") {
			log.Fatal("Database must end with _test")
		}
	}
	return mysqlDSN
}

func NewTestHelper(t *testing.T) TestHelper {
	log.EnableDebug()
	mysqlDSN := GetTestMysqlDSN()
	db, err := sqlx.Connect("mysql", mysqlDSN)
	if err != nil {
		log.Fatal(err)
	}

	siteURL, err := url.Parse("http://example.com")
	if err != nil {
		log.Fatal(err)
	}

	ctx := &models.Context{
		DB:         db,
		SSOSiteURL: siteURL,
		SMTPAddr:   "smtp.example.com:25",
		EmailFrom:  "sso@example.com",
		Back:       &testbackend.TestBackend{},
	}

	clearDatabase(ctx)

	ctx.Lock = lock.New(mysqlDSN, "testlock")

	return TestHelper{
		T:   t,
		Ctx: ctx,
	}
}
