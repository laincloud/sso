package server

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"
)

func DummyHandle(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return ctx
}

type RCase struct {
	name     string
	params   []interface{}
	expected string
}

func TestRouterFuncs(t *testing.T) {
	ctx := context.Background()
	srv := New(ctx, true)
	srv.EnableAssetsPrefix("http://example.com")
	srv.Get("/", "Index", DummyHandle)
	srv.Get("/p1/:name", "Param1", DummyHandle)
	srv.Get("/p2/:name/:and", "Param2", DummyHandle)
	srv.Get("/p2x/:name/edit/:and", "Param2x", DummyHandle)
	srv.Get("/star/*whatever", "Star", DummyHandle)
	srv.Files("/assets/*filepath", http.Dir("public"))

	cases := []RCase{
		RCase{"Index", []interface{}{}, "/"},
		RCase{"Index", []interface{}{1}, "/"},
		RCase{"Param1", []interface{}{"mijia"}, "/p1/mijia"},
		RCase{"Param2", []interface{}{"mijia", "hello"}, "/p2/mijia/hello"},
		RCase{"Param2x", []interface{}{"mijia", "hello"}, "/p2x/mijia/edit/hello"},
		RCase{"Param1", []interface{}{}, "/p1/:name"},
		RCase{"Param1", []interface{}{"mijia", "and"}, "/p1/mijia"},
		RCase{"Star", []interface{}{"/yeah/you/got/it"}, "/star/yeah/you/got/it"},
	}

	for _, rCase := range cases {
		if urlPath := srv.Reverse(rCase.name, rCase.params...); urlPath != rCase.expected {
			t.Errorf("Wrong result from reverse url path name=%s, expected=%s, got=%s",
				rCase.name, rCase.expected, urlPath)
		}
	}

	if urlPath := srv.Assets("images/test.png"); urlPath != "http://example.com/assets/images/test.png" {
		t.Errorf("Wrong result from assets, expected=/assets/images/test.png, got=%s", urlPath)
	}
}
