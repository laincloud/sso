package ssolib

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/context"
)

var (
	slugPattern  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*$`)
	emailPattern = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4}$`)
)

func ValidateSlug(slug string) error {
	if slug == "" {
		return errors.New("Empty slug")
	}

	if len(slug) > 32 {
		return errors.New("Slug too long")
	}

	if !slugPattern.MatchString(slug) {
		return errors.New("Invalid slug")
	}

	return nil
}

func ValidateFullName(fullname string) error {
	if fullname == "" {
		return errors.New("Empty fullname")
	}

	if len(fullname) > 128 {
		return errors.New("Fullname too long")
	}

	return nil
}

func ValidateURI(uri string) error {
	if !(strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")) {
		return errors.New("URL must has http:// or https:// prefix")
	}
	return nil
}

func ValidateUserEmail(email string, ctx context.Context) error {
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("Invalid email: %s", email)
	}
	if getEmailSuffix(ctx) != "" && getEmailSuffix(ctx) != "@example.com" {
		if !strings.HasSuffix(email, getEmailSuffix(ctx)) {
			return errors.New("Unsupported email")
		}
	}
	return nil
}
