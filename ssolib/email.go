package ssolib

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
)

type Mail struct {
	From        string
	To          string
	Subject     string
	HTML        string
}


type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		}
	}
	return nil, nil
}


// Send does what it is supposed to do.
func (m *Mail) Send(host string, port int, user, pass string) error {
	// validate from address
	from, err := mail.ParseAddress(m.From)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// validate to address
	tos, err := mail.ParseAddressList(m.To)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var addresses string
	for _, to := range tos {
		addresses = addresses + "," + to.Address
	}

	// set headers for html email
	header := textproto.MIMEHeader{}
	header.Set(textproto.CanonicalMIMEHeaderKey("from"), from.Address)
	header.Set(textproto.CanonicalMIMEHeaderKey("to"), addresses[1:])
	header.Set(textproto.CanonicalMIMEHeaderKey("content-type"), "text/html; charset=UTF-8")
	header.Set(textproto.CanonicalMIMEHeaderKey("mime-version"), "1.0")
	header.Set(textproto.CanonicalMIMEHeaderKey("subject"), fmt.Sprintf("=?utf-8?B?%s?=", b64.StdEncoding.EncodeToString([]byte(m.Subject))))

	// init empty message
	var buffer bytes.Buffer

	// write header
	for key, value := range header {
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value[0]))
	}

	// write body
	buffer.WriteString(fmt.Sprintf("\r\n%s", m.HTML))

	// send email
	addr := fmt.Sprintf("%s:%d", host, port)
	auth := LoginAuth(user, pass)
	mails := strings.Split(m.To, ",")

	if mails[len(mails)-1] == "" {
		mails = mails[:len(mails)-1]
	}

	return smtp.SendMail(addr, auth, from.Address, mails, buffer.Bytes())
}


func SendTo(subject, content, to string) error {
	m := Mail{
		From:    "noreply@bdp.yixin.com",
		To:      to,
		Subject: subject,
	}
	m.HTML = content

	err := m.Send("mail.bdp.cc", 25, "username", "password")
	return err
}
