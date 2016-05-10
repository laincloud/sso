package form

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

type A struct {
	Name string
	Tag  string
}

func TestValue(t *testing.T) {
	r := basicRequest()
	r.Form.Add("gender", "male")
	r.Form.Add("int", "98")
	r.Form.Add("bool", "true")

	a := A{"Hello", "World"}
	data, _ := json.Marshal(a)
	r.Form.Add("json", string(data))

	if gender := ParamStringOptions(r, "gender", []string{"female", "male"}, "female"); gender != "male" {
		t.Errorf("Should return 'male' for gender")
	}
	if v := ParamInt(r, "int", -1); v != 98 {
		t.Errorf("ParamInt failed")
	}
	if v := ParamInt64(r, "int", -1); v != 98 {
		t.Errorf("ParamInt64 failed")
	}
	if v := ParamFloat64(r, "int", -1); v != 98 {
		t.Errorf("ParamFloat64 failed")
	}
	if v := ParamFloat32(r, "int", -1); v != 98 {
		t.Errorf("ParamFloat32 failed")
	}
	if v := ParamBoolean(r, "bool", false); !v {
		t.Errorf("ParamBoolean failed")
	}

	var b A
	if err := ParamJson(r, "json", &b); err != nil || b.Name != "Hello" {
		t.Errorf("ParamJson failed, %s", err)
	}
	fmt.Println(b)

	var m map[string]interface{}
	if err := ParamJson(r, "json", &m); err != nil {
		t.Errorf("Should be able to unmarshal to a map, %s", err)
	}
	fmt.Println(m)
}

func TestValidation(t *testing.T) {
	r := basicRequest()
	r.Form.Add("hello", "world")
	r.Form.Add("mobile", "13811110000")
	r.Form.Add("mobile2", "1x811110000")
	r.Form.Add("email", "winters.mI@gmail.com")
	r.Form.Add("email2", "winters.mIgmail.com")
	r.Form.Add("int", "98")

	if ok := ValidateString(r, "hello"); !ok {
		t.Errorf("Should containing the hello")
	}
	if ok := ValidateString(r, "hello2"); ok {
		t.Errorf("Should not containing the hello2 key")
	}
	if ok := ValidateEmail(r, "email"); !ok {
		t.Errorf("Should be an email address")
	}
	if ok := ValidateEmail(r, "email2"); ok {
		t.Errorf("Should not be an email address")
	}
	if ok := ValidateMobile(r, "mobile"); !ok {
		t.Errorf("Should be a mobile phone number")
	}
	if ok := ValidateMobile(r, "mobile2"); ok {
		t.Errorf("Should not be a mobile phone number")
	}
	if ok := ValidateInt(r, "int"); !ok {
		t.Errorf("Should be a int number")
	}
	if ok := ValidateInt(r, "int", 10); !ok {
		t.Errorf("Should be a int number more than 10")
	}
	if ok := ValidateInt(r, "int", 10, 90); ok {
		t.Errorf("Should be validate as a number less than 90")
	}
}

func basicRequest() *http.Request {
	return &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{
			Host:   "example.com",
			Scheme: "http",
		},
		Form: url.Values{},
	}
}
