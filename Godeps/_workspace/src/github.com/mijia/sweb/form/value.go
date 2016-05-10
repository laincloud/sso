package form

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// ParamString returns string value from the form or query
func ParamString(r *http.Request, name string, defaultValue string) string {
	value := r.FormValue(name)
	if value == "" {
		value = defaultValue
	}
	return value
}

// ParamStringOptions will check if the form value is contained in options, otherwise will return the default value.
func ParamStringOptions(r *http.Request, name string, options []string, defaultValue string) string {
	value := r.FormValue(name)
	isOption := false
	for _, option := range options {
		if option == value {
			isOption = true
			break
		}
	}
	if isOption {
		return value
	}
	return defaultValue
}

// ParamDefined will check if has such param defined in url queries.
func ParamDefined(r *http.Request, name string) bool {
	_, defined := r.URL.Query()[name]
	return defined
}

// ParamInt returns int value from the form or query
func ParamInt(r *http.Request, name string, defaultValue int) int {
	value, err := strconv.ParseInt(r.FormValue(name), 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(value)
}

// ParamInt64 returns int64 value from the form or query
func ParamInt64(r *http.Request, name string, defaultValue int64) int64 {
	value, err := strconv.ParseInt(r.FormValue(name), 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// ParamFloat64 returns float64 value from the form or query
func ParamFloat64(r *http.Request, name string, defaultValue float64) float64 {
	value, err := strconv.ParseFloat(r.FormValue(name), 10)
	if err != nil {
		return defaultValue
	}
	return value
}

// ParamFloat32 returns float32 value from the form or query
func ParamFloat32(r *http.Request, name string, defaultValue float32) float32 {
	value, err := strconv.ParseFloat(r.FormValue(name), 10)
	if err != nil {
		return defaultValue
	}
	return float32(value)
}

// ParamBoolean returns boolean value from the form or query
func ParamBoolean(r *http.Request, name string, defaultValue bool) bool {
	value, err := strconv.ParseBool(r.FormValue(name))
	if err != nil {
		return defaultValue
	}
	return value
}

// ParamJson returns a json unmarshal result from the form or query
func ParamJson(r *http.Request, name string, v interface{}) error {
	value := r.FormValue(name)
	return json.Unmarshal([]byte(value), v)
}

// ParamBodyJson returns a json unmarshal result from the request body
func ParamBodyJson(r *http.Request, v interface{}, closeBody ...bool) error {
	defer func() {
		if len(closeBody) == 0 || closeBody[0] {
			r.Body.Close()
		}
	}()
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}
