package server

import "usermanage/gen/proto/conf"

var maskedOperations = []string{"/auth.v1.AuthService/Login"}

func generateMaskedOperations(c *conf.Server) []string {
	if c.Debug {
		return []string{}
	}
	return []string{
		"/auth.v1.AuthService/Login",
		"/auth.v1.AuthService/RegisterRequest",
		"/auth.v1.AuthService/ChangePasswordRequest",
	}
}
