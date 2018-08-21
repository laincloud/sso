package ssoldap

import (
	"github.com/laincloud/sso/ssolib"
)

func AddHandlers(s *ssolib.Server) {
	s.Post("/api/users", "UsersPost", UsersPost)
}
