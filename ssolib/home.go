package ssolib

import (
	"net/http"

	"golang.org/x/net/context"
)

func (s *Server) Home(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	http.Redirect(w, r, "/spa/", http.StatusMovedPermanently)
	return ctx
}

func (s *Server) PageApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	s.renderHtmlOr500(w, http.StatusOK, "spa", nil)
	return ctx
}
