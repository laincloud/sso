package ssomysql

import (
	"github.com/laincloud/sso/ssolib"
)

func AddHandlers(s *ssolib.Server) {
	s.Post("/api/users", "UsersRegistration", UsersPost)
	s.AddRestfulResource("/api/activateuser", "ActivateUserResource", ActivateUserResource{})
	s.AddRestfulResource("/api/request-reset-password-by-email", "RequestResetPasswordResourceByEmail", RequestResetPasswordResourceByEmail{})
	s.AddRestfulResource("/api/users/:username/request-reset-password", "RequestResetPasswordResource", RequestResetPasswordResource{})
	s.AddRestfulResource("/api/users/:username/reset-password", "ResetPasswordResource", ResetPasswordResource{})
}
