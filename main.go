package main

import (
	"flag"
	"os"
	"runtime"
	"strings"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib"
	"github.com/laincloud/sso/ssomysql"
	"github.com/laincloud/sso/ssomysql/user"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var webAddr, mysqlDSN, siteURL, smtpAddr, emailUserPassword, emailFrom, emailPassword, emailSuffix string
	var prikeyfile, pubkeyfile string
	var isDebug bool
	var emailtls bool
	var sentryDSN string
	var queryUser bool
	flag.StringVar(&webAddr, "web", ":14000", "The address which SSO service is listening on")
	flag.StringVar(&mysqlDSN, "mysql", "user:password@tcp(127.0.0.1:3306)/dbname",
		"Data source name of mysql connection")
	flag.StringVar(&siteURL, "site", "http://sso.example.com", "Base URL of SSO site")
	flag.StringVar(&smtpAddr, "smtp", "mail.example.com:25", "SMTP address for sending mail")
	flag.StringVar(&emailUserPassword, "from", "sso@example.com", "Email address and password to send register mail from, format: email[:password]")
	flag.BoolVar(&emailtls, "emailtls", false, "enable TLS when send email.")
	flag.StringVar(&emailSuffix, "domain", "@example.com", "Valid email suffix")
	flag.BoolVar(&isDebug, "debug", false, "Debug mode switch")
	flag.StringVar(&prikeyfile, "private", "certs/server.key", "private key file for jwt")
	flag.StringVar(&pubkeyfile, "public", "certs/server.pem", "public key file for jwt")
	flag.StringVar(&sentryDSN, "sentry", "http://7:6@sentry.example.com/3", "sentry Data Source Name")
	flag.BoolVar(&queryUser, "queryuser", true, "when authenticating the end user, whether to check the user exists")

	var adminPasswd, adminEmail string
	flag.StringVar(&adminPasswd, "adminpassword", "admin", "initial password of admin")
	flag.StringVar(&adminEmail, "adminemail", "admin@example.com", "email of admin")
	flag.Parse()

	if isDebug {
		log.EnableDebug()
	}

	parts := strings.Split(emailUserPassword, ":")
	if len(parts) == 1 {
		emailFrom = parts[0]
	} else if len(parts) == 2 {
		emailFrom, emailPassword = parts[0], parts[1]
	} else {
		log.Errorf("invalid from value")
		os.Exit(-1)
	}
	userback := user.New(mysqlDSN, adminEmail, adminPasswd)

	server := ssolib.NewServer(
		mysqlDSN, siteURL, smtpAddr, emailFrom, emailPassword, emailSuffix,
		emailtls, isDebug, prikeyfile, pubkeyfile, sentryDSN, queryUser)

	server.SetUserBackend(userback)

	log.Error(server.ListenAndServe(webAddr, ssomysql.AddHandlers))
}
