package ssomysql

import (
	"github.com/laincloud/sso/ssolib"
)

func AddHandlers(s *ssolib.Server) {
	s.Post("/api/users", "UsersRegistration", UsersPost)
	s.Get("/api/usernameofemail", "UserOfEmail", UserOfEmail)
	s.AddRestfulResource("/api/activateuser", "ActivateUserResource", ActivateUserResource{})
	s.AddRestfulResource("/api/inactiveusers", "InactiveUsersResource", InactiveUsersResource{})
	s.AddRestfulResource("/api/request-reset-password-by-email", "RequestResetPasswordResourceByEmail", RequestResetPasswordResourceByEmail{})
	s.AddRestfulResource("/api/users/:username/request-reset-password", "RequestResetPasswordResource", RequestResetPasswordResource{})
	s.AddRestfulResource("/api/users/:username/reset-password", "ResetPasswordResource", ResetPasswordResource{})
}
