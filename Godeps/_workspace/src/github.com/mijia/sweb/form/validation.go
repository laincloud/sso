package form

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
)

var (
	validMobile = regexp.MustCompile(`^[\d]{11}$`)
	validEmail  = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
)

// ValidateString will validate a string field for the form, if the re pattern is passing in, validation would be check
// on the regexp.MatchString, else only check for emptyness.
func ValidateString(r *http.Request, name string, re ...*regexp.Regexp) bool {
	value := r.FormValue(name)
	if len(re) == 0 {
		return value != ""
	}

	return re[0].MatchString(value)
}

// ValidateEmail will validate a form field and check if it is an Email address.
func ValidateEmail(r *http.Request, name string) bool {
	return ValidateString(r, name, validEmail)
}

// ValidateMobile will validate a form field and check if it is a mobile phone number (for China format, e.g. "13911119999")
func ValidateMobile(r *http.Request, name string) bool {
	return ValidateString(r, name, validMobile)
}

// ValidateInt will avlidate a form field and check if it is a int number or you can pass in the min/max value for this
func ValidateInt(r *http.Request, name string, minmax ...int64) bool {
	value := r.FormValue(name)
	if value == "" {
		return false
	}

	if intValue, err := strconv.ParseInt(value, 10, 64); err != nil {
		return false
	} else {
		if len(minmax) > 0 {
			if intValue < minmax[0] {
				return false
			}
		}
		if len(minmax) > 1 {
			if intValue > minmax[1] {
				return false
			}
		}
	}
	return true
}

// ValidateFileExts will validate a form field and check if it's file extension in the allowed list.
func ValidateFileExts(r *http.Request, name string, exts ...string) bool {
	if len(exts) == 0 {
		return true
	}

	_, header, err := r.FormFile(name)
	if err != nil {
		return false
	}
	ext, allowed := filepath.Ext(header.Filename), false
	for _, allowExt := range exts {
		if ext == allowExt {
			allowed = true
			break
		}
	}
	return allowed
}
