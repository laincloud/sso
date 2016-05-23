package main

import (
	"flag"
	"runtime"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib"
	"github.com/laincloud/sso/ssomysql"
	"github.com/laincloud/sso/ssomysql/user"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var webAddr, mysqlDSN, siteURL, smtpAddr, emailFrom, emailSuffix string
	var prikeyfile, pubkeyfile string
	var isDebug bool
	var sentryDSN string
	flag.StringVar(&webAddr, "web", ":14000", "The address which SSO service is listening on")
	flag.StringVar(&mysqlDSN, "mysql", "user:password@tcp(127.0.0.1:3306)/dbname",
		"Data source name of mysql connection")
	flag.StringVar(&siteURL, "site", "http://sso.example.com", "Base URL of SSO site")
	flag.StringVar(&smtpAddr, "smtp", "mail.example.com:25", "SMTP address for sending mail")
	flag.StringVar(&emailFrom, "from", "sso@example.com", "Email address to send register mail from")
	flag.StringVar(&emailSuffix, "domain", "@example.com", "Valid email suffix")
	flag.BoolVar(&isDebug, "debug", false, "Debug mode switch")
	flag.StringVar(&prikeyfile, "private", "certs/server.key", "private key file for jwt")
	flag.StringVar(&pubkeyfile, "public", "certs/server.pem", "public key file for jwt")
	flag.StringVar(&sentryDSN, "sentry", "http://7:6@sentry.example.com/3", "sentry Data Source Name")

	var adminPasswd, adminEmail string
	flag.StringVar(&adminPasswd, "adminpassword", "admin", "initial password of admin")
	flag.StringVar(&adminEmail, "adminemail", "admin@example.com", "email of admin")
	flag.Parse()

	if isDebug {
		log.EnableDebug()
	}

	userback := user.New(mysqlDSN, adminEmail, adminPasswd)

	server := ssolib.NewServer(mysqlDSN, siteURL, smtpAddr, emailFrom, emailSuffix, isDebug, prikeyfile, pubkeyfile, sentryDSN)

	server.SetUserBackend(userback)

	log.Error(server.ListenAndServe(webAddr, ssomysql.AddHandlers))
}
